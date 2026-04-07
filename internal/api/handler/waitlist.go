package handler

import (
	"log/slog"
	"net/http"
	"regexp"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/response"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// WaitlistStore defines storage for waitlist entries.
type WaitlistStore interface {
	AddToWaitlist(email, plan, source string) error
}

// MemWaitlistStore is an in-memory waitlist store for development.
type MemWaitlistStore struct {
	mu      sync.Mutex
	entries map[string]string // email -> plan
}

// NewMemWaitlistStore creates an in-memory waitlist store.
func NewMemWaitlistStore() *MemWaitlistStore {
	return &MemWaitlistStore{entries: make(map[string]string)}
}

// AddToWaitlist registers an email in the waitlist.
func (s *MemWaitlistStore) AddToWaitlist(email, plan, source string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[email] = plan
	return nil
}

// WaitlistHandler handles email waitlist registration.
type WaitlistHandler struct {
	store WaitlistStore
}

// NewWaitlistHandler creates a waitlist handler.
func NewWaitlistHandler(store WaitlistStore) *WaitlistHandler {
	return &WaitlistHandler{store: store}
}

// Register adds an email to the waitlist.
func (h *WaitlistHandler) Register(c echo.Context) error {
	var req struct {
		Email  string `json:"email"`
		Plan   string `json:"plan"`
		Source string `json:"source"`
	}
	if err := c.Bind(&req); err != nil {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
	}

	if !emailRegex.MatchString(req.Email) {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "invalid email address")
	}

	if req.Plan == "" {
		req.Plan = "pro"
	}
	if req.Source == "" {
		req.Source = "web"
	}

	if err := h.store.AddToWaitlist(req.Email, req.Plan, req.Source); err != nil {
		slog.Error("waitlist.register.failed", "email", req.Email, "error", err)
		return response.ErrMsg(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to register")
	}

	slog.Info("waitlist.registered", "email", req.Email, "plan", req.Plan, "source", req.Source)
	return response.OK(c, http.StatusCreated, map[string]string{
		"message": "registered",
		"email":   req.Email,
	})
}
