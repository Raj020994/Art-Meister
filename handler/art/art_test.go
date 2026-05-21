package art

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mockArtRepo struct {
	database.ArtRepository
	arts         map[uuid.UUID]database.Art
	createErr    error
	getErr       error
	updateErr    error
	deleteErr    error
}

func (m *mockArtRepo) CreateArt(ctx context.Context, arg database.CreateArtParams) (database.Art, error) {
	if m.createErr != nil {
		return database.Art{}, m.createErr
	}
	a := database.Art{
		ID:          arg.ID,
		Name:        arg.Name,
		Description: arg.Description,
		Image:       arg.Image,
		Tags:        arg.Tags,
		Status:      database.ArtStatusPending,
		UserID:      arg.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.arts[arg.ID] = a
	return a, nil
}

func (m *mockArtRepo) GetArtByID(ctx context.Context, id uuid.UUID) (database.Art, error) {
	if m.getErr != nil {
		return database.Art{}, m.getErr
	}
	a, ok := m.arts[id]
	if !ok {
		return database.Art{}, sql.ErrNoRows
	}
	return a, nil
}

func (m *mockArtRepo) GetArtByUser(ctx context.Context, userID uuid.UUID) ([]database.Art, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []database.Art
	for _, a := range m.arts {
		if a.UserID == userID {
			res = append(res, a)
		}
	}
	return res, nil
}

func (m *mockArtRepo) UpdateArt(ctx context.Context, arg database.UpdateArtParams) (database.Art, error) {
	if m.updateErr != nil {
		return database.Art{}, m.updateErr
	}
	a, ok := m.arts[arg.ID]
	if !ok {
		return database.Art{}, sql.ErrNoRows
	}
	if a.UserID != arg.UserID {
		return database.Art{}, sql.ErrNoRows // SQL update limits to user_id
	}
	a.Name = arg.Name
	a.Tags = arg.Tags
	if arg.Description.Valid {
		a.Description = arg.Description
	}
	a.UpdatedAt = time.Now()
	m.arts[arg.ID] = a
	return a, nil
}

func (m *mockArtRepo) DeleteArt(ctx context.Context, arg database.DeleteArtParams) (uuid.UUID, error) {
	if m.deleteErr != nil {
		return uuid.UUID{}, m.deleteErr
	}
	a, ok := m.arts[arg.ID]
	if !ok || a.UserID != arg.UserID {
		return uuid.UUID{}, sql.ErrNoRows
	}
	delete(m.arts, arg.ID)
	return arg.ID, nil
}

func newMockArtRepo() *mockArtRepo {
	return &mockArtRepo{
		arts: make(map[uuid.UUID]database.Art),
	}
}

func TestHandleGetArtById(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artUUID := uuid.New()
	userUUID := uuid.New()
	artSeed := database.Art{
		ID:     artUUID,
		Name:   "Mona Lisa",
		UserID: userUUID,
		Status: database.ArtStatusApproved,
	}
	repo.arts[artUUID] = artSeed

	tests := []struct {
		name           string
		artIDParam     string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Success Retrieve",
			artIDParam:     artUUID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID Format",
			artIDParam:     "bad-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent Art",
			artIDParam:     uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty ID Param",
			artIDParam:     "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "DB Error",
			artIDParam:     artUUID.String(),
			mockErr:        fmt.Errorf("connection refused"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/"+tc.artIDParam, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleGetArtById(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleGetArts(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	userUUID := uuid.New()
	artUUID := uuid.New()
	repo.arts[artUUID] = database.Art{
		ID:     artUUID,
		Name:   "Sunset",
		UserID: userUUID,
	}

	tests := []struct {
		name           string
		userIDParam    string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Success List User Art",
			userIDParam:    userUUID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid User UUID",
			userIDParam:    "bad-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "No Art for User (Empty List)",
			userIDParam:    uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty User ID Param",
			userIDParam:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "DB Error",
			userIDParam:    userUUID.String(),
			mockErr:        fmt.Errorf("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.getErr = tc.mockErr

			req := httptest.NewRequest(http.MethodGet, "/art/user/"+tc.userIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleGetArts(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleArtDeletion(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artUUID := uuid.New()
	userUUID := uuid.New()
	artSeed := database.Art{
		ID:     artUUID,
		Name:   "Mona Lisa",
		UserID: userUUID,
		Status: database.ArtStatusApproved,
	}
	repo.arts[artUUID] = artSeed

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User // nil = unauthenticated
		mockErr        error
		expectedStatus int
	}{
		{
			name:       "Success Delete Own Art",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Fail Delete Other User's Art",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: uuid.New(),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unauthenticated Delete",
			artIDParam:     artUUID.String(),
			authCtxUser:    nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid UUID Param",
			artIDParam: "not-a-uuid",
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Non-existent Art ID",
			artIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "DB Internal Error",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			mockErr:        fmt.Errorf("disk full"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset repository seed
			repo.arts[artUUID] = artSeed
			repo.deleteErr = tc.mockErr

			req := httptest.NewRequest(http.MethodDelete, "/art/"+tc.artIDParam, nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleArtDeletion(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandlerArtUpdation(t *testing.T) {
	repo := newMockArtRepo()
	h := &Handler{Repo: repo}

	artUUID := uuid.New()
	userUUID := uuid.New()
	artSeed := database.Art{
		ID:     artUUID,
		Name:   "Mona Lisa",
		UserID: userUUID,
		Status: database.ArtStatusApproved,
	}
	repo.arts[artUUID] = artSeed

	tests := []struct {
		name           string
		artIDParam     string
		authCtxUser    *middleware.User // nil = unauthenticated
		formDataName   string
		formDataDesc   string
		expectedStatus int
	}{
		{
			name:       "Success Update Own Art",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			formDataName:   "NewName",
			formDataDesc:   "NewDescription",
			expectedStatus: http.StatusOK,
		},
		{
			name:       "Name Too Short",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			formDataName:   "Hi",
			formDataDesc:   "NewDescription",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Update Other User Art fails",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: uuid.New(),
			},
			formDataName:   "ValidName",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unauthenticated Update",
			artIDParam:     artUUID.String(),
			authCtxUser:    nil,
			formDataName:   "ValidName",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid UUID Param",
			artIDParam: "not-a-uuid",
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			formDataName:   "ValidName",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Empty Name Field",
			artIDParam: artUUID.String(),
			authCtxUser: &middleware.User{
				ID: userUUID,
			},
			formDataName:   "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.arts[artUUID] = artSeed

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/art/%s?name=%s&description=%s", tc.artIDParam, tc.formDataName, tc.formDataDesc), nil)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandlerArtUpdation(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
