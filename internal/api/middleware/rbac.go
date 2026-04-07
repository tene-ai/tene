package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/response"
)

// TeamMemberChecker verifies team membership and environment permissions.
type TeamMemberChecker interface {
	GetEnvPermissions(teamID, userID string) ([]string, error)
	IsMember(teamID, userID string) bool
}

// RequireTeamMember returns middleware that ensures the user is a team member.
func RequireTeamMember(checker TeamMemberChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaims(c)
			if claims == nil {
				return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			}

			teamID := c.Param("id")
			if teamID == "" {
				return next(c)
			}

			if !checker.IsMember(teamID, claims.UserID) {
				return response.ErrMsg(c, http.StatusForbidden, "NOT_TEAM_MEMBER", "not a team member")
			}

			return next(c)
		}
	}
}

// RequireEnvAccess returns middleware that checks if the user has permission for the requested environment.
// The environment is read from query param "env" or request body field "environment".
func RequireEnvAccess(checker TeamMemberChecker) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := GetClaims(c)
			if claims == nil {
				return response.ErrMsg(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			}

			teamID := c.Param("id")
			if teamID == "" {
				return next(c) // no team context, skip env check
			}

			// Determine requested environment from query param only
			// (body is NOT consumed here to avoid breaking downstream handlers)
			env := c.QueryParam("env")
			if env == "" {
				env = "dev" // default environment
			}

			// Validate environment name
			if !isValidEnvName(env) {
				return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "invalid environment name")
			}

			// Check permission
			perms, err := checker.GetEnvPermissions(teamID, claims.UserID)
			if err != nil {
				return response.ErrMsg(c, http.StatusForbidden, "NOT_TEAM_MEMBER", "not a team member")
			}

			if !containsEnv(perms, env) {
				return response.ErrMsg(c, http.StatusForbidden, "ENV_ACCESS_DENIED",
					"no access to environment '"+env+"'. Contact your team admin.")
			}

			// Store env in context for handlers
			c.Set("env", env)
			return next(c)
		}
	}
}

var validEnvNames = map[string]bool{
	"dev": true, "staging": true, "prod": true, "default": true, "test": true,
}

func isValidEnvName(env string) bool {
	return validEnvNames[env]
}

func containsEnv(perms []string, env string) bool {
	for _, p := range perms {
		if p == env || p == "*" {
			return true
		}
	}
	return false
}
