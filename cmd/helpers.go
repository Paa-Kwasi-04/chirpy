package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	
	"github.com/Paa-Kwasi-04/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func Startup() *ApiConfig {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Couldn't connect to database")
		os.Exit(1)
	}

	dbQueries := database.New(db)

	var cfg = ApiConfig{
		DB:       dbQueries,
		Platform: platform,
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

func createUser(email string, ctx context.Context, cfg *ApiConfig) (*database.User, error) {

	now := time.Now()
	createUserParam := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Email:     email,
	}

	user, err := cfg.DB.CreateUser(ctx, createUserParam)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func deleteusers(ctx context.Context, cfg *ApiConfig) error {
	return cfg.DB.DeleteUsers(ctx)
}
