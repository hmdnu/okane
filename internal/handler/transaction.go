package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hmdnu/okane/constant"
	"github.com/hmdnu/okane/internal/dto"
	appmiddleware "github.com/hmdnu/okane/internal/middleware"
	"github.com/hmdnu/okane/internal/model"
	"github.com/hmdnu/okane/internal/service"
	"github.com/hmdnu/okane/lib"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
}

func TransactionHandlerInit(t *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactionService: t}
}

func (t *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil {
		t.redirectWithTransactionErrors(w, r, []lib.FormError{{
			Field:   "amount",
			Rule:    "number",
			Message: "amount must be a valid number",
		}})
		return
	}

	if r.FormValue("categoryId") == "" {
		t.redirectWithTransactionErrors(w, r, []lib.FormError{{
			Field:   "categoryId",
			Rule:    "required",
			Message: "categoryId is required",
		}})
		return
	}

	parsedCategoryID, err := strconv.ParseUint(r.FormValue("categoryId"), 10, 64)
	if err != nil {
		t.redirectWithTransactionErrors(w, r, []lib.FormError{{
			Field:   "categoryId",
			Rule:    "number",
			Message: "category must be valid",
		}})
		return
	}
	categoryID := uint(parsedCategoryID)

	data := model.TransactionDto{
		UserID:     userID,
		CategoryID: &categoryID,
		Name:       r.FormValue("name"),
		Amount:     amount,
		Type:       model.TransactionType(r.FormValue("type")),
		Note:       r.FormValue("note"),
	}

	if err := lib.ValidateStruct(data); err != nil {
		t.redirectWithTransactionErrors(w, r, lib.ParseFormErrors(err))
		return
	}

	if err := t.transactionService.CreateTransaction(r.Context(), data); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			t.redirectWithTransactionErrors(w, r, []lib.FormError{{
				Field:   "categoryId",
				Rule:    "not_found",
				Message: "category not found",
			}})
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (t *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := transactionIDFromRequest(r)
	if err != nil {
		http.Error(w, "transaction must be valid", http.StatusBadRequest)
		return
	}

	if err := t.transactionService.DeleteTransaction(r.Context(), id, userID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (t *TransactionHandler) DashboardView(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	filter := dashboardFilterFromRequest(r)
	data, err := t.transactionService.GetDashboard(r.Context(), userID, filter)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if profile, ok := appmiddleware.SidebarProfileFromCtx(r.Context()); ok {
		data.Profile = profile
	}
	data.TransactionErrors = setSessionErr(w, r)

	t.transactionService.RenderDashboard(w, r, data)
}

func (t *TransactionHandler) redirectWithTransactionErrors(w http.ResponseWriter, r *http.Request, formErrors []lib.FormError) {
	if err := lib.SetSession(w, r, constant.FLASH_SESSION, constant.ERR_FLASH_KEY, formErrors); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func transactionIDFromRequest(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func dashboardFilterFromRequest(r *http.Request) dto.DashboardFilter {
	selectedDate, err := time.ParseInLocation("2006-01-02", r.URL.Query().Get("date"), time.Local)
	if err != nil {
		return dto.DashboardFilter{}
	}

	return dto.DashboardFilter{
		Day:      selectedDate.Day(),
		Month:    int(selectedDate.Month()),
		Year:     selectedDate.Year(),
		IsActive: true,
	}
}
