package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
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

	err = stat.Router().Run(":81")
	if err != nil {
		log.Fatalf("FATAL: Error starting server: %s\n", err)
	}
}
