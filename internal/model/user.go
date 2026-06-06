package model

import "time"

type User struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"-"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
