package dto

type RegisterDto struct {
	Username string `form:"username" json:"username" validate:"required,min=1,max=100"`
	Email    string `form:"email" json:"email" validate:"required,email,max=255"`
	Password string `form:"password" json:"password" validate:"required,min=8,max=255"`
}

type LoginDto struct {
	Identifier   string `form:"identifier" json:"identifier" validate:"required"`
	Password     string `form:"password" json:"password" validate:"required"`
	StayLoggedIn bool   `form:"stay_logged_in" json:"stayLoggedIn"`
}
