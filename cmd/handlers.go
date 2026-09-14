package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"uuid"
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
	user, err := createUser(reqBody.Email,reqBody.Password,r.Context(), cfg)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	responseUser := User{ // this struct has json tags
		ID:        uuid.UUID(user.ID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, 201, responseUser)
}

func (cfg *ApiConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var reqBody usersRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	user,err := getUser(r.Context(),reqBody.Email,reqBody.Password,cfg)
	if err != nil{
		respondWithError(w,401,err.Error())
		return
	}

	responseUser := User{
		ID: uuid.UUID(user.ID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	respondWithJson(w,200,responseUser)
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

		parsedID,err := convertStringToUUID( reqBody.UserID.String())
		if err != nil{
			respondWithError(w,500,err.Error())
			return
		}

		chirp, err := createChirp(r.Context(), cfg, cleaned_Body, parsedID)
		if err != nil {
			respondWithError(w, 500, err.Error())
			return
		}
		responseChirp := createChirpsResponse{
			ID:        uuid.UUID(chirp.ID),
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    uuid.UUID(chirp.UserID),
		}
		respondWithJson(w, 201, responseChirp)
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

	responseChirps := make([]createChirpsResponse, len(chirps))
	for i, chirp := range chirps {
		responseChirp := createChirpsResponse{
			ID:        uuid.UUID(chirp.ID),
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    uuid.UUID(chirp.UserID),
		}
		responseChirps[i] = responseChirp
	}
	respondWithJson(w, 200, responseChirps)
}

func (cfg *ApiConfig) HandleGetChirp(w http.ResponseWriter,r *http.Request){
	chirpID := r.PathValue("chirpID")

	parsedID,err := convertStringToUUID(chirpID)
	if err != nil{
		respondWithError(w,500,err.Error())
		return
	}

	chirp,err := getChirp(r.Context(),cfg,parsedID)
	if err != nil{
		respondWithError(w,404,err.Error())
		return
	}

	responseChirp := createChirpsResponse{
		ID: uuid.UUID(chirp.ID),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: uuid.UUID(chirp.UserID),
	}
	respondWithJson(w,200,responseChirp)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
