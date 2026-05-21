package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Blue-Onion/ArtmeisterBackend/handler/admin"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/art"
	artmetadata "github.com/Blue-Onion/ArtmeisterBackend/handler/artMetaData"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/event"
	"github.com/Blue-Onion/ArtmeisterBackend/handler/user"
	"github.com/Blue-Onion/ArtmeisterBackend/internal/database"
	"github.com/Blue-Onion/ArtmeisterBackend/middleware"
	"github.com/Blue-Onion/ArtmeisterBackend/model"
	"github.com/Blue-Onion/ArtmeisterBackend/utlis"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) (*sql.DB, *database.Queries) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://adityasinghrawat@localhost:5432/artmeister_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	return db, database.New(db)
}

func clearDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("TRUNCATE TABLE users, art, events, event_attendees RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatalf("failed to clear database: %v", err)
	}
}

func setupTestRouter(db *sql.DB, query *database.Queries) *chi.Mux {
	userHandler := &user.Handler{Repo: query}
	middlewareHandler := &middleware.Handler{Repo: query}
	artHandler := &art.Handler{Repo: query}
	artMetadataHandler := &artmetadata.Handler{Repo: query}
	eventHandler := &event.EventHandler{Repo: query}
	eventAttendeeHandler := &event.EventAttendeeHandler{Repo: query}

	router := chi.NewRouter()

	// Mount routes
	router.Mount("/auth", user.UserRouter(userHandler, middlewareHandler))
	router.Mount("/art", art.ArtRouter(artHandler, artMetadataHandler, middlewareHandler))
	router.Mount("/event", event.EventRouter(eventHandler, eventAttendeeHandler, middlewareHandler))
	router.Mount("/admin", admin.AdminRoute(userHandler, artHandler, middlewareHandler))

	return router
}

// Helper: Decode custom envelope from RespondWithJson
func decodeEnvelope(body []byte, target any) error {
	var response struct {
		Success bool
		Data    json.RawMessage
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	return json.Unmarshal(response.Data, target)
}

func TestIntegrationUserFlow(t *testing.T) {
	db, query := getTestDB(t)
	defer db.Close()
	clearDB(t, db)

	router := setupTestRouter(db, query)

	// 1. Create User
	userCreds := model.CreateUser{
		Name:     "Integration User",
		Email:    "integration@example.com",
		Password: "strongpassword123",
		Batch:    "2026",
	}
	bodyBytes, _ := json.Marshal(userCreds)
	req := httptest.NewRequest(http.MethodPost, "/auth/users", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var createdUser database.CreateUserRow
	if err := decodeEnvelope(w.Body.Bytes(), &createdUser); err != nil {
		t.Fatalf("failed to decode user response: %v", err)
	}

	if createdUser.Email != userCreds.Email {
		t.Errorf("expected email %s, got %s", userCreds.Email, createdUser.Email)
	}

	// 2. Login
	loginBody, _ := json.Marshal(model.AuthenticateUser{
		Email:    userCreds.Email,
		Password: userCreds.Password,
	})
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: expected 200, got %d", w.Code)
	}

	// Extract Cookie
	cookies := w.Result().Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "authToken" {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatal("authToken cookie not found in login response")
	}

	// 3. Update Profile (Authorized)
	updatePayload := model.PatchUserProfileRequest{
		Name: strPtr("New Integration Name"),
	}
	updateBytes, _ := json.Marshal(updatePayload)
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/auth/users/%s", createdUser.ID), bytes.NewReader(updateBytes))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for profile update, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 4. Update Profile (Unauthorized - other user)
	otherUserUUID := uuid.New().String()
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/auth/users/%s", otherUserUUID), bytes.NewReader(updateBytes))
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for updating other user, got %d", w.Code)
	}

	// 5. Logout
	req = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(authCookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("logout failed: expected 200, got %d", w.Code)
	}
}

