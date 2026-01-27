package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/flogit2161/Chirpy/internal/auth"
	"github.com/flogit2161/Chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwt            string
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

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithJSON(w, 500, "Could not load user's token")
	}

	validatedUUID, err := auth.ValidateJWT(bearerToken, cfg.jwt)
	if err != nil {
		respondWithJSON(w, 400, "Could not authenticate user, please log in again")
	}

	type BodyJSON struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()
	body := BodyJSON{}

	err = decoder.Decode(&body)
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

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: validatedUUID,
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

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, 500, "Error hashing the users password")
		return
	}

	createdUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hashedPassword,
	})
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

func (cfg *apiConfig) handlerRetrieveAllChirps(w http.ResponseWriter, r *http.Request) {

	chirps, err := cfg.db.RetrieveAllChirps(r.Context())
	if err != nil {
		respondWithJSON(w, 400, "Error retrieving the chirps")
		return
	}

	jsonChirpsList := []Chirps{}
	for _, ch := range chirps {
		jsonChirp := Chirps{
			ID:        ch.ID,
			CreatedAt: ch.CreatedAt,
			UpdatedAt: ch.UpdatedAt,
			Body:      ch.Body,
			UserID:    ch.UserID,
		}
		jsonChirpsList = append(jsonChirpsList, jsonChirp)
	}
	respondWithJSON(w, 200, jsonChirpsList)
}

func (cfg *apiConfig) handlerRetrieveChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	parsedID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, 400, "Error parsing Chirp ID into a UUID")
		return
	}

	chirp, err := cfg.db.RetrieveChirp(r.Context(), parsedID)
	if err != nil {
		respondWithError(w, 404, "Error trying to load the chirp at this ID")
		return
	}

	jsonChirp := Chirps{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 200, jsonChirp)

}

func (cfg *apiConfig) handlerLogIn(w http.ResponseWriter, r *http.Request) {
	type loginParams struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)

	defer r.Body.Close()
	logs := loginParams{}

	err := decoder.Decode(&logs)
	if err != nil {
		respondWithError(w, 500, "Error decoding the request")
		return
	}

	userLogs, err := cfg.db.GetUserByEmail(r.Context(), logs.Email)
	if err != nil {
		respondWithError(w, 401, "Error accessing user via email, please create user before logging in")
		return
	}

	match, err := auth.CheckPasswordHash(logs.Password, userLogs.HashedPassword)
	if err != nil {
		respondWithError(w, 500, "Error with password hashing")
		return
	}

	if !match {
		respondWithError(w, 401, "Password doesnt match")
		return
	}

	expiration := time.Hour
	if logs.ExpiresInSeconds > 0 && logs.ExpiresInSeconds < 3600 {
		expiration = time.Duration(logs.ExpiresInSeconds) * time.Second
	}

	token, err := auth.MakeJWT(userLogs.ID, cfg.jwt, expiration)
	if err != nil {
		respondWithError(w, 500, "Unable to create token for user")
		return
	}

	jsonUser := User{
		ID:        userLogs.ID,
		CreatedAt: userLogs.CreatedAt,
		UpdatedAt: userLogs.UpdatedAt,
		Email:     userLogs.Email,
		Token:     token,
	}
	respondWithJSON(w, 200, jsonUser)

}
