package main

import (
	"net/http"

	"github.com/Paa-Kwasi-04/chirpy/cmd"
	_ "github.com/lib/pq"
)

func main() {
	cfg := cmd.Startup()

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./fileserver"))
	mux.Handle("/app/", http.StripPrefix("/app", cfg.MiddlewareMetricsInc(fileServer)))

	mux.HandleFunc("GET /api/healthz", cmd.HandleHealth)

	mux.HandleFunc("POST /api/users", cfg.HandleCreateUser)
	mux.HandleFunc("POST /api/login",cfg.HandleLogin)
	
	mux.HandleFunc("POST /api/chirps", cfg.HandleCreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.HandleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.HandleGetChirp)

	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
