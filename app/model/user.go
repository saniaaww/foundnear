package model

import "time"

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type User struct {
	ID           int64     `json:"id"`
	RoleID       int64     `json:"role_id"`
	RoleName     string    `json:"role"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Phone        *string   `json:"phone,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
