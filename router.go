package main

import (
	"log"
	"net/http"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v %v", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (stat *Stat) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", stat.Ping)
	mux.HandleFunc("/stat", stat.Stats)
	mux.HandleFunc("/collect", stat.Collect)
	mux.HandleFunc("/report", stat.Report)
	return Logging(mux)
}
