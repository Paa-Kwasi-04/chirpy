package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	// TokenTypeAccess -
	TokenTypeAccess TokenType = "chirpy-access"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	//creating claim
	now := time.Now()
	RegisteredClaims := jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userID.String(),
	}

	// 2. Create token with claims and specify signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, RegisteredClaims)

	// 3. Sign the token with your secret key
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	// 1. Parse the token string and validate it
	//first arg is the tokenString from session
	//second arg is a pointer to an empty claims struct yet to be populated
	//third arg is the call back func that returns the secret key the algo should use
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})

	if err != nil {
		return uuid.Nil, err
	}

	// 2. Extract the claims and validate them
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {

	// Extract the Authorization header from the request headers
	authHeader:= headers.Get("Authorization")

	// Check if the Authorization header is present and starts with "Bearer "
	if !strings.HasPrefix(authHeader,"Bearer"){
		return "",fmt.Errorf("Unauthorized: Missing or invalid token format")
	}

	// Remove the "Bearer " prefix from the token string
	token := strings.TrimSpace(strings.TrimPrefix(authHeader,"Bearer"))
	return token,nil
}
