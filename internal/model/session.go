package model

import "time"

type Session struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	TokenHash string    `json:"tokenHash"`
	ExpiresAt time.Time `json:"expiresAt"`
}
