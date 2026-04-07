package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/domain"
)

// DeviceStore defines database operations for devices.
type DeviceStore interface {
	RegisterDevice(d *domain.Device) error
	ListDevices(userID string) ([]domain.Device, error)
	DeleteDevice(id, userID string) error
}

// MemDeviceStore is an in-memory DeviceStore for development.
type MemDeviceStore struct {
	mu      sync.RWMutex
	devices map[string]*domain.Device
}

// NewMemDeviceStore creates an in-memory device store.
func NewMemDeviceStore() *MemDeviceStore {
	return &MemDeviceStore{devices: make(map[string]*domain.Device)}
}

// RegisterDevice stores a new device record.
func (s *MemDeviceStore) RegisterDevice(d *domain.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d.ID = uuid.New().String()
	d.CreatedAt = time.Now().UTC()
	d.LastSeenAt = d.CreatedAt
	s.devices[d.ID] = d
	return nil
}

// ListDevices returns all devices for a user.
func (s *MemDeviceStore) ListDevices(userID string) ([]domain.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []domain.Device
	for _, d := range s.devices {
		if d.UserID == userID {
			result = append(result, *d)
		}
	}
	return result, nil
}

// DeleteDevice removes a device by ID and owner.
func (s *MemDeviceStore) DeleteDevice(id, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.devices[id]
	if !ok || d.UserID != userID {
		return domain.ErrNotFound
	}
	delete(s.devices, id)
	return nil
}

// DeviceHandler handles device registration and management.
type DeviceHandler struct {
	store DeviceStore
}

// NewDeviceHandler creates a device handler.
func NewDeviceHandler(store DeviceStore) *DeviceHandler {
	return &DeviceHandler{store: store}
}

// Register creates a new device record.
func (h *DeviceHandler) Register(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	var req struct {
		DeviceName      string `json:"device_name"`
		X25519PublicKey []byte `json:"x25519_public_key"`
	}
	if err := c.Bind(&req); err != nil || req.DeviceName == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "device_name required")
	}
	if len(req.X25519PublicKey) > 0 && len(req.X25519PublicKey) != 32 {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "x25519_public_key must be 32 bytes")
	}

	device := &domain.Device{
		UserID:          claims.UserID,
		DeviceName:      req.DeviceName,
		X25519PublicKey: req.X25519PublicKey,
	}
	if err := h.store.RegisterDevice(device); err != nil {
		return response.Err(c, err)
	}

	return response.OK(c, http.StatusCreated, device)
}

// List returns all devices for the authenticated user.
func (h *DeviceHandler) List(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	devices, err := h.store.ListDevices(claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}
	return response.OK(c, http.StatusOK, devices)
}

// Delete removes a device.
func (h *DeviceHandler) Delete(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	deviceID := c.Param("id")
	if err := h.store.DeleteDevice(deviceID, claims.UserID); err != nil {
		return response.Err(c, err)
	}
	return response.OK(c, http.StatusOK, map[string]string{"message": "device removed"})
}
