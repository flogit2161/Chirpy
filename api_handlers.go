package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/flogit2161/Chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}
	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Could not reset users table")
	}
	w.WriteHeader(200)
	w.Write([]byte("Users table reseted successfully"))
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type BodyJSON struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()
	body := BodyJSON{}

	err := decoder.Decode(&body)
	if err != nil {
		respondWithError(w, 500, "Error decoding the request")
		return
	}

	if len(body.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	//Filtering body.Body bad words
	split := strings.Fields(body.Body)
	for i, s := range split {
		lower := strings.ToLower(s)
		if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
			split[i] = "****"
		}
	}
	cleanedBody := strings.Join(split, " ")

	parsedUserID, err := uuid.Parse(body.UserID)
	if err != nil {
		respondWithError(w, 400, "Error parsing User's ID into a UUID")
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: parsedUserID,
	})

	jsonChirp := Chirps{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 201, jsonChirp)

}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()
	user := User{}

	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, 500, "Error decoding the request")
		return
	}

	createdUser, err := cfg.db.CreateUser(r.Context(), user.Email)
	if err != nil {
		respondWithError(w, 500, "Error creating user")
		return
	}
	// Map database.User to main.User to send back JSON
	user = User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	respondWithJSON(w, 201, user)
}
