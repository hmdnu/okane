package dto

import (
	"time"

	"github.com/hmdnu/okane/lib"
)

type DashboardFilter struct {
	Day      int
	Month    int
	Year     int
	IsActive bool
}

type DashboardSummary struct {
	Balance         float64
	Income          float64
	Expense         float64
	AverageSpending float64
}

type DashboardTransaction struct {
	ID           uint
	Name         string
	Amount       float64
	Type         string
	Note         string
	CategoryName string
	CreatedAt    time.Time
}

type DashboardCategory struct {
	ID   uint
	Name string
}

type DashboardCategorySpending struct {
	CategoryName string
	Total        float64
}

type DashboardData struct {
	Filter            DashboardFilter
	Profile           SidebarProfile
	Summary           DashboardSummary
	Transactions      []DashboardTransaction
	Categories        []DashboardCategory
	CategorySpendings []DashboardCategorySpending
	TransactionErrors []lib.FormError
}
