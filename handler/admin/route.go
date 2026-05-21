package admin

import (
	"github.com/Blue-Onion/ArtmeisterBackend/handler/art"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/user"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
	"net/http"
)

func AdminRoute(userHandler *user.Handler, artHandler *art.Handler, middlewareHandler *middleware.Handler) *chi.Mux {
	r := chi.NewRouter()

	adminAuth := middlewareHandler.MiddlewareAdminAuth
	r.Use(func(next http.Handler) http.Handler {
		return adminAuth(next)
	})
	adminUserHandler := &UserHandler{Repo: userHandler.Repo}
	adminArtHandler := &ArtHandler{Repo: artHandler.Repo}

	r.Patch("/users/{id}/status", adminUserHandler.HandlerUserStatus)
	r.Patch("/arts/{art_id}/status", adminArtHandler.HandlerArtStatus)

	return r
}
