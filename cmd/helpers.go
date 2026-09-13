package cmd

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

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
	return strings.Join(words," ")
}
