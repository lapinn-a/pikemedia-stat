package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ipinfo/go/v2/ipinfo"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
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

func (stat *Stat) Ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"status": "up"})
	if err != nil {
		log.Printf("Ping failed: %v\n", err)
	}
}

func (stat *Stat) Stats(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var Count int
	err := stat.conn.QueryRow(`select count(*) from "stats"`).Scan(&Count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Stats failed: %v\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(map[string]any{"count": Count, "uptime": time.Since(stat.startTime).Seconds()})
	if err != nil {
		log.Printf("Stats failed: %v\n", err)
	}
}

func (stat *Stat) Collect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sqlStr := `INSERT INTO stats("viewerId","name","lastName","isChatName","email","isChatEmail","joinTime","leaveTime","spentTime","spentTimeDeltaPercent","chatCommentsTotal","chatCommentsDeltaPercent","anotherFields","userIP","userRegion","userProvider","platformName","platformVersion","platformArchitecture","browserClientName","browserClientVersion","screenData_viewPortX","screenData_viewPortY","screenData_resolutionX","screenData_resolutionY") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Collect failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(map[string]string{"result": "failed"})
		if err != nil {
			log.Fatalf("FATAL: Collect failed: %v\n", err)
		}
		return
	}
	var targets []Viewer

	err = json.Unmarshal(body, &targets)
	if err != nil {
		log.Printf("Collect failed: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(map[string]string{"result": "failed"})
		if err != nil {
			log.Fatalf("FATAL: Collect failed: %v\n", err)
		}
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
		_, err = stat.conn.Exec(sqlStr, t.ViewerId, t.Name, t.LastName, t.IsChatName, t.Email, t.IsChatEmail, t.JoinTime, t.LeaveTime, t.SpentTime, t.SpentTimeDeltaPercent, t.ChatCommentsTotal, t.ChatCommentsDeltaPercent, t.AnotherFields, t.UserIP, t.UserRegion, t.UserProvider, t.Platform.Name, t.Platform.Version, t.Platform.Architecture, t.BrowserClient.Name, t.BrowserClient.Version /**/, t.ScreenDataViewPort.X, t.ScreenDataViewPort.Y /**/, t.ScreenDataResolution.X, t.ScreenDataResolution.Y)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Collect failed: %v\n", err)
			return
		}
	}

	err = json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	if err != nil {
		log.Printf("Collect failed: %v\n", err)
		os.Exit(1)
	}
}

func countPeaks(rows *sql.Rows) (peakStartTime time.Time, peakEndTime time.Time, peakCount int) {
	var currentCount int

	for rows.Next() {
		var timeValue time.Time
		var change int
		err := rows.Scan(&timeValue, &change)
		if err != nil {
			log.Fatal(err)
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

func (stat *Stat) Report(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	var sqlStr string
	var rows *sql.Rows
	var err error

	if r.URL.Query().Has("platformName") {
		sqlStr = `SELECT "platformVersion", count(*) FROM "stats" WHERE "platformName" = $1 GROUP BY "platformVersion"`
		rows, err = stat.conn.Query(sqlStr, r.URL.Query().Get("platformName"))
	} else if r.URL.Query().Has("browserClientName") {
		sqlStr = `SELECT "browserClientVersion", count(*) FROM "stats" WHERE "browserClientName" = $1 GROUP BY "browserClientVersion"`
		rows, err = stat.conn.Query(sqlStr, r.URL.Query().Get("browserClientName"))
	} else if r.URL.Query().Has("column") {
		switch r.URL.Query().Get("column") {
		case "platformName":
			sqlStr = `SELECT "platformName", count(*) FROM "stats" GROUP BY "platformName"`
		case "browserClientName":
			sqlStr = `SELECT "browserClientName", count(*) FROM "stats" GROUP BY "browserClientName"`
		case "browserClient":
			sqlStr = `SELECT CONCAT("browserClientName", ' ', "browserClientVersion") AS browserClient, count(*) FROM "stats" GROUP BY browserClient`
		case "screenData_resolution":
			sqlStr = `SELECT CONCAT("screenData_resolutionX", 'x', "screenData_resolutionY") as screenData_resolution, count(*) FROM "stats" GROUP BY screenData_resolution`
		case "userRegion":
			sqlStr = `SELECT "userRegion", count(*) FROM "stats" GROUP BY "userRegion"`
		case "userProvider":
			sqlStr = `SELECT "userProvider", count(*) FROM "stats" GROUP BY "userProvider"`
		case "viewsPeaks":
			sqlStr = `SELECT "joinTime", 1 FROM stats UNION ALL SELECT "leaveTime", -1 FROM stats ORDER BY "joinTime"`
			rows, err = stat.conn.Query(sqlStr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("Report failed: %v\n", err)
				return
			}
			peakStartTime, peakEndTime, peakCount := countPeaks(rows)
			_, err = fmt.Fprintf(w, "startTime,endTime,count\n%v,%v,%v", peakStartTime, peakEndTime, peakCount)
			if err != nil {
				log.Printf("Report failed: %v\n", err)
				return
			}
			return
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, err = fmt.Fprintf(w, "failed")
			if err != nil {
				log.Fatalf("FATAL: Report failed: %v\n", err)
			}
			return
		}
		rows, err = stat.conn.Query(sqlStr)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		_, err = fmt.Fprintf(w, "failed")
		if err != nil {
			log.Fatalf("FATAL: Report failed: %v\n", err)
		}
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Report failed: %v\n", err)
		return
	}
	defer rows.Close()

	var name string
	var cnt int

	_, err = fmt.Fprintf(w, "%s,count", r.URL.Query().Get("column"))
	if err != nil {
		log.Printf("Report failed: %v\n", err)
		return
	}
	for rows.Next() {
		err := rows.Scan(&name, &cnt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("Report failed: %v\n", err)
			return
		}
		_, err = fmt.Fprintf(w, "\n%v,%v", name, cnt)
		if err != nil {
			log.Fatalf("FATAL: Report failed: %v\n", err)
		}
	}
}
