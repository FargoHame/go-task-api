package model

import (
	"errors"
)

type Task struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Done   bool   `json:"done"`
	UserID int    `json:"user_id"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" validate:"required, min=3,max=20"`
	Password string `json:"-"`
}
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Token string `json:"token"`
}

var ErrNotFound = errors.New("task not found or unauthorized")
