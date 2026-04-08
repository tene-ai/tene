package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tomo-kay/tene/internal/api/middleware"
	"github.com/tomo-kay/tene/internal/api/response"
	"github.com/tomo-kay/tene/internal/api/storage"
	"github.com/tomo-kay/tene/internal/domain"
)

var validProjectName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,127}$`)

// VaultStore defines the database operations for vaults.
type VaultStore interface {
	CreateVault(v *domain.Vault) error
	GetVault(id, userID string) (*domain.Vault, error)
	GetVaultByProject(userID, projectName string) (*domain.Vault, error)
	ListVaults(userID string) ([]domain.Vault, error)
	UpdateVaultVersion(id string, currentVersion int64, hash []byte, size int64, secretCount int) (int64, error)
	DeleteVault(id, userID string) error
	CreateAuditLog(log *domain.AuditLog) error
}

// VaultKeyMetadataStore defines operations for vault key metadata.
type VaultKeyMetadataStore interface {
	GetKeyMetadata(ctx context.Context, vaultID, env string) ([]domain.VaultKeyMeta, error)
	GetEnvironments(ctx context.Context, vaultID string) ([]string, error)
	UpsertKeyMetadata(ctx context.Context, vaultID string, metadata domain.VaultMetadataPayload) error
}

// VaultHandler handles vault CRUD and push/pull endpoints.
type VaultHandler struct {
	store    VaultStore
	keyStore VaultKeyMetadataStore
	storage  *storage.S3Client
}

// NewVaultHandler creates a vault handler.
func NewVaultHandler(store VaultStore, keyStore VaultKeyMetadataStore, s3 *storage.S3Client) *VaultHandler {
	return &VaultHandler{store: store, keyStore: keyStore, storage: s3}
}

// Get returns a single vault by ID for the authenticated user.
func (h *VaultHandler) Get(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	vaultID := c.Param("id")
	vault, err := h.store.GetVault(vaultID, claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}

	return response.OK(c, http.StatusOK, vault)
}

// List returns all vaults for the authenticated user.
func (h *VaultHandler) List(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	vaults, err := h.store.ListVaults(claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}

	return response.OK(c, http.StatusOK, vaults)
}

// Create creates a new vault record.
func (h *VaultHandler) Create(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	// Pro plan required for cloud sync
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	var req struct {
		ProjectName string `json:"project_name"`
	}
	if err := c.Bind(&req); err != nil || req.ProjectName == "" {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "project_name required")
	}

	// Validate project name to prevent path traversal in S3 keys
	if !validProjectName.MatchString(req.ProjectName) {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "project_name must be 1-128 alphanumeric, dash, or underscore characters")
	}

	vault := &domain.Vault{
		UserID:      claims.UserID,
		ProjectName: req.ProjectName,
		S3Key:       "vaults/" + claims.UserID + "/" + req.ProjectName + "/vault.enc",
		Version:     0,
		VaultHash:   make([]byte, 32),
	}

	if err := h.store.CreateVault(vault); err != nil {
		return response.Err(c, err)
	}

	h.audit(claims.UserID, vault.ID, "vault.create", vault.ProjectName, c.RealIP())
	return response.OK(c, http.StatusCreated, vault)
}

// Push handles encrypted vault blob upload with optimistic locking.
//
//	POST /api/v1/vaults/:id/push
//	Headers: If-Match (expected version)
//	Body: encrypted envelope blob (binary)
func (h *VaultHandler) Push(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	vaultID := c.Param("id")

	// Get current vault to verify ownership
	vault, err := h.store.GetVault(vaultID, claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}

	// Optimistic locking: check If-Match header
	ifMatch := c.Request().Header.Get("If-Match")
	if ifMatch != "" {
		expectedVersion, parseErr := strconv.ParseInt(ifMatch, 10, 64)
		if parseErr != nil {
			return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "invalid If-Match header")
		}
		if vault.Version > expectedVersion {
			return response.ErrMsg(c, http.StatusConflict, "VERSION_CONFLICT",
				"server version "+strconv.FormatInt(vault.Version, 10)+" > expected "+ifMatch+". Pull first or use --force.")
		}
	}

	// Read encrypted blob — supports both raw binary and multipart/form-data
	var body []byte
	var metadata *domain.VaultMetadataPayload

	contentType := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// New format: blob + optional metadata
		blobFile, blobErr := c.FormFile("blob")
		if blobErr != nil {
			return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "missing 'blob' part in multipart")
		}
		f, openErr := blobFile.Open()
		if openErr != nil {
			return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "failed to open blob")
		}
		defer func() { _ = f.Close() }()
		body, err = io.ReadAll(io.LimitReader(f, 50<<20))
		if err != nil {
			return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "failed to read blob")
		}

		// Parse optional metadata
		if metaJSON := c.FormValue("metadata"); metaJSON != "" {
			var m domain.VaultMetadataPayload
			if json.Unmarshal([]byte(metaJSON), &m) == nil {
				metadata = &m
			}
		}
	} else {
		// Legacy format: raw binary body
		body, err = io.ReadAll(io.LimitReader(c.Request().Body, 50<<20))
		if err != nil {
			return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "failed to read body")
		}
	}

	if len(body) == 0 {
		return response.ErrMsg(c, http.StatusBadRequest, "BAD_REQUEST", "empty body")
	}

	// Compute checksum
	checksum := sha256.Sum256(body)

	// Upload to S3
	if err := h.storage.Upload(c.Request().Context(), vault.S3Key, body); err != nil {
		slog.Error("vault.push.s3_upload_failed", "vault_id", vaultID, "error", err)
		return response.ErrMsg(c, http.StatusInternalServerError, "UPLOAD_FAILED", "failed to upload vault")
	}

	// Determine secret count from metadata or fallback
	secretCount := 0
	if metadata != nil {
		secretCount = metadata.SecretCount
	}

	// Update vault version with optimistic lock
	newVersion, err := h.store.UpdateVaultVersion(vaultID, vault.Version, checksum[:], int64(len(body)), secretCount)
	if err != nil {
		return response.Err(c, err)
	}

	// Store key metadata if provided
	if metadata != nil && h.keyStore != nil {
		if upsertErr := h.keyStore.UpsertKeyMetadata(c.Request().Context(), vaultID, *metadata); upsertErr != nil {
			slog.Error("vault.push.metadata_upsert_failed", "vault_id", vaultID, "error", upsertErr)
		}
	}

	h.audit(claims.UserID, vaultID, "vault.push", vault.ProjectName, c.RealIP())

	return response.OK(c, http.StatusOK, map[string]any{
		"vault_id":      vaultID,
		"vault_version": newVersion,
		"vault_hash":    hex.EncodeToString(checksum[:]),
		"size":          len(body),
		"pushed_at":     time.Now().UTC().Format(time.RFC3339),
	})
}

// Pull returns a presigned S3 download URL for the vault blob.
//
//	GET /api/v1/vaults/:id/pull
func (h *VaultHandler) Pull(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	vaultID := c.Param("id")

	vault, err := h.store.GetVault(vaultID, claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}

	if vault.Version == 0 {
		return response.ErrMsg(c, http.StatusNotFound, "VAULT_EMPTY", "vault has never been pushed")
	}

	// Generate 5-minute presigned URL (short-lived to limit exposure)
	url, err := h.storage.GeneratePresignedURL(c.Request().Context(), vault.S3Key, 5*time.Minute)
	if err != nil {
		slog.Error("vault.pull.presign_failed", "vault_id", vaultID, "error", err)
		return response.ErrMsg(c, http.StatusInternalServerError, "PRESIGN_FAILED", "failed to generate download URL")
	}

	h.audit(claims.UserID, vaultID, "vault.pull", vault.ProjectName, c.RealIP())

	c.Response().Header().Set("Cache-Control", "no-store")
	return response.OK(c, http.StatusOK, map[string]any{
		"vault_id":      vaultID,
		"vault_version": vault.Version,
		"vault_hash":    hex.EncodeToString(vault.VaultHash),
		"download_url":  url,
		"expires_in":    300, // 5 minutes
	})
}

// Delete removes a vault and its S3 blob.
func (h *VaultHandler) Delete(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}

	vaultID := c.Param("id")

	vault, err := h.store.GetVault(vaultID, claims.UserID)
	if err != nil {
		return response.Err(c, err)
	}

	// Delete S3 blob (best effort, log failures)
	if h.storage != nil {
		if err := h.storage.Delete(c.Request().Context(), vault.S3Key); err != nil {
			slog.Error("vault.delete.s3_failed", "vault_id", vaultID, "s3_key", vault.S3Key, "error", err)
		}
	}

	if err := h.store.DeleteVault(vaultID, claims.UserID); err != nil {
		return response.Err(c, err)
	}

	h.audit(claims.UserID, vaultID, "vault.delete", vault.ProjectName, c.RealIP())
	return response.OK(c, http.StatusOK, map[string]string{"message": "vault deleted"})
}

// ListKeys returns secret key metadata for a vault (names, versions, timestamps — no values).
func (h *VaultHandler) ListKeys(c echo.Context) error {
	claims := middleware.GetClaims(c)
	if claims == nil {
		return response.Err(c, domain.ErrUnauthorized)
	}
	if claims.Plan != "pro" {
		return response.Err(c, domain.ErrProPlanRequired)
	}

	vaultID := c.Param("id")
	env := c.QueryParam("env")

	// Verify vault ownership
	if _, err := h.store.GetVault(vaultID, claims.UserID); err != nil {
		return response.Err(c, err)
	}

	envs, _ := h.keyStore.GetEnvironments(c.Request().Context(), vaultID)
	if env == "" && len(envs) > 0 {
		env = envs[0]
	}

	keys, _ := h.keyStore.GetKeyMetadata(c.Request().Context(), vaultID, env)
	if keys == nil {
		keys = []domain.VaultKeyMeta{}
	}

	return response.OK(c, http.StatusOK, map[string]any{
		"vault_id":     vaultID,
		"environment":  env,
		"keys":         keys,
		"environments": envs,
	})
}

func (h *VaultHandler) audit(userID, vaultID, action, detail, ip string) {
	if err := h.store.CreateAuditLog(&domain.AuditLog{
		UserID:    userID,
		VaultID:   vaultID,
		Action:    action,
		Detail:    detail,
		IPAddress: ip,
	}); err != nil {
		slog.Error("audit.log.failed", "action", action, "user_id", userID, "error", err)
	}
}
