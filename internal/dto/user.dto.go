package dto

type UserSettingData struct {
	Profile       SidebarProfile
	Username      string
	Email         string
	CreatedAt     string
	UpdatedAt     string
	SalarySetting SalarySettingData
	Categories    []DashboardCategory
}

type SidebarProfile struct {
	Username string
	Email    string
}

type SalarySettingData struct {
	Enabled    bool
	DayOfMonth int
	Amount     float64
	CategoryID uint
}
