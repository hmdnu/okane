package model

import "time"

type TransactionType string

const (
	TransactionTypeIncome         TransactionType = "income"
	TransactionTypeExpense        TransactionType = "expense"
	TransactionTypeInitialBalance TransactionType = "initial_balance"
)

type TransactionDto struct {
	UserID     uint            `form:"userId" json:"userId" validate:"required,gt=0"`
	CategoryID *uint           `form:"categoryId" json:"categoryId" validate:"required,gt=0"`
	Name       string          `form:"name" json:"name" validate:"required,min=1,max=100"`
	Amount     float64         `form:"amount" json:"amount" validate:"required,gt=0"`
	Type       TransactionType `form:"type" json:"type" validate:"required,oneof=income expense initial_balance"`
	Note       string          `form:"note" json:"note" validate:"max=500"`
}

type Transaction struct {
	ID         uint            `json:"id"`
	UserID     uint            `json:"userId"`
	CategoryID *uint           `json:"categoryId"`
	Name       string          `json:"name"`
	Amount     float64         `json:"amount"`
	Type       TransactionType `json:"type"`
	Note       string          `json:"note"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	DeletedAt  *time.Time      `json:"deletedAt,omitempty"`
}
