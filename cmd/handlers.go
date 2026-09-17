package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Paa-Kwasi-04/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	hits := fmt.Sprintf(`
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
	`, cfg.fileserverHits.Load())
	w.Write([]byte(hits))
}

const dev_platform = "dev"

func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {

	if cfg.Platform != dev_platform {
		w.WriteHeader(403)
		w.Write([]byte("403 Forbidden"))
		return
	}

	err := deleteusers(r.Context(), cfg)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users successfully deleted"))
}

func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var reqBody usersRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user, err := createUser(reqBody, r.Context(), cfg)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithJson(w, 201, user)
}

func (cfg *ApiConfig) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {

	var reqBody usersRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	// Get the bearer token from the request headers
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	// Validate the JWT token and extract the user ID ie authorization
	userID, err := auth.ValidateJWT(accessToken, cfg.TokenSecret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	// Update the user in the database
	updatedUser, err := updateUserLogin(userID, reqBody, r.Context(), cfg)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	respondWithJson(w, 200, updatedUser)
}

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {

	var reqBody usersRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user, err := getUser(r.Context(), reqBody, cfg)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	respondWithJson(w, 200, user)
}

func (cfg *ApiConfig) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	// Get the user associated with the refresh token from the database
	user, err := getUserFromRefreshToken(r.Context(), refreshToken, cfg)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	// Check if the refresh token has expired or been revoked
	remainingTime := user.ExpiresAt.Sub(time.Now())
	if remainingTime <= 0 {
		respondWithError(w, 401, "refresh token has expired")
		return
	}

	if user.RevokedAt.Valid {
		respondWithError(w, 401, "refresh token has been revoked")
		return
	}

	// Generate JWT token for the user
	token, err := auth.MakeJWT(user.ID, cfg.TokenSecret, TokenExpiresIn)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	newAccessToken := accessTokenRequest{
		Token: token,
	}
	respondWithJson(w, 200, newAccessToken)

}

func (cfg *ApiConfig) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	// Get the refresh token from the request headers
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	// Revoke the refresh token in the database
	err = revokeRefreshToken(r.Context(), cfg, refreshToken)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	respondWithText(w, 204, "OK")
}

func (cfg *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	var reqBody createChirpsRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	if len(reqBody.Body) <= 140 {

		cleaned_Body := checkProfanity(reqBody.Body)

		// Get the bearer token from the request headers
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, 401, err.Error())
			return
		}

		// Validate the JWT token and extract the user ID
		userID, err := auth.ValidateJWT(token, cfg.TokenSecret)
		if err != nil {
			respondWithError(w, 401, err.Error())
			return
		}

		// Create the chirp in the database
		chirp, err := createChirp(r.Context(), cfg, cleaned_Body, userID)
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}
		respondWithJson(w, 201, chirp)

	} else {
		respondWithError(w, 400, "Chirp is too long")
	}
}

func (cfg *ApiConfig) HandleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := getChirps(r.Context(), cfg)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	respondWithJson(w, 200, chirps)
}

func (cfg *ApiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")

	parsedID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	chirp, err := getChirp(r.Context(), cfg, parsedID)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	respondWithJson(w, 200, chirp)
}


func (cfg *ApiConfig) HandleDeleteChirp(w http.ResponseWriter,r *http.Request){
	chirpID := r.PathValue("chirpID")

	parsedID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	chirp,err := getChirp(r.Context(),cfg,parsedID)
	if err != nil{
		respondWithError(w,500,err.Error())
		return
	}

	// Get the bearer token from the request headers
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	user_id,err := auth.ValidateJWT(accessToken,cfg.TokenSecret)
	if err != nil{
		respondWithError(w,401,err.Error())
		return
	}

	if user_id != chirp.UserID{
		respondWithError(w,403,"403 Forbidden: You are not authorized to delete this chirp")
		return
	}

	err = deleteChirp(r.Context(),cfg,chirp.ID)
	if err != nil{
		respondWithError(w,404,err.Error())
		return
	}
	respondWithText(w,204,"OK")
}

const polkaEvent = "user.upgraded"
func (cfg *ApiConfig) HandlePolkaWebhook(w http.ResponseWriter,r *http.Request){
	var reqBody polkaWebhookRequest

	apiKey,err := auth.GetAPIKey(r.Header)
	if err != nil{
		respondWithError(w,401,err.Error())
		return
	}

	if apiKey != cfg.Polka_Key{
		respondWithError(w,401,"Unauthorized Webhook Request")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	if reqBody.Event != polkaEvent{
		respondWithText(w,204,"OK")
		return
	}

	user_id,err := uuid.Parse(reqBody.Data.UserID)
	if err != nil{
		respondWithError(w,500,err.Error())
		return
	}

	err = upgradeUserToChirpRed(r.Context(),cfg,user_id)
	if err != nil{
		respondWithError(w,404,err.Error())
		return
	}
	respondWithText(w,204,"Upgrade Successful")
}


func HandleHealth(w http.ResponseWriter, r *http.Request) {
	respondWithText(w, 204, "OK")
}
