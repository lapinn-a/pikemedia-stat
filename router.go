package main

import "net/http"

func (stat *Stat) Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", stat.Ping)
	mux.HandleFunc("/stat", stat.Stats)
	mux.HandleFunc("/collect", stat.Collect)
	mux.HandleFunc("/report", stat.Report)
	return mux
}
