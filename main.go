package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"log"
	"net/http"
	"time"
)

func main() {
	startTime := time.Now()
	urlExample := "postgres://postgres:123@localhost:5432/postgres8"
	conn, err := pgx.Connect(context.Background(), urlExample)
	if err != nil {
		log.Fatalf("FATAL: Unable to connect to database: %v\n", err)
	}
	defer func(conn *pgx.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {

		}
	}(conn, context.Background())

	stat := NewStat(conn, startTime)
	err = http.ListenAndServe(":81", stat.Router())

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("Server closed\n")
	} else if err != nil {
		log.Fatalf("FATAL: Error starting server: %s\n", err)
	}
}
