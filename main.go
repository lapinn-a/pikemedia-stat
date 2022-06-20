package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v4"
	"io/ioutil"
	"time"

	//"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type BrowserClientInfo struct {
	UserIP               string `json:"userIP"`
	Platform             string `json:"platform"`
	BrowserClient        string `json:"browserClient"`
	ScreenDataViewPort   string `json:"screenData_viewPort"`
	ScreenDataResolution string `json:"screenData_resolution"`
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
	UserIP                   string        `json:"userIP"`
	Platform                 string        `json:"platform"`
	BrowserClient            string        `json:"browserClient"`
	ScreenDataViewPort       string        `json:"screenData_viewPort"`
	ScreenDataResolution     string        `json:"screenData_resolution"`
}

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /ping request\n")
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]string{"status": "up"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ping failed: %v\n", err)
		os.Exit(1)
	}
}

func stat(conn *pgx.Conn, startTime time.Time) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var Count int
		err := conn.QueryRow(context.Background(), `select count(*) from "stats"`).Scan(&Count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stat failed: %v\n", err)
			os.Exit(1)
		}
		err = json.NewEncoder(w).Encode(map[string]any{"count": Count, "uptime": time.Since(startTime).Seconds()})
		if err != nil {
			fmt.Fprintf(os.Stderr, "stat failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func collect(conn *pgx.Conn) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sqlStr := `INSERT INTO stats("viewerId","name","lastName","isChatName","email","isChatEmail","joinTime","leaveTime","spentTime","spentTimeDeltaPercent","chatCommentsTotal","chatCommentsDeltaPercent","anotherFields","userIP","platform","browserClient","screenData_viewPort","screenData_resolution") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`
		fmt.Printf("got /collect request\n")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "collect failed: %v\n", err)
			os.Exit(1)
		}
		targets := []Viewer{}

		err = json.Unmarshal(body, &targets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "collect failed: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(map[string]string{"result": "failed"})
			if err != nil {
				fmt.Fprintf(os.Stderr, "collect failed: %v\n", err)
				os.Exit(1)
			}
			return
		}

		for _, t := range targets {
			fmt.Println(t.ViewerId, "-", t.Name)
			_, err = conn.Exec(context.Background(), sqlStr, t.ViewerId, t.Name, t.LastName, t.IsChatName, t.Email, t.IsChatEmail, t.JoinTime, t.LeaveTime, t.SpentTime, t.SpentTimeDeltaPercent, t.ChatCommentsTotal, t.ChatCommentsDeltaPercent, t.AnotherFields, t.BrowserClientInfo.UserIP, t.BrowserClientInfo.Platform, t.BrowserClientInfo.BrowserClient, t.BrowserClientInfo.ScreenDataViewPort, t.BrowserClientInfo.ScreenDataResolution)
			if err != nil {
				fmt.Fprintf(os.Stderr, "collect failed: %v\n", err)
				os.Exit(1)
			}
		}

		err = json.NewEncoder(w).Encode(map[string]string{"result": "success"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "collect failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func report(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /report request\n")
	io.WriteString(w, "TODO\n")
}

func main() {
	startTime := time.Now()
	urlExample := "postgres://postgres:123@localhost:5432/postgres"
	conn, err := pgx.Connect(context.Background(), urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", ping)
	mux.HandleFunc("/stat", stat(conn, startTime))
	mux.HandleFunc("/collect", collect(conn))
	mux.HandleFunc("/report", report)

	err = http.ListenAndServe(":80", mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
