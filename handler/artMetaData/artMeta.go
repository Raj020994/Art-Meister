package artmetadata

import (
	"encoding/json"
	"github.com/Blue-Onion/ArtmeisterBackend/handler"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct {
	Repo database.ArtMetaDataRepository
}

func (h *Handler) HandleGetArtComments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	artId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	comments, err := h.Repo.GetArtCommentsByArtID(r.Context(), artId)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to get comments")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, comments)
}
func (h *Handler) HandleGetArtCommentsCount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	artId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	commentsCount, err := h.Repo.GetArtCommentsCount(r.Context(), artId)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to get comments count")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, commentsCount)
}
func (h *Handler) HandleGetArtLikeCount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	artId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	likeCount, err := h.Repo.GetArtLikesCount(r.Context(), artId)
	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to get likes count")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, likeCount)
}
func (h *Handler) HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	userId := user.ID
	id := chi.URLParam(r, "id")
	commentId, err := uuid.Parse(id)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	param := database.DeleteArtCommentParams{
		UserID: userId,
		ID:     commentId,
	}
	_, err = h.Repo.DeleteArtComment(r.Context(), param)
	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Comment not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, "Deleted Successfully")
}
func (h *Handler) HandleComment(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	id := chi.URLParam(r, "art_id")
	artId, err := uuid.Parse(id)

	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	req := model.AddComment{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	userId := user.ID
	param := database.AddArtCommentParams{
		Comment: req.Comment,
		UserID:  userId,
		ArtID:   artId,
	}
	comment, err := h.Repo.AddArtComment(r.Context(), param)

	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, comment)
}
func (h *Handler) HandleLike(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	id := chi.URLParam(r, "art_id")
	artId, err := uuid.Parse(id)

	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	userId := user.ID
	param := database.LikeArtParams{
		UserID: userId,
		ArtID:  artId,
	}
	comment, err := h.Repo.LikeArt(r.Context(), param)

	if err != nil {
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to like art")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, comment)
}
func (h *Handler) HandleUnLike(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		handler.RespondWithError(w, 401, "Not Authorized")
		return
	}
	id := chi.URLParam(r, "art_id")
	artId, err := uuid.Parse(id)

	if err != nil {
		handler.RespondWithError(w, 400, err.Error())
		return
	}
	userId := user.ID
	param := database.UnlikeArtParams{
		UserID: userId,
		ArtID:  artId,
	}
	_, err = h.Repo.UnlikeArt(r.Context(), param)

	if err != nil {
		if utlis.IsNotFound(err) {
			handler.RespondWithError(w, http.StatusNotFound, "Like not found")
			return
		}
		handler.RespondWithError(w, http.StatusInternalServerError, "Failed to unlike art")
		return
	}
	handler.RespondWithJson(w, http.StatusOK, "Ok")
}
