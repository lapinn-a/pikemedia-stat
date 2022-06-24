package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

func main() {

	db, err := sql.Open("sqlite3", "./stats.db")
	if err != nil {
		panic(err)
	}
	//defer db.Close()
	//result, err := db.Exec(`insert into products(model, company, price) values ('iPhone X', $1, $2)`,
	//	"Apple", 72000)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(result.LastInsertId()) // id последнего добавленного объекта
	//fmt.Println(result.RowsAffected()) // количество добавленных строк

	startTime := time.Now()
	//urlExample := "postgres://postgres:123@localhost:5432/postgres"
	//conn, err := pgx.Connect(context.Background(), urlExample)
	//if err != nil {
	//	log.Fatalf("FATAL: Unable to connect to database: %v\n", err)
	//}
	//defer func(conn *pgx.Conn, ctx context.Context) {
	//	err := conn.Close(ctx)
	//	if err != nil {
	//		log.Printf("Error closing database connection\n")
	//	}
	//}(conn, context.Background())

	stat := NewStat(db, startTime)
	err = http.ListenAndServe(":81", stat.Router())

	if errors.Is(err, http.ErrServerClosed) {
		log.Printf("Server closed\n")
	} else if err != nil {
		log.Fatalf("FATAL: Error starting server: %s\n", err)
	}
}
