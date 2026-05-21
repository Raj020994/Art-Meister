package event

import (
	"context"
	"database/sql"
	"errors"
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

type mockEventRepo struct {
	database.EventRepository
	database.EventAttendeesRepository
	events       map[uuid.UUID]database.Event
	attendees    map[uuid.UUID][]uuid.UUID // event_id -> list of user_ids
	createErr    error
	getErr       error
	attendeeErr  error
}

func (m *mockEventRepo) CreateEvent(ctx context.Context, arg database.CreateEventParams) (database.Event, error) {
	if m.createErr != nil {
		return database.Event{}, m.createErr
	}
	e := database.Event{
		ID:          arg.ID,
		Name:        arg.Name,
		Description: arg.Description,
		Venue:       arg.Venue,
		Image:       arg.Image,
		BannerImage: arg.BannerImage,
		EventDate:   arg.EventDate,
		Status:      arg.Status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.events[arg.ID] = e
	return e, nil
}

func (m *mockEventRepo) GetEventByID(ctx context.Context, id uuid.UUID) (database.Event, error) {
	if m.getErr != nil {
		return database.Event{}, m.getErr
	}
	e, ok := m.events[id]
	if !ok {
		return database.Event{}, sql.ErrNoRows
	}
	return e, nil
}

func (m *mockEventRepo) ListEvents(ctx context.Context) ([]database.Event, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	var res []database.Event
	for _, e := range m.events {
		res = append(res, e)
	}
	return res, nil
}

func (m *mockEventRepo) DeleteEvent(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	if m.getErr != nil {
		return uuid.UUID{}, m.getErr
	}
	if _, ok := m.events[id]; !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	delete(m.events, id)
	return id, nil
}

func (m *mockEventRepo) EnrollUserToEvent(ctx context.Context, arg database.EnrollUserToEventParams) (database.EventAttendee, error) {
	if m.attendeeErr != nil {
		return database.EventAttendee{}, m.attendeeErr
	}
	// Check if already joined
	list := m.attendees[arg.EventID]
	for _, uid := range list {
		if uid == arg.UserID {
			return database.EventAttendee{}, errors.New("duplicate key violation (already joined)")
		}
	}
	m.attendees[arg.EventID] = append(list, arg.UserID)
	return database.EventAttendee{
		ID:        arg.ID,
		EventID:   arg.EventID,
		UserID:    arg.UserID,
		JoinedAt:  time.Now(),
	}, nil
}

func (m *mockEventRepo) RemoveUserFromEvent(ctx context.Context, arg database.RemoveUserFromEventParams) (uuid.UUID, error) {
	if m.attendeeErr != nil {
		return uuid.UUID{}, m.attendeeErr
	}
	list, ok := m.attendees[arg.EventID]
	if !ok {
		return uuid.UUID{}, sql.ErrNoRows
	}
	found := false
	var updated []uuid.UUID
	for _, uid := range list {
		if uid == arg.UserID {
			found = true
		} else {
			updated = append(updated, uid)
		}
	}
	if !found {
		return uuid.UUID{}, sql.ErrNoRows
	}
	m.attendees[arg.EventID] = updated
	return arg.EventID, nil
}

func (m *mockEventRepo) ListEventAttendees(ctx context.Context, eventID uuid.UUID) ([]database.User, error) {
	if m.attendeeErr != nil {
		return nil, m.attendeeErr
	}
	list := m.attendees[eventID]
	var users []database.User
	for _, uid := range list {
		users = append(users, database.User{
			ID:   uid,
			Name: "Attendee User",
		})
	}
	return users, nil
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events:    make(map[uuid.UUID]database.Event),
		attendees: make(map[uuid.UUID][]uuid.UUID),
	}
}

func TestHandleGetEventById(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	eventUUID := uuid.New()
	repo.events[eventUUID] = database.Event{ID: eventUUID, Name: "Art Fest"}

	tests := []struct {
		name           string
		eventIDParam   string
		expectedStatus int
	}{
		{"Success", eventUUID.String(), http.StatusOK},
		{"Invalid UUID", "bad", http.StatusBadRequest},
		{"Not Found", uuid.New().String(), http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/event/"+tc.eventIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.HandleGetEventById(rr, req)
			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}

func TestHandleGetAllEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	repo.events[uuid.New()] = database.Event{Name: "E1"}

	req := httptest.NewRequest(http.MethodGet, "/event", nil)
	rr := httptest.NewRecorder()
	h.HandleGetAllEvent(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandleDeleteEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventHandler{Repo: repo}

	eventUUID := uuid.New()
	repo.events[eventUUID] = database.Event{
		ID:   eventUUID,
		Name: "Exhibition",
	}

	tests := []struct {
		name           string
		eventIDParam   string
		expectedStatus int
	}{
		{
			name:           "Success Delete",
			eventIDParam:   eventUUID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID Format",
			eventIDParam:   "not-a-uuid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent Event",
			eventIDParam:   uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.events[eventUUID] = database.Event{ID: eventUUID, Name: "Exhibition"}
			req := httptest.NewRequest(http.MethodDelete, "/event/"+tc.eventIDParam, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleDeleteEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}

func TestHandleJoinEvent(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventUUID := uuid.New()
	userUUID := uuid.New()

	tests := []struct {
		name           string
		eventIDParam   string
		authCtxUser    middleware.User
		authOk         bool
		mockErr        error
		expectedStatus int
	}{
		{
			name:         "Success Join",
			eventIDParam: eventUUID.String(),
			authCtxUser: middleware.User{
				ID: userUUID,
			},
			authOk:         true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthenticated Fail",
			eventIDParam:   eventUUID.String(),
			authOk:         false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:         "Double Join Fail",
			eventIDParam: eventUUID.String(),
			authCtxUser: middleware.User{
				ID: userUUID,
			},
			authOk:         true,
			mockErr:        errors.New("duplicate key"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:         "Invalid Event UUID",
			eventIDParam: "bad-uuid",
			authCtxUser:  middleware.User{ID: userUUID},
			authOk:       true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo.attendeeErr = tc.mockErr
			repo.attendees[eventUUID] = nil

			req := httptest.NewRequest(http.MethodPost, "/event/"+tc.eventIDParam+"/join", nil)
			if tc.authOk {
				ctx := middleware.WithUser(req.Context(), tc.authCtxUser)
				req = req.WithContext(ctx)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleJoinEvent(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleDeleteEventAttendee(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventUUID := uuid.New()
	userUUID := uuid.New()
	repo.attendees[eventUUID] = []uuid.UUID{userUUID}

	tests := []struct {
		name           string
		eventIDParam   string
		queryUserID    string
		expectedStatus int
	}{
		{
			name:           "Success Remove",
			eventIDParam:   eventUUID.String(),
			queryUserID:    userUUID.String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid User UUID",
			eventIDParam:   eventUUID.String(),
			queryUserID:    "invalid-user",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent Attendance",
			eventIDParam:   eventUUID.String(),
			queryUserID:    uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid Event UUID",
			eventIDParam:   "bad-event",
			queryUserID:    userUUID.String(),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "Success Remove" {
				repo.attendees[eventUUID] = []uuid.UUID{userUUID}
			}
			repo.attendeeErr = nil

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/event/%s/attendee?user_id=%s", tc.eventIDParam, tc.queryUserID), nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.HandleDeleteEventAttendee(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d: %s", tc.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleAllEventAttendee(t *testing.T) {
	repo := newMockEventRepo()
	h := &EventAttendeeHandler{Repo: repo}

	eventUUID := uuid.New()
	repo.attendees[eventUUID] = []uuid.UUID{uuid.New(), uuid.New()}

	tests := []struct {
		name           string
		eventIDParam   string
		expectedStatus int
	}{
		{"Success", eventUUID.String(), http.StatusOK},
		{"Invalid UUID", "bad", http.StatusBadRequest},
		{"Empty List", uuid.New().String(), http.StatusOK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/event/"+tc.eventIDParam+"/attendees", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.eventIDParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.HandleAllEventAttendee(rr, req)
			if rr.Code != tc.expectedStatus {
				t.Errorf("expected %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}
