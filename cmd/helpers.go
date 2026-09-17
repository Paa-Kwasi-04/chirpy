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
	polka_key := os.Getenv("POLKA_KEY")

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
		Polka_Key: polka_key,
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

func respondWithText(w http.ResponseWriter,statusCode int,text string){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
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


func createUser(reqBody usersRequest, ctx context.Context, cfg *ApiConfig) (*userCreate, error) {

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

	createdUser := userCreate{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	return &createdUser, nil
}

func updateUserLogin(user_id uuid.UUID,reqBody usersRequest, ctx context.Context, cfg *ApiConfig)(*userCreate, error){
	
	hashed_password,err := auth.HashPassword(reqBody.Password)
	if err != nil{
		return nil,err
	}
	
	now := time.Now()
	updateUserParam := database.UpdateUserLoginParams{
		Email: reqBody.Email,
		HashedPassword: hashed_password,
		UpdatedAt: now,
		ID: user_id,
	}

	user,err := cfg.DB.UpdateUserLogin(ctx,updateUserParam)
	if err != nil{
		return nil,err
	}

	updatedUser := userCreate{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Is_Chirpy_Red: user.IsChirpyRed,
	}

	return &updatedUser,nil
}

func deleteusers(ctx context.Context, cfg *ApiConfig) error {
	return cfg.DB.DeleteUsers(ctx)
}

const TokenExpiresIn = 1 * time.Hour
func getUser(ctx context.Context, reqBody usersRequest, cfg *ApiConfig) (*userLogin, error) {

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

	
	// Generate JWT token for the user
	token,err := auth.MakeJWT(user.ID,cfg.TokenSecret,TokenExpiresIn)
	if err != nil{
		return nil,err
	}

	refreshToken:= auth.MakeRefreshToken()

	refreshTokenObj ,err := createRefreshToken(ctx,refreshToken,user.ID,cfg)
	if err != nil{
		return nil,err
	}

	// Create a response user struct to hold the response data
	responseUser := userLogin{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Is_Chirpy_Red: user.IsChirpyRed,
		Token: token,
		RefreshToken: refreshTokenObj.Token,
	}

	return &responseUser, nil
}


func createRefreshToken(ctx context.Context,token string,user_ID uuid.UUID,cfg *ApiConfig)(*database.RefreshToken,error){
	
	now := time.Now()
	inSixtyDays := now.AddDate(0, 0, 60) // expires in 60days
	refreshTokenParam := database.CreateRefreshTokenParams{
		Token: token,
		CreatedAt: now,
		UpdatedAt: now,
		UserID: user_ID,
		ExpiresAt: inSixtyDays,
	}
	
	refreshToken,err := cfg.DB.CreateRefreshToken(ctx,refreshTokenParam)
	if err != nil{
		return nil,err
	}
	return &refreshToken,nil
}

func getUserFromRefreshToken(ctx context.Context,refreshToken string,cfg *ApiConfig)(*database.GetUserFromRefreshTokenRow,error){
	userFromRefreshToken,err := cfg.DB.GetUserFromRefreshToken(ctx,refreshToken)
	if err != nil{
		return nil,err
	}
	return &userFromRefreshToken,nil
}

func revokeRefreshToken(ctx context.Context,cfg *ApiConfig,token string)error{
	now := sql.NullTime{
		Time: time.Now(),
		Valid: true,
	}
	revokeTokenParam := database.RevokeRefreshTokenParams{
		RevokedAt: now,
		Token: token,
	}

	return cfg.DB.RevokeRefreshToken(ctx,revokeTokenParam)
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


func deleteChirp(ctx context.Context, cfg *ApiConfig, chirp_id uuid.UUID)error{
	return cfg.DB.DeleteChirp(ctx,chirp_id)
}


//helper for the polka webhook
func upgradeUserToChirpRed(ctx context.Context,cfg *ApiConfig,user_id uuid.UUID)error{
	now := time.Now()
	upgradeToRedParam := database.UpgradeUserToChirpRedParams{
		UpdatedAt: now,
		ID: user_id,
	}

	return cfg.DB.UpgradeUserToChirpRed(ctx,upgradeToRedParam)
}
