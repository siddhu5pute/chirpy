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

func (cfg *apiConfig) handlerReqCount(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerResetCount(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(200)
}

func main() {
	apiConf := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiConf.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /healthz", handlerOne)
	mux.HandleFunc("GET /metrics", apiConf.handlerReqCount)
	mux.HandleFunc("POST /reset", apiConf.handlerResetCount)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
