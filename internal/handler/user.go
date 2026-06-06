package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/hmdnu/okane/constant"
	"github.com/hmdnu/okane/internal/dto"
	appmiddleware "github.com/hmdnu/okane/internal/middleware"
	"github.com/hmdnu/okane/internal/service"
	"github.com/hmdnu/okane/lib"
)

type UserHandler struct {
	userService *service.UserService
}

func UserHandlerInit(u *service.UserService) *UserHandler {
	return &UserHandler{userService: u}
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := dto.RegisterDto{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if err := lib.ValidateStruct(data); err != nil {
		if err := lib.SetSession(w, r, constant.FLASH_SESSION, constant.ERR_FLASH_KEY, lib.ParseFormErrors(err)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if err := u.userService.Register(&data, r.Context()); err != nil {
		formErrors := lib.ParseSQLiteError(err)
		if len(formErrors) == 0 {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := lib.SetSession(w, r, constant.FLASH_SESSION, constant.ERR_FLASH_KEY, formErrors); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	if err := lib.SetSession(w, r, constant.FLASH_SESSION, constant.SUCCESS_FLASH_KEY, "Account created! You can now log in."); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := dto.LoginDto{
		Identifier:   r.FormValue("identifier"),
		Password:     r.FormValue("password"),
		StayLoggedIn: r.FormValue("stay_logged_in") == "on",
	}

	if err := lib.ValidateStruct(data); err != nil {
		if err := lib.SetSession(w, r, constant.FLASH_SESSION, constant.ERR_FLASH_KEY, lib.ParseFormErrors(err)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := u.userService.Login(&data, r.Context())
	if err != nil {
		formErrors := []lib.FormError{{Field: "email", Rule: "invalid", Message: err.Error()}}
		if setErr := lib.SetSession(w, r, constant.FLASH_SESSION, constant.ERR_FLASH_KEY, formErrors); setErr != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	maxAge := lib.SESSION_EXPIRES_24H
	if data.StayLoggedIn {
		maxAge = lib.SESSION_EXPIRES_REMEMBERED
	}

	if err := lib.SetSessionWithMaxAge(w, r, constant.AUTH_SESSION, constant.USER_ID_KEY, userID, maxAge); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (u *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := lib.ClearSession(w, r, constant.AUTH_SESSION); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (u *UserHandler) SaveSalarySetting(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dayOfMonth, err := strconv.Atoi(r.FormValue("dayOfMonth"))
	if err != nil || dayOfMonth < 1 || dayOfMonth > 31 {
		http.Error(w, "salary date must be between 1 and 31", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "salary amount must be greater than 0", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.ParseUint(r.FormValue("categoryId"), 10, 64)
	if err != nil || categoryID == 0 {
		http.Error(w, "salary category must be valid", http.StatusBadRequest)
		return
	}

	err = u.userService.SaveSalarySetting(r.Context(), userID, dto.SalarySettingData{
		Enabled:    r.FormValue("enabled") == "on",
		DayOfMonth: dayOfMonth,
		Amount:     amount,
		CategoryID: uint(categoryID),
	})
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			http.Error(w, "salary category not found", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (u *UserHandler) UserSettingView(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data, err := u.userService.GetUserSetting(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if profile, ok := appmiddleware.SidebarProfileFromCtx(r.Context()); ok {
		data.Profile = profile
	}

	u.userService.RenderUserSetting(w, r, data)
}

func (u *UserHandler) RegisterView(w http.ResponseWriter, r *http.Request) {
	u.userService.RenderRegister(w, r, setSessionErr(w, r))
}

func (u *UserHandler) LoginView(w http.ResponseWriter, r *http.Request) {
	u.userService.RenderLogin(w, r, setSessionErr(w, r), setSessionSuccess(w, r))
}

func setSessionErr(w http.ResponseWriter, r *http.Request) []lib.FormError {
	validationErrors := make([]lib.FormError, 0)

	if data, ok := lib.GetSession(r, constant.FLASH_SESSION, constant.ERR_FLASH_KEY); ok {
		if err, ok := data.([]lib.FormError); ok {
			validationErrors = err
		}
		if err := lib.ClearSession(w, r, constant.FLASH_SESSION); err != nil {
			log.Println(err)
		}
	}

	return validationErrors
}

func setSessionSuccess(w http.ResponseWriter, r *http.Request) string {
	data, ok := lib.GetSession(r, constant.FLASH_SESSION, constant.SUCCESS_FLASH_KEY)
	if !ok {
		return ""
	}

	msg, ok := data.(string)
	if !ok {
		return ""
	}

	if err := lib.ClearSession(w, r, constant.FLASH_SESSION); err != nil {
		log.Println(err)
	}

	return msg
}
