package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/flogit2161/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	count := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, count)))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Reset number of hits to server"))
}

func (cfg *apiConfig) handlerValidate(w http.ResponseWriter, r *http.Request) {
	type bodyJSON struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()
	body := bodyJSON{}

	err := decoder.Decode(&body)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(body.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	//MOVE THIS TO HELPER FUNC
	split := strings.Fields(body.Body)
	for i, s := range split {
		lower := strings.ToLower(s)
		if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
			split[i] = "****"
		}
	}
	cleanedWords := strings.Join(split, " ")

	type validResponse struct {
		Cleaned string `json:"cleaned_body"`
	}
	respondWithJSON(w, 200, validResponse{
		Cleaned: cleanedWords,
	})

}
