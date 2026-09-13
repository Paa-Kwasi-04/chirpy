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
	var reqBody createUsersRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	user, err := createUser(reqBody.Email, r.Context(), cfg)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}
	
	responseUser := User{  // this struct has json tags
		ID: uuid.UUID(user.ID),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}

	respondWithJson(w, 201, responseUser)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func HandleValidate(w http.ResponseWriter, r *http.Request) {

	var reqBody validateRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	if len(reqBody.Body) <= 140 {
		respBody := validateResponse{
			cleaned_Body: checkProfanity(reqBody.Body),
		}
		respondWithJson(w, 200, respBody)
	} else {
		respondWithError(w, 400, "Chirp is too long")
	}

}
