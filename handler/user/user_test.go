package user

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	database.UserRepository
	users        map[uuid.UUID]database.User
	emails       map[string]database.User
	createErr    error
	getErr       error
	updateErr    error
	passwordErr  error
	imagesErr    error
}

func (m *mockUserRepo) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error) {
	if m.createErr != nil {
		return database.CreateUserRow{}, m.createErr
	}
	id := uuid.New()
	u := database.User{
		ID:          id,
		Name:        arg.Name,
		Email:       arg.Email,
		Password:    arg.Password,
		Batch:       arg.Batch,
		Description: arg.Description,
		Status:      arg.Status,
		Role:        arg.Role,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.users[id] = u
	m.emails[arg.Email] = u
	return database.CreateUserRow{
		ID:          id,
		Name:        u.Name,
		Email:       u.Email,
		Batch:       u.Batch,
		Status:      u.Status,
		Role:        u.Role,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (m *mockUserRepo) GetUser(ctx context.Context, id uuid.UUID) (database.GetUserRow, error) {
	if m.getErr != nil {
		return database.GetUserRow{}, m.getErr
	}
	u, ok := m.users[id]
	if !ok {
		return database.GetUserRow{}, sql.ErrNoRows
	}
	return database.GetUserRow{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Batch:       u.Batch,
		Status:      u.Status,
		Role:        u.Role,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error) {
	u, ok := m.emails[email]
	if !ok {
		return database.GetUserByEmailRow{}, sql.ErrNoRows
	}
	return database.GetUserByEmailRow{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Password:    u.Password,
		Batch:       u.Batch,
		Status:      u.Status,
		Role:        u.Role,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (m *mockUserRepo) PatchUserProfile(ctx context.Context, arg database.PatchUserProfileParams) (database.PatchUserProfileRow, error) {
	if m.updateErr != nil {
		return database.PatchUserProfileRow{}, m.updateErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return database.PatchUserProfileRow{}, sql.ErrNoRows
	}
	if arg.Name.Valid {
		u.Name = arg.Name.String
	}
	if arg.Email.Valid {
		u.Email = arg.Email.String
	}
	m.users[arg.ID] = u
	return database.PatchUserProfileRow{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Batch:       u.Batch,
		Status:      u.Status,
		Role:        u.Role,
		Description: u.Description,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}

func (m *mockUserRepo) PatchUserPassword(ctx context.Context, arg database.PatchUserPasswordParams) (database.PatchUserPasswordRow, error) {
	if m.passwordErr != nil {
		return database.PatchUserPasswordRow{}, m.passwordErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return database.PatchUserPasswordRow{}, sql.ErrNoRows
	}
	u.Password = arg.Password
	m.users[arg.ID] = u
	return database.PatchUserPasswordRow{
		ID:        u.ID,
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserRepo) PatchUserImages(ctx context.Context, arg database.PatchUserImagesParams) (database.PatchUserImagesRow, error) {
	if m.imagesErr != nil {
		return database.PatchUserImagesRow{}, m.imagesErr
	}
	u, ok := m.users[arg.ID]
	if !ok {
		return database.PatchUserImagesRow{}, sql.ErrNoRows
	}
	if arg.Image.Valid {
		u.Image = arg.Image
	}
	if arg.BannerImage.Valid {
		u.BannerImage = arg.BannerImage
	}
	m.users[arg.ID] = u
	return database.PatchUserImagesRow{
		ID:          u.ID,
		Image:       u.Image,
		BannerImage: u.BannerImage,
		UpdatedAt:   time.Now(),
	}, nil
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[uuid.UUID]database.User),
		emails: make(map[string]database.User),
	}
}

func strPtr(s string) *string {
	return &s
}

func TestHandleCreateUser(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	tests := []struct {
		name           string
		body           any
		mockErr        error
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "Success Registration",
			body: model.CreateUser{
				Name:        "Alice",
				Email:       "alice@example.com",
				Password:    "password123",
				Batch:       "2026",
				Description: "Art Lover",
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
		},
		{
			name:           "Invalid JSON Body",
			body:           "{invalid-json}",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
		{
			name: "Repo Create Error (Duplicate Email)",
			body: model.CreateUser{
				Name:     "Alice Dupe",
				Email:    "alice@example.com",
				Password: "password",
				Batch:    "2026",
			},
			mockErr:        fmt.Errorf("pq: duplicate key value violates unique constraint"),
			expectedStatus: http.StatusInternalServerError,
			expectSuccess:  false,
		},
		{
			name: "Empty Required Fields",
			body: model.CreateUser{
				Name:     "",
				Email:    "",
				Password: "",
				Batch:    "",
			},
			expectedStatus: http.StatusCreated, // Handler does not validate empty fields
			expectSuccess:  true,
		},
		{
			name:           "Empty Body",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.createErr = tc.mockErr
			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/users", &buf)
			rr := httptest.NewRecorder()

			h.HandleCreateUser(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			var response struct {
				Success bool
				Data    any
			}
			if err := json.NewDecoder(rr.Body).Decode(&response); err == nil {
				if response.Success != tc.expectSuccess {
					t.Errorf("expected Success to be %v, got %v", tc.expectSuccess, response.Success)
				}
			}
		})
	}
}

func TestHandleLogin(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	// Hash password and seed mock DB
	pwd := "mypassword"
	hashed, _ := utlis.HashPassword(pwd)
	userUUID := uuid.New()
	userSeed := database.User{
		ID:       userUUID,
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: hashed,
		Status:   database.AccountStatusApproved,
	}
	repo.users[userUUID] = userSeed
	repo.emails[userSeed.Email] = userSeed

	tests := []struct {
		name           string
		body           any
		expectedStatus int
		expectCookie   bool
	}{
		{
			name: "Successful Login",
			body: model.AuthenticateUser{
				Email:    "alice@example.com",
				Password: pwd,
			},
			expectedStatus: http.StatusOK,
			expectCookie:   true,
		},
		{
			name: "Non-existent user",
			body: model.AuthenticateUser{
				Email:    "unknown@example.com",
				Password: pwd,
			},
			expectedStatus: http.StatusNotFound,
			expectCookie:   false,
		},
		{
			name: "Incorrect Password",
			body: model.AuthenticateUser{
				Email:    "alice@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectCookie:   false,
		},
		{
			name:           "Invalid JSON Body",
			body:           "{bad-json}",
			expectedStatus: http.StatusBadRequest,
			expectCookie:   false,
		},
		{
			name: "Empty Email",
			body: model.AuthenticateUser{
				Email:    "",
				Password: pwd,
			},
			expectedStatus: http.StatusNotFound,
			expectCookie:   false,
		},
		{
			name: "Empty Password",
			body: model.AuthenticateUser{
				Email:    "alice@example.com",
				Password: "",
			},
			expectedStatus: http.StatusUnauthorized,
			expectCookie:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", &buf)
			rr := httptest.NewRecorder()

			h.HandleLogin(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}

			cookies := rr.Result().Cookies()
			foundCookie := false
			for _, cookie := range cookies {
				if cookie.Name == "authToken" {
					foundCookie = true
				}
			}
			if tc.expectCookie != foundCookie {
				t.Errorf("expected cookie presence: %v, got %v", tc.expectCookie, foundCookie)
			}
		})
	}
}

func TestHandleLogOut(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rr := httptest.NewRecorder()

	h.HandleLogOut(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	cookies := rr.Result().Cookies()
	foundClearedCookie := false
	for _, cookie := range cookies {
		if cookie.Name == "authToken" && cookie.MaxAge == -1 {
			foundClearedCookie = true
		}
	}
	if !foundClearedCookie {
		t.Errorf("expected authToken cookie to be cleared")
	}
}

func TestHandleUpdateUserProfile(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	userUUID := uuid.New()
	userSeed := database.User{
		ID:     userUUID,
		Name:   "Alice",
		Email:  "alice@example.com",
		Role:   database.UserRoleUser,
		Status: database.AccountStatusApproved,
	}
	repo.users[userUUID] = userSeed

	tests := []struct {
		name           string
		userIDParam    string
		authCtxUser    *middleware.User // nil means unauthenticated
		body           any
		mockErr        error
		expectedStatus int
	}{
		{
			name:        "Success Profile Update (Self)",
			userIDParam: userUUID.String(),
			authCtxUser: &middleware.User{
				ID:   userUUID,
				Role: database.UserRoleUser,
			},
			body: model.PatchUserProfileRequest{
				Name:  strPtr("Alice Updated"),
				Email: strPtr("alice.updated@example.com"),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Unauthorized access (Other User)",
			userIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID:   userUUID,
				Role: database.UserRoleUser,
			},
			body: model.PatchUserProfileRequest{
				Name: strPtr("Hacked Name"),
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "Admin Can Update Anyone",
			userIDParam: uuid.New().String(),
			authCtxUser: &middleware.User{
				ID:   userUUID,
				Role: database.UserRoleAdmin,
			},
			body: model.PatchUserProfileRequest{
				Name: strPtr("Admin Edit"),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthenticated Request",
			userIDParam:    userUUID.String(),
			authCtxUser:    nil,
			body:           model.PatchUserProfileRequest{Name: strPtr("Sneaky")},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "Invalid UUID Param",
			userIDParam: "not-a-uuid",
			authCtxUser: &middleware.User{
				ID:   userUUID,
				Role: database.UserRoleUser,
			},
			body:           model.PatchUserProfileRequest{Name: strPtr("X")},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Invalid JSON Body",
			userIDParam: userUUID.String(),
			authCtxUser: &middleware.User{
				ID:   userUUID,
				Role: database.UserRoleUser,
			},
			body:           "{bad-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Repo Error on Update",
			userIDParam: userUUID.String(),
			authCtxUser: &middleware.User{
				ID:   userUUID,
				Role: database.UserRoleUser,
			},
			body:           model.PatchUserProfileRequest{Name: strPtr("Fail")},
			mockErr:        fmt.Errorf("db connection lost"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Seed mock for arbitrary updates if target is not userUUID
			targetUUID, _ := uuid.Parse(tc.userIDParam)
			if targetUUID != userUUID {
				repo.users[targetUUID] = database.User{
					ID: targetUUID,
				}
			}

			repo.updateErr = tc.mockErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/auth/users/"+tc.userIDParam, &buf)

			// Add auth ctx only if authenticated
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			// Add chi URL parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.userIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleUpdateUserProfile(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandlePasswordChange(t *testing.T) {
	repo := newMockUserRepo()
	h := &Handler{Repo: repo}

	pwd := "original_password"
	hashed, _ := utlis.HashPassword(pwd)
	userUUID := uuid.New()
	userSeed := database.User{
		ID:       userUUID,
		Email:    "alice@example.com",
		Password: hashed,
	}
	repo.users[userUUID] = userSeed
	repo.emails[userSeed.Email] = userSeed

	tests := []struct {
		name           string
		authCtxUser    *middleware.User // nil means unauthenticated
		body           any
		mockPasswordErr error
		expectedStatus int
	}{
		{
			name: "Success Password Change",
			authCtxUser: &middleware.User{
				ID:    userUUID,
				Email: "alice@example.com",
			},
			body: model.PatchUserPassword{
				OldPassword: pwd,
				Password:    "new_password_123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Wrong Old Password",
			authCtxUser: &middleware.User{
				ID:    userUUID,
				Email: "alice@example.com",
			},
			body: model.PatchUserPassword{
				OldPassword: "wrong_old_pwd",
				Password:    "new_password_123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unauthenticated Request",
			authCtxUser:    nil,
			body:           model.PatchUserPassword{OldPassword: pwd, Password: "x"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid JSON Body",
			authCtxUser: &middleware.User{
				ID:    userUUID,
				Email: "alice@example.com",
			},
			body:           "{not-json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Repo Save Failure",
			authCtxUser: &middleware.User{
				ID:    userUUID,
				Email: "alice@example.com",
			},
			body: model.PatchUserPassword{
				OldPassword: pwd,
				Password:    "new_password_456",
			},
			mockPasswordErr: fmt.Errorf("db write error"),
			expectedStatus:  http.StatusInternalServerError,
		},
		{
			name: "User Email Not Found in DB",
			authCtxUser: &middleware.User{
				ID:    userUUID,
				Email: "ghost@example.com",
			},
			body: model.PatchUserPassword{
				OldPassword: pwd,
				Password:    "new_password_789",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Re-seed to ensure consistent state after "Success" test mutates the password
			repo.users[userUUID] = userSeed
			repo.emails[userSeed.Email] = userSeed
			repo.passwordErr = tc.mockPasswordErr

			var buf bytes.Buffer
			if s, ok := tc.body.(string); ok {
				buf.WriteString(s)
			} else {
				json.NewEncoder(&buf).Encode(tc.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/auth/users/password", &buf)
			if tc.authCtxUser != nil {
				ctx := middleware.WithUser(req.Context(), *tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			h.HandlePasswordChange(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
