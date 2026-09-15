package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userID,
			wantErr:     false,
		},
		{
			name:        "Invalid token",
			tokenString: "invalid.token.string",
			tokenSecret: "secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "Wrong secret",
			tokenString: validToken,
			tokenSecret: "wrong_secret",
			wantUserID:  uuid.Nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	tests := []struct {
		name        string
		header      http.Header
		wantedToken string
	}{
		{
			name: "Valid standard bearer token",
			header: http.Header{
				"Authorization": []string{"Bearer secret_token_123"},
			},
			wantedToken: "secret_token_123",
		},
		{
			name: "Case-insensitive prefix (lowercase)",
			header: http.Header{
				"Authorization": []string{"bearer secret_token_123"},
			},
			wantedToken: "secret_token_123",
		},
		{
			name: "Case-insensitive prefix (uppercase)",
			header: http.Header{
				"Authorization": []string{"BEARER secret_token_123"},
			},
			wantedToken: "secret_token_123",
		},
		{
			name: "Extra spacing between prefix and token",
			header: http.Header{
				"Authorization": []string{"Bearer    padded_token"},
			},
			wantedToken: "padded_token",
		},
		{
			name: "Missing Authorization header",
			header: http.Header{
				"X-Custom-Header": []string{"SomeValue"},
			},
			wantedToken: "",
		},
		{
			name: "Empty Authorization header value",
			header: http.Header{
				"Authorization": []string{""},
			},
			wantedToken: "",
		},
		{
			name: "Wrong authorization scheme (e.g., Basic)",
			header: http.Header{
				"Authorization": []string{"Basic dXNlcjpwYXNz"},
			},
			wantedToken: "",
		},
		{
			name: "Missing token value (prefix only)",
			header: http.Header{
				"Authorization": []string{"Bearer "},
			},
			wantedToken: "",
		},
		{
			name: "Multiple Authorization headers (takes first valid or fails)",
			header: http.Header{
				"Authorization": []string{"Bearer first_token", "Bearer second_token"},
			},
			wantedToken: "first_token",
		},
		{
			name: "Malformed header (no space separation)",
			header: http.Header{
				"Authorization": []string{"Bearernospace"},
			},
			wantedToken: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GetBearerToken(tc.header)

			if err != nil {
				t.Errorf(err.Error())
			}

			if tc.wantedToken != token {
				t.Errorf("ExtractBearerToken() = %v, want %v", token, tc.wantedToken)
			}
		})
	}
}
