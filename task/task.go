package task

import "github.com/google/uuid"

type Task struct {
	TaskID      uuid.UUID `json:"taskID"`
	UserID      uuid.UUID `json:"userID"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
}
