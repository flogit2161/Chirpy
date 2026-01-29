package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flogit2161/Chirpy/internal/auth"
	"github.com/flogit2161/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 500, "Could not load user's token")
		return
	}

	validatedUUID, err := auth.ValidateJWT(bearerToken, cfg.jwt)
	if err != nil {
		respondWithError(w, 401, "Could not authenticate user, please log in again")
		return
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

func (cfg *apiConfig) handlerRetrieveAllChirps(w http.ResponseWriter, r *http.Request) {

	author_id := r.URL.Query().Get("author_id")
	if author_id == "" {
		chirps, err := cfg.db.RetrieveAllChirps(r.Context())
		if err != nil {
			respondWithError(w, 400, "Error retrieving the chirps")
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
	} else {
		parsedAuthorID, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, 400, "Error parsing author ID into a UUID")
			return
		}
		usersChirps, err := cfg.db.RetrieveAllChirpsFromUser(r.Context(), parsedAuthorID)
		if err != nil {
			respondWithError(w, 400, "Error retrieving user's chirps")
			return
		}

		usersChirpsList := []Chirps{}
		for _, ch := range usersChirps {
			jsonChirp := Chirps{
				ID:        ch.ID,
				CreatedAt: ch.CreatedAt,
				UpdatedAt: ch.UpdatedAt,
				Body:      ch.Body,
				UserID:    ch.UserID,
			}
			usersChirpsList = append(usersChirpsList, jsonChirp)
		}
		respondWithJSON(w, 200, usersChirpsList)
	}

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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Token is either expired or does not exist")
		return
	}

	userUUID, err := auth.ValidateJWT(bearerToken, cfg.jwt)
	if err != nil {
		respondWithError(w, 401, "Error validating token, token is not valid anymore")
		return
	}

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

	if chirp.UserID != userUUID {
		respondWithError(w, 403, "User is not allowed to delete a chirp thats not his")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, 400, "Error deleting chirp")
		return
	}

	w.WriteHeader(204)
}