func TestIntegrationArtFlow(t *testing.T) {
	db, query := getTestDB(t)
	defer db.Close()
	clearDB(t, db)

	router := setupTestRouter(db, query)

	// Create User and Login to get Auth Cookie
	pwd := "password"
	hashedPwd, _ := utlis.HashPassword(pwd)
	hashed, _ := query.CreateUser(context.Background(), database.CreateUserParams{
		Name:     "Artist A",
		Email:    "artist@example.com",
		Password: hashedPwd,
		Batch:    "2026",
		Status:   database.AccountStatusApproved,
		Role:     database.UserRoleUser,
	})
	loginBody, _ := json.Marshal(model.AuthenticateUser{
		Email:    "artist@example.com",
		Password: pwd,
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected cookies from login, got none. Body: %s", w.Body.String())
	}
	cookie := cookies[0]

	// 1. Create Art Work (Using multipart form)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("image", "canvas.png")
	part.Write([]byte("fake-png-data"))
	writer.WriteField("name", "Starry Night")
	writer.WriteField("description", "A beautiful painting")
	writer.Close()

	req = httptest.NewRequest(http.MethodPost, "/art", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var createdArt database.Art
	if err := decodeEnvelope(w.Body.Bytes(), &createdArt); err != nil {
		t.Fatalf("failed to decode art: %v", err)
	}

	if createdArt.Name != "Starry Night" {
		t.Errorf("expected art name Starry Night, got %s", createdArt.Name)
	}
	if createdArt.Status != database.ArtStatusPending {
		t.Errorf("expected art status pending, got %v", createdArt.Status)
	}

	// Cleanup saved local file (uploads/userID/art/artID.png)
	defer os.RemoveAll(fmt.Sprintf("uploads/%s", hashed.ID.String()))

	// 2. Admin Approve Art
	// Create Admin and Login
	adminPwd := "adminpass"
	adminHashedPwd, _ := utlis.HashPassword(adminPwd)
	adminUser, _ := query.CreateUser(context.Background(), database.CreateUserParams{
		Name:     "Admin User",
		Email:    "admin@example.com",
		Password: adminHashedPwd,
		Batch:    "Admin",
		Status:   database.AccountStatusApproved,
		Role:     database.UserRoleAdmin,
	})
	adminLoginBody, _ := json.Marshal(model.AuthenticateUser{
		Email:    adminUser.Email,
		Password: adminPwd,
	})
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(adminLoginBody))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	adminCookies := w.Result().Cookies()
	if len(adminCookies) == 0 {
		t.Fatalf("expected admin cookies, got none. Body: %s", w.Body.String())
	}
	adminCookie := adminCookies[0]

	// Approve artwork
	req = httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/admin/arts/%s/status?status=approved", createdArt.ID), nil)
	req.AddCookie(adminCookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 for art approval, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 3. Delete Art (Authorized)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/art/%s", createdArt.ID), nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("failed to delete art: expected 200, got %d", w.Code)
	}

	// Verify art is gone
	_, err := query.GetArtByID(context.Background(), createdArt.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected art to be deleted, got error: %v", err)
	}
}

func TestIntegrationEventConcurrency(t *testing.T) {
	db, query := getTestDB(t)
	defer db.Close()
	clearDB(t, db)

	router := setupTestRouter(db, query)

	// Create Event
	eventID := uuid.New()
	_, err := query.CreateEvent(context.Background(), database.CreateEventParams{
		ID:        eventID,
		Name:      "Tech Conference 2026",
		EventDate: time.Now().Add(24 * time.Hour),
		Status:    database.ModeOfConductOffline,
	})
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	// Create 10 Users and their cookies
	const numUsers = 10
	var cookies []*http.Cookie
	for i := 0; i < numUsers; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		pwd := "password123"
		hashedPwd, _ := utlis.HashPassword(pwd)
		_, err := query.CreateUser(context.Background(), database.CreateUserParams{
			Name:     fmt.Sprintf("User %d", i),
			Email:    email,
			Password: hashedPwd,
			Batch:    "2026",
			Status:   database.AccountStatusApproved,
			Role:     database.UserRoleUser,
		})
		if err != nil {
			t.Fatalf("failed to create user %d: %v", i, err)
		}

		loginBody, _ := json.Marshal(model.AuthenticateUser{
			Email:    email,
			Password: pwd,
		})
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		resCookies := w.Result().Cookies()
		if len(resCookies) == 0 {
			t.Fatalf("failed to login user %d, body: %s", i, w.Body.String())
		}
		cookies = append(cookies, resCookies[0])
	}

	// 10 Goroutines concurrently attempting to join the event
	var wg sync.WaitGroup
	wg.Add(numUsers)
	results := make([]int, numUsers)

	for i := 0; i < numUsers; i++ {
		go func(index int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/event/%s/join", eventID), nil)
			req.AddCookie(cookies[index])
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results[index] = w.Code
		}(i)
	}
	wg.Wait()

	// Verify all returned 200 OK
	for i, code := range results {
		if code != http.StatusOK {
			t.Errorf("user %d failed to join, got status: %d", i, code)
		}
	}

	// Check count in database
	count, err := query.CountEventAttendees(context.Background(), eventID)
	if err != nil {
		t.Fatalf("failed to count attendees: %v", err)
	}
	if count != int32(numUsers) {
		t.Errorf("expected %d attendees, got %d", numUsers, count)
	}
}

func TestIntegrationRateLimiter(t *testing.T) {
	db, _ := getTestDB(t)
	defer db.Close()

	// Setup a local test router that uses the Rate Limit middleware
	router := chi.NewRouter()
	router.Use(middleware.MiddlewareRateLimit)
	router.Get("/test-limit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Make 6 fast consecutive requests (limit is 5)
	var lastStatus int
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test-limit", nil)
		req.RemoteAddr = "127.0.0.1:12345" // ensure consistent IP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		lastStatus = w.Code
	}

	if lastStatus != http.StatusTooManyRequests {
		t.Errorf("expected 6th request to be rate limited (429), got: %d", lastStatus)
	}
}

func strPtr(s string) *string {
	return &s
}
