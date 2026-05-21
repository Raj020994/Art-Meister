package event

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type EventHandler struct {
	Repo database.EventRepository
}
type EventAttendeeHandler struct {
	Repo database.EventAttendeesRepository
}

func (h *EventHandler) HandleCreateEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}
	id := uuid.New()
	path := fmt.Sprintf("uploads/event/%s", id.String())
	name := r.FormValue("name")
	if len(name) < 3 {
		handler.RespondWithError(w, http.StatusBadRequest, "Too Short Name")
		return
	}
	form_date := r.FormValue("date")

	eventDate, err := time.Parse("2006-01-02", form_date)

	if err != nil {
		handler.RespondWithError(w, 400, "invalid date format (expected YYYY-MM-DD)")
		return
	}

	desc := r.FormValue("description")
	venue := r.FormValue("venue")

	status := r.FormValue("status")
	mode := database.ModeOfConduct(status)
	if mode == "" {
		handler.RespondWithError(w, http.StatusBadRequest, "Unknown Mode")
		return
	}
	hasUpdate := false
	params := database.CreateEventParams{
		ID:          id,
		Name:        name,
		Description: utlis.ToNilStr(&desc),
		Venue:       utlis.ToNilStr(&venue),
		Status:      mode,
		EventDate:   eventDate,
	}
	userfile, _, err := r.FormFile("image")
	if err == nil && userfile != nil {
		defer userfile.Close()
		userImageFilePath, saveErr := utlis.SaveLocal(userfile, "event-logo", path)
		if saveErr != nil {
			handler.RespondWithError(w, http.StatusInternalServerError, "Failed to save Event image")
			return
		}
		params.Image = utlis.ToNilStr(&userImageFilePath)
		hasUpdate = true
	}

	bannerFile, _, err := r.FormFile("banner_image")
	if err == nil && bannerFile != nil {
		defer bannerFile.Close()
		bannerImageFilePath, saveErr := utlis.SaveLocal(bannerFile, "banner_image", path)
		if saveErr != nil {
			handler.RespondWithError(w, http.StatusInternalServerError, "Failed to save banner image")
			return
		}
		params.BannerImage = utlis.ToNilStr(&bannerImageFilePath)
		hasUpdate = true
	}

	if !hasUpdate {
		handler.RespondWithError(w, http.StatusBadRequest, "At least one image (image or banner_image) is required")
		return
	}
	res, err := h.Repo.CreateEvent(r.Context(), params)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update images")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventHandler) HandleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	eventId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid Id")
		return
	}
	_, err = h.Repo.DeleteEvent(r.Context(), eventId)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to delete event")
		return
	}
	path := fmt.Sprintf("%s", id)
	err = utlis.DeleteLocal(path)
	if err != nil {
		fmt.Println(err.Error())
	}
	handler.RespondWithJson(w, http.StatusOK, "ok")
}
func (h *EventHandler) HandleGetEventById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	eventId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Invalid Id")
		return
	}
	res, err := h.Repo.GetEventByID(r.Context(), eventId)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to get event")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventHandler) HandleGetAllEvent(w http.ResponseWriter, r *http.Request) {
	res, err := h.Repo.ListEvents(r.Context())
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to list events")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventHandler) HandleUpdateEvent(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}
	eventId := chi.URLParam(r, "id")
	id, err := uuid.Parse(eventId)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	path := fmt.Sprintf("uploads/event/%s", id.String())
	name := r.FormValue("name")
	if len(name) < 3 {
		handler.RespondWithError(w, http.StatusBadRequest, "Too Short Name")
		return
	}
	form_date := r.FormValue("date")

	eventDate, err := time.Parse("2006-01-02", form_date)

	if err != nil {
		handler.RespondWithError(w, 400, "invalid date format (expected YYYY-MM-DD)")
		return
	}

	desc := r.FormValue("description")
	venue := r.FormValue("venue")

	status := r.FormValue("status")
	mode := database.ModeOfConduct(status)
	if mode == "" {
		handler.RespondWithError(w, http.StatusBadRequest, "Unknown Mode")
		return
	}
	hasUpdate := false
	params := database.UpdateEventParams{
		ID:          id,
		Name:        name,
		Description: utlis.ToNilStr(&desc),
		Venue:       utlis.ToNilStr(&venue),
		Status:      mode,
		EventDate:   eventDate,
	}
	userfile, _, err := r.FormFile("image")
	if err == nil && userfile != nil {
		defer userfile.Close()
		userImageFilePath, saveErr := utlis.SaveLocal(userfile, "event-logo", path)
		if saveErr != nil {
			handler.RespondWithError(w, http.StatusInternalServerError, "Failed to save Event image")
			return
		}
		params.Image = utlis.ToNilStr(&userImageFilePath)
		hasUpdate = true
	}

	bannerFile, _, err := r.FormFile("banner_image")
	if err == nil && bannerFile != nil {
		defer bannerFile.Close()
		bannerImageFilePath, saveErr := utlis.SaveLocal(bannerFile, "banner_image", path)
		if saveErr != nil {
			handler.RespondWithError(w, http.StatusInternalServerError, "Failed to save banner image")
			return
		}
		params.BannerImage = utlis.ToNilStr(&bannerImageFilePath)
		hasUpdate = true
	}

	if !hasUpdate {
		handler.RespondWithError(w, http.StatusBadRequest, "At least one image (image or banner_image) is required")
		return
	}
	res, err := h.Repo.UpdateEvent(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Event not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update event")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventAttendeeHandler) HandleJoinEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, http.StatusUnauthorized, "Not Authorized")
		return
	}
	userId := user.ID
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.EnrollUserToEventParams{
		ID:      uuid.New(),
		EventID: event_id,
		UserID:  userId,
	}
	res, err := h.Repo.EnrollUserToEvent(r.Context(), param)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
func (h *EventAttendeeHandler) HandleDeleteEventAttendee(w http.ResponseWriter, r *http.Request) {
	user_id := r.URL.Query().Get("user_id")
	userId, err := uuid.Parse(user_id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.RemoveUserFromEventParams{
		EventID: event_id,
		UserID:  userId,
	}
	_, err = h.Repo.RemoveUserFromEvent(r.Context(), param)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Attendance record not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to remove user from event")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, "ok")
}
func (h *EventAttendeeHandler) HandleAllEventAttendee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	event_id, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	res, err := h.Repo.ListEventAttendees(r.Context(), event_id)

	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to list event attendees")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, res)
}
