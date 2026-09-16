package cmd

import (
	"sync/atomic"
	"time"

	"github.com/Paa-Kwasi-04/chirpy/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	TokenSecret    string
}

type userLogin struct { //converts db user struct t user struct with tags
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type userCreate struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// struct for POST /api/chirps
type createChirpsRequest struct {
	Body string `json:"body"`
}

type createChirpsResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
type usersRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

// struct for POST /api/refresh
type accessTokenRequest struct {
	Token string `json:"token"`
}
