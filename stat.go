package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/ipinfo/go/v2/ipinfo"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Stat struct {
	conn      *sql.DB
	startTime time.Time
}

func NewStat(conn *sql.DB, startTime time.Time) *Stat {
	return &Stat{conn: conn, startTime: startTime}
}

type Resolution struct {
	X int
	Y int
}

type Platform struct {
	Name         string
	Version      string
	Architecture string
}

type BrowserClient struct {
	Name    string
	Version string
}

type BrowserClientInfo struct {
	ScreenDataViewPort   Resolution    `json:"screenData_viewPort"`
	ScreenDataResolution Resolution    `json:"screenData_resolution"`
	Platform             Platform      `json:"platform"`
	BrowserClient        BrowserClient `json:"browserClient"`
	UserIP               string        `json:"userIP"`
	UserRegion           string
	UserProvider         string
}

type Viewer struct {
	BrowserClientInfo        `json:"browserClientInfo"`
	ViewerId                 int32         `json:"viewerId"`
	Name                     string        `json:"name"`
	LastName                 string        `json:"lastName"`
	IsChatName               bool          `json:"isChatName"`
	Email                    string        `json:"email"`
	IsChatEmail              bool          `json:"isChatEmail"`
	JoinTime                 string        `json:"joinTime"`
	LeaveTime                string        `json:"leaveTime"`
	SpentTime                int64         `json:"spentTime"`
	SpentTimeDeltaPercent    uint8         `json:"spentTimeDeltaPercent"`
	ChatCommentsTotal        int32         `json:"chatCommentsTotal"`
	ChatCommentsDeltaPercent uint8         `json:"chatCommentsDeltaPercent"`
	AnotherFields            []interface{} `json:"anotherFields"`
}

func (resolution *Resolution) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	split := strings.Split(s, "x")
	resolution.X, err = strconv.Atoi(split[0])
	if err != nil {
		return err
	}
	resolution.Y, err = strconv.Atoi(split[1])
	if err != nil {
		return err
	}
	return nil
}

func (platform *Platform) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	split := strings.Split(s, " ")
	platform.Architecture = split[len(split)-1]
	platform.Version = split[len(split)-2]
	platform.Name = s[0 : len(s)-len(platform.Architecture)-len(platform.Version)-2]
	return nil
}

func (browserClient *BrowserClient) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	split := strings.Split(s, " ")
	browserClient.Version = split[len(split)-1]
	browserClient.Name = s[0 : len(s)-len(browserClient.Version)-1]
	return nil
}

func (stat *Stat) Ping(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"status": "up"})
}

func (stat *Stat) Stats(c *gin.Context) {
	var Count int
	err := stat.conn.QueryRow(`select count(*) from "stats"`).Scan(&Count)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal server error")
		log.Printf("Stats failed: %v\n", err)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"count": Count, "uptime": time.Since(stat.startTime).Seconds()})
}

