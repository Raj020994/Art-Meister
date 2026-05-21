package art

import (
	"fmt"
	"net/http"

	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type Handler struct {
	Repo database.ArtRepository
}

func (h *Handler) HandleArtCreation(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}
	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, "Image file is required")
		return
	}
	defer file.Close()

	if fileHeader != nil && fileHeader.Size > 5<<20 {
		handler.RespondWithError(w, http.StatusRequestEntityTooLarge, "File too large")
		return
	}

	name := r.FormValue("name")
	if len(name) < 3 {
		handler.RespondWithError(w, http.StatusBadRequest, "Name is too Short")
		return
	}
	desc := r.FormValue("description")
	tags := r.MultipartForm.Value["tags"]
	if tags == nil {
		tags = []string{}
	}
	path := fmt.Sprintf("uploads/%s/art", user.ID.String())
	id := uuid.New()
	url, err := utlis.SaveLocal(file, id.String(), path)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}
	params := database.CreateArtParams{
		ID:          id,
		Name:        name,
		Description: utlis.ToNilStr(&desc),
		Tags:        tags,
		UserID:      user.ID,
		Image:       url,
	}
	art, err := h.Repo.CreateArt(r.Context(), params)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	handler.RespondWithJson(w, http.StatusOK, art)
}
func (h *Handler) HandleGetArts(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "user_id")
	if userId == "" {
		handler.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	id, err := uuid.Parse(userId)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	arts, err := h.Repo.GetArtByUser(r.Context(), id)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to get user arts")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, arts)
}
func (h *Handler) HandleGetArtById(w http.ResponseWriter, r *http.Request) {
	Id := chi.URLParam(r, "id")
	if Id == "" {
		handler.RespondWithError(w, http.StatusBadRequest, "Art ID is required")
		return
	}
	artId, err := uuid.Parse(Id)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	art, err := h.Repo.GetArtByID(r.Context(), artId)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Art not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to get art")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, art)
}
func (h *Handler) HandleArtDeletion(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	userId := user.ID.String()
	id := chi.URLParam(r, "id")
	if id == "" {
		handler.RespondWithError(w, http.StatusBadRequest, "Art ID is required")
		return
	}
	artId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	param := database.DeleteArtParams{
		ID:     artId,
		UserID: user.ID,
	}
	_, err = h.Repo.DeleteArt(r.Context(), param)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Art not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to delete art")
		return
	}
	path := fmt.Sprintf("%s/art/%s.png", userId, id)
	err = utlis.DeleteLocal(path)
	if err != nil {
		fmt.Println(err.Error())
	}
	handler.RespondWithJson(w, http.StatusOK, "Art Work Deleted")

}
func (h *Handler) HandlerArtUpdation(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, http.StatusUnauthorized, "Not Authorized")
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		handler.RespondWithError(w, http.StatusBadRequest, "Art ID is required")
		return
	}
	artId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	name := r.FormValue("name")
	if len(name) < 3 {
		handler.RespondWithError(w, http.StatusBadRequest, "Name is too Short")
		return
	}
	desc := r.FormValue("description")
	var tags []string
	if r.MultipartForm != nil {
		tags = r.MultipartForm.Value["tags"]
	}
	if tags == nil {
		tags = []string{}
	}

	params := database.UpdateArtParams{
		ID:          artId,
		UserID:      user.ID,
		Name:        name,
		Tags:        tags,
		Description: utlis.ToNilStr(&desc),
	}
	updatedWork, err := h.Repo.UpdateArt(r.Context(), params)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Art not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to update art")
		return
	}

	handler.RespondWithJson(w, http.StatusOK, updatedWork)
}
