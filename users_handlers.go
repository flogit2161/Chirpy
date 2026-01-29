package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/flogit2161/Chirpy/internal/auth"
	"github.com/flogit2161/Chirpy/internal/database"
)

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

func (cfg *apiConfig) handlerLogIn(w http.ResponseWriter, r *http.Request) {
	type loginParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	token, err := auth.MakeJWT(userLogs.ID, cfg.jwt, 1*time.Hour)
	if err != nil {
		respondWithError(w, 500, "Unable to create token for user")
		return
	}

	encodedRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "Unable to create encoded string refresh token")
		return
	}

	_, err = cfg.db.GenerateRefreshToken(
		r.Context(),
		database.GenerateRefreshTokenParams{
			Token:  encodedRefreshToken,
			UserID: userLogs.ID,
		},
	)
	if err != nil {
		respondWithError(w, 500, "Unable to create refresh token for user")
		return
	}

	jsonUser := User{
		ID:           userLogs.ID,
		CreatedAt:    userLogs.CreatedAt,
		UpdatedAt:    userLogs.UpdatedAt,
		Email:        userLogs.Email,
		Token:        token,
		RefreshToken: encodedRefreshToken,
	}
	respondWithJSON(w, 200, jsonUser)

}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Refresh token is either expired or does not exist")
		return
	}

	userByToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, 401, "Could not acces user via refresh token, token does not exist")
		return
	}

	newJWTToken, err := auth.MakeJWT(userByToken.ID, cfg.jwt, 1*time.Hour)
	if err != nil {
		respondWithError(w, 500, "Could not re-create JWT Token for user")
		return
	}

	type tokenResponse struct {
		Token string `json:"token"`
	}

	tokenResp := tokenResponse{
		Token: newJWTToken,
	}

	respondWithJSON(w, 200, tokenResp)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Refresh token is either expired or does not exist")
		return
	}

	err = cfg.db.RevokeToken(r.Context(), bearerToken)
	if err != nil {
		respondWithError(w, 401, "Unable to revoke refresh token")
		return
	}

	w.WriteHeader(204)
}
