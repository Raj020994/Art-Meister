package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/google/uuid"
)

type mockAuthRepo struct {
	database.UserRepository
	users     map[uuid.UUID]database.User
	checkErr  error
}

func (m *mockAuthRepo) CheckUsrById(ctx context.Context, id uuid.UUID) (database.CheckUsrByIdRow, error) {
	if m.checkErr != nil {
		return database.CheckUsrByIdRow{}, m.checkErr
	}
	u, ok := m.users[id]
	if !ok {
		return database.CheckUsrByIdRow{}, sql.ErrNoRows
	}
	return database.CheckUsrByIdRow{
		ID:     u.ID,
		Status: u.Status,
		Role:   u.Role,
	}, nil
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		users: make(map[uuid.UUID]database.User),
	}
}

func seedAuthUser(repo *mockAuthRepo, overrides ...func(*database.User)) (uuid.UUID, string) {
	id := uuid.New()
	u := database.User{
		ID:     id,
		Name:   "Test User",
		Email:  "test@example.com",
		Status: database.AccountStatusApproved,
		Role:   database.UserRoleUser,
	}
	for _, o := range overrides {
		o(&u)
	}
	repo.users[id] = u

	token, _ := utlis.GenerateJwt(id)
	return id, token
}

func TestMiddlewareAuth(t *testing.T) {
	repo := newMockAuthRepo()
	h := &Handler{Repo: repo}

	userID, validToken := seedAuthUser(repo)
	seedAuthUser(repo, func(u *database.User) {
		u.Role = database.UserRoleAdmin
	})

	tests := []struct {
		name           string
		cookie         *http.Cookie
		mockErr        error
		expectedStatus int
	}{
		{
			name: "authenticated user",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: validToken,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no cookie",
			cookie:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: "invalid-token-value",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: func() string { t, _ := utlis.GenerateJwt(uuid.New()); return t }(),
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "db error",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: validToken,
			},
			mockErr:        sql.ErrConnDone,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.checkErr = tc.mockErr

			nextCalled := false
			handler := h.MiddlewareAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				u, ok := GetUser(r.Context())
				if !ok {
					t.Errorf("expected user in context")
				}
				if u.ID != userID && tc.name == "authenticated user" {
					t.Errorf("expected user ID %s, got %s", userID, u.ID)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			if tc.expectedStatus == http.StatusOK && !nextCalled {
				t.Errorf("expected next handler to be called")
			}
			if tc.expectedStatus != http.StatusOK && nextCalled {
				t.Errorf("expected next handler NOT to be called")
			}
		})
	}
}

func TestMiddlewareAdminAuth(t *testing.T) {
	repo := newMockAuthRepo()
	h := &Handler{Repo: repo}

	_, adminToken := seedAuthUser(repo, func(u *database.User) {
		u.Role = database.UserRoleAdmin
	})
	_, userToken := seedAuthUser(repo, func(u *database.User) {
		u.Role = database.UserRoleUser
	})
	_, bannedToken := seedAuthUser(repo, func(u *database.User) {
		u.Role = database.UserRoleAdmin
		u.Status = database.AccountStatusBanned
	})

	tests := []struct {
		name           string
		cookie         *http.Cookie
		expectedStatus int
	}{
		{
			name: "admin access granted",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: adminToken,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "regular user denied",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: userToken,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "banned admin denied",
			cookie: &http.Cookie{
				Name:  "authToken",
				Value: bannedToken,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no cookie",
			cookie:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			handler := h.MiddlewareAdminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}

			if tc.expectedStatus == http.StatusOK && !nextCalled {
				t.Errorf("expected next handler to be called")
			}
			if tc.expectedStatus != http.StatusOK && nextCalled {
				t.Errorf("expected next handler NOT to be called")
			}
		})
	}
}
