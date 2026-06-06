package model

import "time"

type CategoryDto struct {
	UserID *uint  `form:"userId" json:"userId" validate:"omitempty,gt=0"`
	Name   string `form:"name" json:"name" validate:"required,min=1,max=100"`
}

type Category struct {
	ID        uint       `json:"id"`
	UserID    *uint      `json:"userId"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
