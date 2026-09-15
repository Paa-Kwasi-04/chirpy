package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Paa-Kwasi-04/chirpy/internal/auth"
	"github.com/Paa-Kwasi-04/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func Startup() *ApiConfig {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Couldn't connect to database")
		os.Exit(1)
	}

	dbQueries := database.New(db)

	var cfg = ApiConfig{
		DB:       dbQueries,
		Platform: platform,
		TokenSecret: tokenSecret,
	}
	return &cfg
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	errMessage := validateError{
		Error: msg,
	}
	errJson, _ := json.Marshal(errMessage)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(errJson))
}

func respondWithJson(w http.ResponseWriter, statusCode int, payload any) {

	respJson, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(respJson))
}

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func checkProfanity(body string) string {
	words := strings.Fields(body)

	for i, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}


func createUser(reqBody usersRequest, ctx context.Context, cfg *ApiConfig) (*User, error) {

	hash_password, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	createUserParam := database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      now,
		UpdatedAt:      now,
		Email:          reqBody.Email,
		HashedPassword: hash_password,
	}

	user, err := cfg.DB.CreateUser(ctx, createUserParam)
	if err != nil {
		return nil, err
	}

	// Create a response user struct to hold the response data
	responseUser := User{ 
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	return &responseUser, nil
}

func deleteusers(ctx context.Context, cfg *ApiConfig) error {
	return cfg.DB.DeleteUsers(ctx)
}

func getUser(ctx context.Context, reqBody usersRequest, cfg *ApiConfig) (*User, error) {

	user, err := cfg.DB.GetUser(ctx, reqBody.Email)
	if err != nil {
		return nil, err
	}

	isMatch, err := auth.CheckPasswordHash(reqBody.Password, user.HashedPassword)
	if err != nil {
		return nil, err
	}

	if !isMatch {
		return nil, fmt.Errorf("Incorrect email or password")
	}

	expiresIn := time.Duration(reqBody.ExpiresIn)* time.Second 

	// Generate JWT token for the user
	token,err := auth.MakeJWT(user.ID,cfg.TokenSecret,expiresIn)
	if err != nil{
		return nil,err
	}

	// Create a response user struct to hold the response data
	responseUser := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token: token,
	}

	return &responseUser, nil
}

func createChirp(ctx context.Context, cfg *ApiConfig, body string, userID uuid.UUID) (*createChirpsResponse, error) {
	now := time.Now()
	createChirpParam := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Body:      body,
		UserID:    userID,
	}

	chirp, err := cfg.DB.CreateChirp(ctx, createChirpParam)
	if err != nil {
		return nil, err
	}

	// Create a response chirp struct to hold the response data
	responseChirp := createChirpsResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

	return &responseChirp, nil
}

func getChirps(ctx context.Context, cfg *ApiConfig) ([]createChirpsResponse, error) {
	chirps, err := cfg.DB.GetChirps(ctx)
	if err != nil {
		return nil, err
	}
	
	
	// Create a slice to hold the response chirps 
	responseChirps := make([]createChirpsResponse, len(chirps))
	for i, chirp := range chirps {
		responseChirp := createChirpsResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		responseChirps[i] = responseChirp
	}
	
	return responseChirps, nil
}

func getChirp(ctx context.Context, cfg *ApiConfig, id uuid.UUID) (*createChirpsResponse, error) {

	chirp, err := cfg.DB.GetChirp(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create a response chirp struct to hold the response data
	responseChirp := createChirpsResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	return &responseChirp, nil
}
