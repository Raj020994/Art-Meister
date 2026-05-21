package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateUser struct {
	Name        string           `json:"name"`
	Email       string           `json:"email"`
	Password    string           `json:"password"`
	Description string           `json:"description"`
	Batch       string           `json:"batch"`
	Social      *json.RawMessage `json:"social"`
}
type AuthenticateUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type PatchUserProfileRequest struct {
	Name   *string          `json:"name"`
	Email  *string          `json:"email"`
	Batch  *string          `json:"batch"`
	Desc   *string          `json:"description"`
	Social *json.RawMessage `json:"social"`
}
type PatchUserPassword struct {
	OldPassword string `json:"old_password"`
	Password    string `json:"password"`
}
type PatchUserStatus struct {
	Role   string `json:"role"`
	Status string `json:"status"`
}
type CreateEvent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Venue       string `json:"venue"`
	Image       string `json:"image"`
	BannerImage string `json:"banner_image"`
	EventDate   string `json:"event_date"`
	Status      string `json:"status"`
}
type AddComment struct {
	Comment string `json:"comment"`
}
