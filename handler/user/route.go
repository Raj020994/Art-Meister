package user

import (
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
)

func UserRouter(userHandler *Handler, middlewareHandler *middleware.Handler) *chi.Mux {
	r := chi.NewRouter()

	// Public routes
	r.Post("/users", userHandler.HandleCreateUser)
	r.Post("/login", userHandler.HandleLogin)

	// Protected routes
	auth := middlewareHandler.MiddlewareAuth
	r.Patch("/users/avatar", auth(http.HandlerFunc(userHandler.HandleImageChange)))
	r.Patch("/users/{id}", auth(http.HandlerFunc(userHandler.HandleUpdateUserProfile)))
	r.Patch("/users/password", auth(http.HandlerFunc(userHandler.HandlePasswordChange)))
	r.Post("/logout", auth(http.HandlerFunc(userHandler.HandleLogOut)))

	return r
}
