package models

type UserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
