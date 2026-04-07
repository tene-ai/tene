package handler

import (
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/domain"
)

var validSlug = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// TeamStore defines database operations for teams.
type TeamStore interface {
	CreateTeam(t *domain.Team) error
	GetTeam(id string) (*domain.Team, error)
	ListTeamsByUser(userID string) ([]domain.Team, error)
	AddMember(m *domain.TeamMember) error
	RemoveMember(teamID, userID string) error
	UpdateMemberRole(teamID, userID, role string) error
	ListMembers(teamID string) ([]domain.TeamMember, error)
	IsMember(teamID, userID string) bool
	IsAdmin(teamID, userID string) bool
}

// MemTeamStore is an in-memory TeamStore for development.
type MemTeamStore struct {
	mu      sync.RWMutex
	teams   map[string]*domain.Team
	members map[string][]domain.TeamMember // teamID -> members
}

// NewMemTeamStore creates an in-memory team store.
func NewMemTeamStore() *MemTeamStore {
	return &MemTeamStore{
		teams:   make(map[string]*domain.Team),
		members: make(map[string][]domain.TeamMember),
	}
}

// CreateTeam stores a new team.
func (s *MemTeamStore) CreateTeam(t *domain.Team) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.teams {
		if existing.Slug == t.Slug {
			return domain.ErrProjectAlreadyExists
		}
	}
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now().UTC()
	s.teams[t.ID] = t
	return nil
}

