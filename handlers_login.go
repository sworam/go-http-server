package main

import (
	"encoding/json"
	"main/internal/auth"
	"main/internal/database"
	"net/http"
	"time"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	match, err := auth.CheckPassword(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	expirationTime := time.Hour
	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		expirationTime,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access JWT", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()
	refreshExpirationTime := time.Now().UTC().Add(time.Hour * 24 * 30)
	cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: refreshExpirationTime,
		UserID:    user.ID,
	})

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
