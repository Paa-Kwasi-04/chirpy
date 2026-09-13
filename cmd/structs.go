package cmd

import "sync/atomic"

type ApiConfig struct {
	fileserverHits atomic.Int32
}

type validateRequest struct {
	Body string `json:"body"`
}

type validateError struct {
	Error string `json:"error"`
}

type validateResponse struct {
	Cleaned_Body string `json:"cleaned_body"`
}
