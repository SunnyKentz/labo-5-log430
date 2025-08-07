package models

import "time"

type ApiError struct {
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
	Success   bool      `json:"success"`
}

type ApiSuccess struct {
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
	Success   bool      `json:"success"`
}
