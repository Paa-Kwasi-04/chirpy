package cmd

import (
	"sync/atomic"
	"time"
	"uuid"

	"github.com/Paa-Kwasi-04/chirpy/internal/database"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
}

type User struct {  //converts db user struct t user struct with tags
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}


//struct for POST /api/chirps
type createChirpsRequest struct{
	Body string  `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type createChirpsResponse struct{
	ID uuid.UUID  `json:"id"`
	CREATEDAT time.Time `json:"created_at"`
	UPDATEDAT time.Time `json:"updated_at"`
	BODY string  `json:"body"`
	USERID uuid.UUID  `json:"user_id"`
}


// struct for POST /api/validate_chirp"
type validateRequest struct {
	Body string `json:"body"`
}

type validateError struct {
	Error string `json:"error"`
}

type validateResponse struct {
	cleaned_Body string
}


// struct for POST /api/users
type createUsersRequest struct {
	Email string `json:"email"`
}