func (stat *Stat) Collect(c *gin.Context) {
	sqlStr := `INSERT INTO stats("viewerId","name","lastName","isChatName","email","isChatEmail","joinTime","leaveTime","spentTime","spentTimeDeltaPercent","chatCommentsTotal","chatCommentsDeltaPercent","anotherFields","userIP","userRegion","userProvider","platformName","platformVersion","platformArchitecture","browserClientName","browserClientVersion","screenData_viewPortX","screenData_viewPortY","screenData_resolutionX","screenData_resolutionY") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`
	//body, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	log.Printf("Collect failed: %v\n", err)
	//	c.JSON(http.StatusBadRequest, gin.H{"result": "failed"})
	//	return
	//}
	var targets []Viewer

	//err = json.Unmarshal(body, &targets)

	err := c.BindJSON(&targets)
	if err != nil {
		log.Printf("Collect failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"result": "failed"})
		return
	}

	for _, t := range targets {
		client := ipinfo.NewClient(nil, nil, "887d18d82ff5e2")
		info, err := client.GetIPInfo(net.ParseIP(t.UserIP))
		if err != nil {
			log.Printf("GetIPInfo failed: %v\n", err)
		} else {
			t.UserRegion = info.Region
			t.UserProvider = info.Org
		}
		anotherFields, _ := json.Marshal(t.AnotherFields)
		_, err = stat.conn.Exec(sqlStr, t.ViewerId, t.Name, t.LastName, t.IsChatName, t.Email, t.IsChatEmail, t.JoinTime, t.LeaveTime, t.SpentTime, t.SpentTimeDeltaPercent, t.ChatCommentsTotal, t.ChatCommentsDeltaPercent, anotherFields, t.UserIP, t.UserRegion, t.UserProvider, t.Platform.Name, t.Platform.Version, t.Platform.Architecture, t.BrowserClient.Name, t.BrowserClient.Version, t.ScreenDataViewPort.X, t.ScreenDataViewPort.Y, t.ScreenDataResolution.X, t.ScreenDataResolution.Y)
		if err != nil {
			log.Printf("Collect failed: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"result": "failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"result": "success"})
}

func countPeaks(rows *sql.Rows) (peakStartTime time.Time, peakEndTime time.Time, peakCount int) {
	var currentCount int

	for rows.Next() {
		var timeValueString string
		var timeValue time.Time
		var change int
		err := rows.Scan(&timeValueString, &change)
		if err != nil {
			log.Printf("countPeaks failed: %v\n", err)
			return
		}
		timeValue, err = time.Parse(time.RFC3339, timeValueString)
		if err != nil {
			log.Printf("countPeaks failed: %v\n", err)
			return
		}
		currentCount += change
		if currentCount > peakCount {
			peakCount = currentCount
			peakStartTime = timeValue
			peakEndTime = peakStartTime
		} else if peakEndTime == peakStartTime {
			peakEndTime = timeValue
		}
	}
	return peakStartTime, peakEndTime, peakCount
}

func (stat *Stat) Report(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/csv")
	var sqlStr string
	var rows *sql.Rows
	var err error

	if c.Query("platformName") != "" {
		sqlStr = `SELECT "platformVersion", count(*) FROM "stats" WHERE "platformName" = $1 GROUP BY "platformVersion"`
		rows, err = stat.conn.Query(sqlStr, c.Query("platformName"))
	} else if c.Query("browserClientName") != "" {
		sqlStr = `SELECT "browserClientVersion", count(*) FROM "stats" WHERE "browserClientName" = $1 GROUP BY "browserClientVersion"`
		rows, err = stat.conn.Query(sqlStr, c.Query("browserClientName"))
	} else if c.Query("column") != "" {
		switch c.Query("column") {
		case "platformName":
			sqlStr = `SELECT "platformName", count(*) FROM "stats" GROUP BY "platformName"`
		case "browserClientName":
			sqlStr = `SELECT "browserClientName", count(*) FROM "stats" GROUP BY "browserClientName"`
		case "browserClient":
			sqlStr = `SELECT "browserClientName" || ' ' || "browserClientVersion" AS browserClient, count(*) FROM "stats" GROUP BY browserClient`
		case "screenData_resolution":
			sqlStr = `SELECT "screenData_resolutionX" || 'x' || "screenData_resolutionY" as screenData_resolution, count(*) FROM "stats" GROUP BY screenData_resolution`
		case "userRegion":
			sqlStr = `SELECT "userRegion", count(*) FROM "stats" GROUP BY "userRegion"`
		case "userProvider":
			sqlStr = `SELECT "userProvider", count(*) FROM "stats" GROUP BY "userProvider"`
		case "viewsPeaks":
			sqlStr = `SELECT "joinTime", 1 FROM stats UNION ALL SELECT "leaveTime", -1 FROM stats ORDER BY "joinTime"`
			rows, err = stat.conn.Query(sqlStr)
			if err != nil {
				c.String(http.StatusInternalServerError, "Internal server error")
				log.Printf("Report failed: %v\n", err)
				return
			}
			peakStartTime, peakEndTime, peakCount := countPeaks(rows)
			c.String(http.StatusOK, "startTime,endTime,count\n%v,%v,%v", peakStartTime, peakEndTime, peakCount)
			return
		default:
			c.String(http.StatusBadRequest, "failed")
			return
		}
		rows, err = stat.conn.Query(sqlStr)
	} else {
		c.String(http.StatusBadRequest, "failed")
		return
	}

	if err != nil {
		c.String(http.StatusInternalServerError, "failed")
		log.Printf("Report failed: %v\n", err)
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Close failed: %v\n", err)
		}
	}(rows)

	var name string
	var cnt int

	c.String(http.StatusOK, "%s,count", c.Query("column"))
	for rows.Next() {
		err := rows.Scan(&name, &cnt)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed")
			log.Printf("Report failed: %v\n", err)
			return
		}
		c.String(http.StatusOK, "\n%v,%v", name, cnt)
		if err != nil {
			log.Fatalf("FATAL: Report failed: %v\n", err)
		}
	}
}
