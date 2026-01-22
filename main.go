package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	count := cfg.fileserverHits.Load()
	w.Write([]byte(fmt.Sprintf("Hits: %v", count)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Reset number of hits to server"))
}

func main() {
	apiCfg := &apiConfig{}
	serveMux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("."))
	handler := http.StripPrefix("/app", fileServer)
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	serveMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	serveMux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	serveMux.HandleFunc("/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error L/S server")
	}
}
