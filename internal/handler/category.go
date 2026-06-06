package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	appmiddleware "github.com/hmdnu/okane/internal/middleware"
	"github.com/hmdnu/okane/internal/model"
	"github.com/hmdnu/okane/internal/service"
	"github.com/hmdnu/okane/lib"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func CategoryHandlerInit(c *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: c}
}

func (c *CategoryHandler) CategoryManagementView(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data, err := c.categoryService.GetCategoryManagement(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if profile, ok := appmiddleware.SidebarProfileFromCtx(r.Context()); ok {
		data.Profile = profile
	}

	c.categoryService.RenderCategoryManagement(w, r, data)
}

func (c *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := model.CategoryDto{
		UserID: &userID,
		Name:   r.FormValue("name"),
	}

	if err := lib.ValidateStruct(data); err != nil {
		http.Error(w, lib.ParseFormErrors(err)[0].Message, http.StatusBadRequest)
		return
	}

	if err := c.categoryService.CreateCategory(r.Context(), data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (c *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := categoryIDFromRequest(r)
	if err != nil {
		http.Error(w, "category must be valid", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := model.CategoryDto{
		UserID: &userID,
		Name:   r.FormValue("name"),
	}

	if err := lib.ValidateStruct(data); err != nil {
		http.Error(w, lib.ParseFormErrors(err)[0].Message, http.StatusBadRequest)
		return
	}

	if err := c.categoryService.UpdateCategory(r.Context(), id, userID, data.Name); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (c *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := appmiddleware.UserIDFromCtx(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := categoryIDFromRequest(r)
	if err != nil {
		http.Error(w, "category must be valid", http.StatusBadRequest)
		return
	}

	if err := c.categoryService.DeleteCategory(r.Context(), id, userID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func categoryIDFromRequest(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}
