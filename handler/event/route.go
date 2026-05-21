package event

import (
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
)

// EventRouter defines routes for the event package.
func EventRouter(eventHandler *EventHandler, attendeeHandler *EventAttendeeHandler, middlewareHandler *middleware.Handler) *chi.Mux {
	r := chi.NewRouter()
	auth := middlewareHandler.MiddlewareAuth
	adminAuth := middlewareHandler.MiddlewareAdminAuth

	// Public routes
	r.Get("/", eventHandler.HandleGetAllEvent)
	r.Get("/{id}", eventHandler.HandleGetEventById)

	// Protected routes (require user authentication)
	r.Post("/{id}/join", auth(http.HandlerFunc(attendeeHandler.HandleJoinEvent)))
	r.Delete("/{id}/attendee", auth(http.HandlerFunc(attendeeHandler.HandleDeleteEventAttendee)))
	r.Get("/{id}/attendees", auth(http.HandlerFunc(attendeeHandler.HandleAllEventAttendee)))

	// Admin-only routes (require admin authentication)
	r.Post("/", adminAuth(http.HandlerFunc(eventHandler.HandleCreateEvent)))
	r.Patch("/{id}", adminAuth(http.HandlerFunc(eventHandler.HandleUpdateEvent)))
	r.Delete("/{id}", adminAuth(http.HandlerFunc(eventHandler.HandleDeleteEvent)))

	return r
}
