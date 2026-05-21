package middleware

import (
	"context"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/google/uuid"
)

const admin contextKey = "admin"

func GetAdmin(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(admin).(User)
	return user, ok
}

func WithAdmin(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, admin, u)
}

func (h Handler) MiddlewareAdminAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("authToken")
		if err != nil {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: login required")
			return
		}

		userId, err := utlis.GetUserIdJwt(tokenCookie)
		if err != nil {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: invalid or expired token")
			return
		}

		id, err := uuid.Parse(userId)
		if err != nil {
			handler.RespondWithError(w, http.StatusBadRequest, "Invalid user ID format")
			return
		}

		dbUser, err := h.Repo.GetUser(r.Context(), id)
		if err != nil {
			if utlis.IsNotFound(err) {
				handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized: user not found")
				return
			}
			handler.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if dbUser.Role != database.UserRoleAdmin {
			handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized:Admin not Found")
			return
		}
		user := User{
			ID:          dbUser.ID,
			Name:        dbUser.Name,
			Email:       dbUser.Email,
			Batch:       dbUser.Batch,
			Status:      dbUser.Status,
			Role:        dbUser.Role,
			Image:       dbUser.Image,
			BannerImage: dbUser.BannerImage,
		}
		ctx := context.WithValue(r.Context(), admin, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
