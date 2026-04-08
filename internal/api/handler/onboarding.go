package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/domain"
)

// OnboardingHandler handles onboarding status endpoints.
type OnboardingHandler struct {
	vaultStore  VaultStore
	deviceStore DeviceStore
}

// NewOnboardingHandler creates an onboarding handler.
func NewOnboardingHandler(vaultStore VaultStore, deviceStore DeviceStore) *OnboardingHandler {
	return &OnboardingHandler{vaultStore: vaultStore, deviceStore: deviceStore}
}

// GetStatus returns the onboarding progress computed from existing data.
func (h *OnboardingHandler) GetStatus(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	// Compute onboarding status from existing data
	vaults, _ := h.vaultStore.ListVaults(claims.UserID)
	devices, _ := h.deviceStore.ListDevices(claims.UserID)

	vaultCount := len(vaults)
	deviceCount := len(devices)

	cliInstalled := deviceCount > 0 || vaultCount > 0
	firstPush := vaultCount > 0
	secondDevice := deviceCount >= 2
	completed := cliInstalled && firstPush

	return response.OK(c, http.StatusOK, map[string]any{
		"cli_installed":  cliInstalled,
		"first_push":     firstPush,
		"second_device":  secondDevice,
		"completed":      completed,
		"dismissed":      false, // TODO: persist dismiss state when onboarding_progress table is populated
	})
}

// Dismiss marks onboarding as dismissed for the user.
func (h *OnboardingHandler) Dismiss(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	// TODO: persist to onboarding_progress table when needed
	return response.OK(c, http.StatusOK, map[string]string{"message": "onboarding dismissed"})
}
