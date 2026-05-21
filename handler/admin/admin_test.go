package admin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mockAdminRepo struct {
	database.UserRepository
	database.ArtRepository
	users     map[uuid.UUID]database.User
	arts      map[uuid.UUID]database.Art
	patchErr  error
	statusErr error
}

func (m *mockAdminRepo) PatchUserAdmin(ctx context.Context, arg database.PatchUserAdminParams) (database.PatchUserAdminRow, error) {
	if m.patchErr != nil {
		return database.PatchUserAdminRow{}, m.patchErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return database.PatchUserAdminRow{}, sql.ErrNoRows
	}
	u.Role = arg.Role
	u.Status = arg.Status
	m.users[arg.ID] = u
	return database.PatchUserAdminRow{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Batch:     u.Batch,
		Status:    u.Status,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockAdminRepo) UpdateArtStatus(ctx context.Context, arg database.UpdateArtStatusParams) (database.Art, error) {
	if m.statusErr != nil {
		return database.Art{}, m.statusErr
	}
	a, ok := m.arts[arg.ID]
	if !ok {
		return database.Art{}, sql.ErrNoRows
	}
	a.Status = arg.Status
	m.arts[arg.ID] = a
	return a, nil
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{
		users: make(map[uuid.UUID]database.User),
		arts:  make(map[uuid.UUID]database.Art),
	}
}

func TestHandlerUserStatus(t *testing.T) {
	repo := newMockAdminRepo()
	h := &UserHandler{Repo: repo}

	userUUID := uuid.New()
	repo.users[userUUID] = database.User{
		ID:     userUUID,
		Name:   "Alice",
		Role:   database.UserRoleUser,
		Status: database.AccountStatusPending,
	}

	tests := []struct {
		name           string
		userIDParam    string
		body           any
		mockErr        error
		expectedStatus int
	}{
		{
			name:        "Success Approve User",
			userIDParam: userUUID.String(),
			body: model.PatchUserStatus{
				Role:   "user",
				Status: "approved",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Invalid Role Provided",
			userIDParam: userUUID.String(),
			body: model.PatchUserStatus{
				Role:   "invalid-role",
				Status: "approved",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Invalid Status Provided",
			userIDParam: userUUID.String(),
			body: model.PatchUserStatus{
				Role:   "user",
				Status: "invalid-status",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Both Empty",
			userIDParam: userUUID.String(),
			body: model.PatchUserStatus{
				Role:   "",
				Status: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Repo Error Handle",
			userIDParam: userUUID.String(),
			body: model.PatchUserStatus{
				Role:   "user",
				Status: "approved",
			},
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "Invalid UUID Param",
			userIDParam: "not-a-uuid",
			body: model.PatchUserStatus{
				Role:   "user",
				Status: "approved",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON Body",
			userIDParam:    userUUID.String(),
			body:           "{bad}",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.patchErr = tc.mockErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/admin/users/"+tc.userIDParam+"/status", &buf)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandlerUserStatus(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandlerArtStatus(t *testing.T) {
	repo := newMockAdminRepo()
	h := &ArtHandler{Repo: repo}

	artUUID := uuid.New()
	repo.arts[artUUID] = database.Art{
		ID:     artUUID,
		Name:   "Artwork",
		Status: database.ArtStatusPending,
	}

	tests := []struct {
		name           string
		artIDParam     string
		queryStatus    string
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Success Approve Art",
			artIDParam:     artUUID.String(),
			queryStatus:    "approved",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid Query Status",
			artIDParam:     artUUID.String(),
			queryStatus:    "invalid-status",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Repo Failure",
			artIDParam:     artUUID.String(),
			queryStatus:    "approved",
			mockErr:        errors.New("db write failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid Art UUID",
			artIDParam:     "not-a-uuid",
			queryStatus:    "approved",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty Status Param",
			artIDParam:     artUUID.String(),
			queryStatus:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.statusErr = tc.mockErr

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/admin/arts/%s/status?status=%s", tc.artIDParam, tc.queryStatus), nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("art_id", tc.artIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandlerArtStatus(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
