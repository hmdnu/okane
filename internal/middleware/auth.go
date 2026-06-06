package middleware

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/hmdnu/okane/constant"
	"github.com/hmdnu/okane/internal/dto"
	"github.com/hmdnu/okane/lib"
)

type contextKey string

const (
	userIDCtxKey  contextKey = "userID"
	profileCtxKey contextKey = "profile"
)

// AuthMiddleware rejects unauthenticated requests by redirecting to /login.
// On success it injects the user ID and sidebar profile into the request context.
func AuthMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, ok := lib.GetSession(r, constant.AUTH_SESSION, constant.USER_ID_KEY)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			userID, ok := data.(uint)
			if !ok || userID == 0 {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			profile, err := sidebarProfile(r.Context(), db, userID)
			if err != nil {
				log.Println(err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), userIDCtxKey, userID)
			ctx = context.WithValue(ctx, profileCtxKey, profile)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sidebarProfile(ctx context.Context, db *sql.DB, userID uint) (dto.SidebarProfile, error) {
	var profile dto.SidebarProfile
	err := db.QueryRowContext(
		ctx,
		`SELECT username, email
		FROM users
		WHERE id = ?
			AND deleted_at IS NULL`,
		userID,
	).Scan(&profile.Username, &profile.Email)
	return profile, err
}

// UserIDFromCtx retrieves the authenticated user ID from the request context.
func UserIDFromCtx(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(userIDCtxKey).(uint)
	return id, ok
}

func SidebarProfileFromCtx(ctx context.Context) (dto.SidebarProfile, bool) {
	profile, ok := ctx.Value(profileCtxKey).(dto.SidebarProfile)
	return profile, ok
}
