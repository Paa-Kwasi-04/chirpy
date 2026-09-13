package main

import (
	"net/http"
	"github.com/Paa-Kwasi-04/chirpy/cmd"
)

func main() {

	var cfg cmd.ApiConfig

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./fileserver"))
	mux.Handle("/app/", http.StripPrefix("/app", cfg.MiddlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /api/healthz", cmd.HandleHealth)
	mux.HandleFunc("POST /api/validate_chirp",cmd.HandleValidate)
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)
	

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}


