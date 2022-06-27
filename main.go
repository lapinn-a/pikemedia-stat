package main

import (
	"database/sql"
	"errors"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

func main() {

	db, err := sql.Open("sqlite3", "./db/stats.db")
	if err != nil {
		log.Fatalf("FATAL: Error opening database: %s\n", err)
	}

	startTime := time.Now()
	stat := NewStat(db, startTime)

	err = stat.RunMigrations()
	if err != nil {
		log.Fatalf("FATAL: Error running migrations: %s\n", err)
	}

	err = http.ListenAndServe(":81", stat.Router())

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("Server closed\n")
	} else if err != nil {
		log.Fatalf("FATAL: Error starting server: %s\n", err)
	}
}