// GetTeam retrieves a team by ID.
func (s *MemTeamStore) GetTeam(id string) (*domain.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.teams[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

// ListTeamsByUser returns all teams where the user is a member or owner.
func (s *MemTeamStore) ListTeamsByUser(userID string) ([]domain.Team, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []domain.Team
	for _, t := range s.teams {
		if t.OwnerID == userID {
			result = append(result, *t)
			continue
		}
		for _, m := range s.members[t.ID] {
			if m.UserID == userID {
				result = append(result, *t)
				break
			}
		}
	}
	return result, nil
}

// AddMember adds a user to a team. Returns error if already a member.
func (s *MemTeamStore) AddMember(m *domain.TeamMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.members[m.TeamID] {
		if existing.UserID == m.UserID {
			return domain.ErrProjectAlreadyExists // duplicate membership
		}
	}
	s.members[m.TeamID] = append(s.members[m.TeamID], *m)
	return nil
}

// RemoveMember removes a user from a team.
func (s *MemTeamStore) RemoveMember(teamID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	members := s.members[teamID]
	for i, m := range members {
		if m.UserID == userID {
			s.members[teamID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotTeamMember
}

// UpdateMemberRole changes a member's role.
func (s *MemTeamStore) UpdateMemberRole(teamID, userID, role string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, m := range s.members[teamID] {
		if m.UserID == userID {
			s.members[teamID][i].Role = role
			return nil
		}
	}
	return domain.ErrNotTeamMember
}

// ListMembers returns all members of a team.
func (s *MemTeamStore) ListMembers(teamID string) ([]domain.TeamMember, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.members[teamID], nil
}

// IsMember checks if a user is a member of a team.
func (s *MemTeamStore) IsMember(teamID, userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.teams[teamID]; ok && t.OwnerID == userID {
		return true
	}
	for _, m := range s.members[teamID] {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

// GetEnvPermissions returns the environment permissions for a user in a team.
func (s *MemTeamStore) GetEnvPermissions(teamID, userID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Owner has access to all environments
	if t, ok := s.teams[teamID]; ok && t.OwnerID == userID {
		return []string{"*"}, nil
	}
	for _, m := range s.members[teamID] {
		if m.UserID == userID {
			if len(m.EnvPermissions) > 0 {
				return m.EnvPermissions, nil
			}
			// Default: admin gets all, member gets dev only
			if m.Role == "admin" {
				return []string{"*"}, nil
			}
			return []string{"dev", "staging"}, nil
		}
	}
	return nil, domain.ErrNotTeamMember
}

// IsAdmin checks if a user is an admin or owner of a team.
func (s *MemTeamStore) IsAdmin(teamID, userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.teams[teamID]; ok && t.OwnerID == userID {
		return true
	}
	for _, m := range s.members[teamID] {
		if m.UserID == userID && m.Role == "admin" {
			return true
		}
	}
	return false
}

// TeamHandler handles team CRUD and member management.
type TeamHandler struct {
	store TeamStore
}

// NewTeamHandler creates a team handler.
func NewTeamHandler(store TeamStore) *TeamHandler {
	return &TeamHandler{store: store}
}

// Create creates a new team. Requires Pro plan.
func (h *TeamHandler) Create(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.Bind(&req); err != nil || req.Name == "" || req.Slug == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "name and slug required")
	}
	if !validSlug.MatchString(req.Slug) {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "slug must be lowercase alphanumeric with dashes")
	}

	team := &domain.Team{
		Name:    req.Name,
		Slug:    req.Slug,
		OwnerID: claims.UserID,
	}
	if err := h.store.CreateTeam(team); err != nil {
		return response.Err(c, err)
	}

	// Auto-add owner as admin member
	if err := h.store.AddMember(&domain.TeamMember{
		TeamID:   team.ID,
		UserID:   claims.UserID,
		Role:     "admin",
		JoinedAt: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		slog.Error("team.create.add_owner_failed", "team_id", team.ID, "error", err)
		return response.ErrMsg(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to add owner to team")
	}

	slog.Info("team.created", "team_id", team.ID, "owner", claims.UserID)
	return response.OK(c, http.StatusCreated, team)
}

// List returns teams the user belongs to.
func (h *TeamHandler) List(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	teams, err := h.store.ListTeamsByUser(claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}
	return response.OK(c, http.StatusOK, teams)
}

// Invite adds a member to a team with a wrapped project key. Both inviter and invitee must be Pro.
func (h *TeamHandler) Invite(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	teamID := c.Param("id")
	if !h.store.IsAdmin(teamID, claims.UserID) {
		return response.Err(c, domain.ErrForbidden)
	}

	var req struct {
		UserID            string `json:"user_id"`
		Role              string `json:"role"`
		WrappedProjectKey []byte `json:"wrapped_project_key"`
	}
	if err := c.Bind(&req); err != nil || req.UserID == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "user_id required")
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "admin" && req.Role != "member" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "role must be admin or member")
	}

	member := &domain.TeamMember{
		TeamID:            teamID,
		UserID:            req.UserID,
		Role:              req.Role,
		WrappedProjectKey: req.WrappedProjectKey,
		JoinedAt:          time.Now().UTC().Format(time.RFC3339),
	}
	if err := h.store.AddMember(member); err != nil {
		return response.Err(c, err)
	}

	slog.Info("team.invite", "team_id", teamID, "user_id", req.UserID, "role", req.Role)
	return response.OK(c, http.StatusCreated, member)
}

// RemoveMember removes a member from a team (triggers key rotation).
func (h *TeamHandler) RemoveMember(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	teamID := c.Param("id")
	uid := c.Param("uid")

	if !h.store.IsAdmin(teamID, claims.UserID) {
		return response.Err(c, domain.ErrForbidden)
	}

	if err := h.store.RemoveMember(teamID, uid); err != nil {
		return response.Err(c, err)
	}

	slog.Info("team.member_removed", "team_id", teamID, "user_id", uid, "by", claims.UserID)
	return response.OK(c, http.StatusOK, map[string]any{
		"message":          "member removed",
		"key_rotation":     true,
		"rotation_pending": true,
	})
}

// UpdateRole changes a member's role.
func (h *TeamHandler) UpdateRole(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	teamID := c.Param("id")
	uid := c.Param("uid")

	if !h.store.IsAdmin(teamID, claims.UserID) {
		return response.Err(c, domain.ErrForbidden)
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.Bind(&req); err != nil || req.Role == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "role required")
	}
	if req.Role != "admin" && req.Role != "member" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "role must be admin or member")
	}

	if err := h.store.UpdateMemberRole(teamID, uid, req.Role); err != nil {
		return response.Err(c, err)
	}

	return response.OK(c, http.StatusOK, map[string]string{"message": "role updated"})
}
