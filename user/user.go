package user

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID `json:"userID"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	APIKey   string    `json:"apiKey"`
}
