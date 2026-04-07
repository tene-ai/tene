package sync

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tomo-kay/tene/internal/crypto"
)

// PushOptions configures a push operation.
type PushOptions struct {
	APIBaseURL  string
	AccessToken string
	VaultID     string
	ProjectName string
	Environment string
	VaultDBPath string
	MasterKey   []byte
	Force       bool // --force: skip conflict check
}

// PushResult contains the result of a push operation.
type PushResult struct {
	VaultID   string `json:"vault_id"`
	Version   int64  `json:"vault_version"`
	Hash      string `json:"vault_hash"`
	Size      int    `json:"size"`
	PushedAt  string `json:"pushed_at"`
}

// PullOptions configures a pull operation.
type PullOptions struct {
	APIBaseURL  string
	AccessToken string
	VaultID     string
	ProjectName string
	Environment string
	VaultDBPath string
	MasterKey   []byte
	Force       bool // --force-pull: overwrite local without merge
}

// PullResult contains the result of a pull operation.
type PullResult struct {
	VaultID  string `json:"vault_id"`
	Version  int64  `json:"vault_version"`
	Hash     string `json:"vault_hash"`
	PulledAt string `json:"pulled_at"`
}

// Engine orchestrates vault sync operations.
type Engine struct {
	httpClient *http.Client
}

// NewEngine creates a new sync engine.
func NewEngine() *Engine {
	return &Engine{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Push encrypts the local vault.db into a Sync Envelope and uploads it.
func (e *Engine) Push(ctx context.Context, opts PushOptions) (*PushResult, error) {
	// 1. Read vault.db
	vaultData, err := os.ReadFile(opts.VaultDBPath)
	if err != nil {
		return nil, fmt.Errorf("sync push: read vault: %w", err)
	}

	// 2. Derive SyncKey
	syncKey, err := DeriveSyncKey(opts.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("sync push: derive key: %w", err)
	}
	defer crypto.ZeroBytes(syncKey)

	// 3. Seal into Sync Envelope (L2 encryption)
	envelope, err := Seal(syncKey, vaultData, opts.ProjectName, opts.Environment)
	if err != nil {
		return nil, fmt.Errorf("sync push: seal: %w", err)
	}

	// 4. Compute checksum
	hash := sha256.Sum256(envelope)
	hashHex := hex.EncodeToString(hash[:])

	// 5. Upload via API with retry
	var result *PushResult
	err = e.withRetry(ctx, 3, func(ctx context.Context) error {
		var pushErr error
		result, pushErr = e.doPush(ctx, opts, envelope, hashHex)
		return pushErr
	})
	if err != nil {
		return nil, err
	}

	// 6. Save sync state
	if err := saveSyncState(opts.VaultDBPath, &syncStateFile{
		VaultID:      result.VaultID,
		Version:      result.Version,
		Hash:         result.Hash,
		LastPushedAt: result.PushedAt,
	}); err != nil {
		slog.Warn("sync.push.state_save_failed", "error", err)
	}

	return result, nil
}

// Pull downloads and decrypts the remote vault blob.
func (e *Engine) Pull(ctx context.Context, opts PullOptions) (*PullResult, error) {
	// 1. Get download URL from API
	manifest, err := e.doGetManifest(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("sync pull: get manifest: %w", err)
	}

	// 2. Download blob from presigned URL
	var envelope []byte
	err = e.withRetry(ctx, 3, func(ctx context.Context) error {
		var dlErr error
		envelope, dlErr = e.doDownload(ctx, manifest.DownloadURL)
		return dlErr
	})
	if err != nil {
		return nil, fmt.Errorf("sync pull: download: %w", err)
	}

	// 3. Verify checksum
	hash := sha256.Sum256(envelope)
	if hex.EncodeToString(hash[:]) != manifest.Hash {
		return nil, fmt.Errorf("sync pull: checksum mismatch: expected %s, got %s",
			manifest.Hash, hex.EncodeToString(hash[:]))
	}

	// 4. Derive SyncKey and decrypt
	syncKey, err := DeriveSyncKey(opts.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("sync pull: derive key: %w", err)
	}
	defer crypto.ZeroBytes(syncKey)

	plaintext, err := Open(syncKey, envelope, opts.ProjectName, opts.Environment)
	if err != nil {
		return nil, fmt.Errorf("sync pull: open envelope: %w", err)
	}

	// 5. Save base snapshot for 3-way merge + backup before overwrite
	if _, statErr := os.Stat(opts.VaultDBPath); statErr == nil {
		// Base snapshot: always keep latest pulled state for future merge
		basePath := filepath.Join(filepath.Dir(opts.VaultDBPath), "vault.db.base")
		_ = copyFile(opts.VaultDBPath, basePath)
		// Timestamped backup for safety
		backupPath := opts.VaultDBPath + ".backup." + time.Now().Format("20060102-150405")
		_ = copyFile(opts.VaultDBPath, backupPath)
	}

	// 6. Write decrypted vault.db
	if err := os.WriteFile(opts.VaultDBPath, plaintext, 0600); err != nil {
		return nil, fmt.Errorf("sync pull: write vault: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if err := saveSyncState(opts.VaultDBPath, &syncStateFile{
		VaultID:      manifest.VaultID,
		Version:      manifest.Version,
		Hash:         manifest.Hash,
		LastPulledAt: now,
	}); err != nil {
		slog.Warn("sync.pull.state_save_failed", "error", err)
	}

	return &PullResult{
		VaultID:  manifest.VaultID,
		Version:  manifest.Version,
		Hash:     manifest.Hash,
		PulledAt: now,
	}, nil
}

// --- API calls ---

type pushAPIResponse struct {
	OK   bool `json:"ok"`
	Data struct {
		VaultID      string `json:"vault_id"`
		VaultVersion int64  `json:"vault_version"`
		VaultHash    string `json:"vault_hash"`
		Size         int    `json:"size"`
		PushedAt     string `json:"pushed_at"`
	} `json:"data"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *Engine) doPush(ctx context.Context, opts PushOptions, blob []byte, hash string) (*PushResult, error) {
	url := opts.APIBaseURL + "/api/v1/vaults/" + opts.VaultID + "/push"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("sync: create push request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+opts.AccessToken)
	req.Header.Set("Content-Type", "application/octet-stream")
	if !opts.Force {
		// Read local sync state for If-Match
		state, _ := loadSyncState(opts.VaultDBPath)
		if state != nil {
			req.Header.Set("If-Match", fmt.Sprintf("%d", state.Version))
		}
	}
	req.Body = io.NopCloser(bytes.NewReader(blob))
	req.ContentLength = int64(len(blob))

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("push API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp pushAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("push API: decode: %w", err)
	}

	if resp.StatusCode == http.StatusConflict {
		return nil, fmt.Errorf("push API: version conflict: %s", apiResp.Message)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("push API: %s - %s", apiResp.Error, apiResp.Message)
	}

	return &PushResult{
		VaultID:  apiResp.Data.VaultID,
		Version:  apiResp.Data.VaultVersion,
		Hash:     apiResp.Data.VaultHash,
		Size:     apiResp.Data.Size,
		PushedAt: apiResp.Data.PushedAt,
	}, nil
}

type pullManifest struct {
	VaultID     string `json:"vault_id"`
	Version     int64  `json:"vault_version"`
	Hash        string `json:"vault_hash"`
	DownloadURL string `json:"download_url"`
}

func (e *Engine) doGetManifest(ctx context.Context, opts PullOptions) (*pullManifest, error) {
	url := opts.APIBaseURL + "/api/v1/vaults/" + opts.VaultID + "/pull"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("sync: create pull request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+opts.AccessToken)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pull API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK   bool         `json:"ok"`
		Data pullManifest `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("pull API: decode: %w", err)
	}
	if !apiResp.OK {
		return nil, fmt.Errorf("pull API: request failed (status %d)", resp.StatusCode)
	}

	return &apiResp.Data, nil
}

func (e *Engine) doDownload(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("sync: create download request: %w", err)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sync: download blob: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: status %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 50<<20)) // 50MB limit
}

// --- Retry ---

func (e *Engine) withRetry(ctx context.Context, maxAttempts int, fn func(ctx context.Context) error) error {
	var lastErr error
	for attempt := range maxAttempts {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if attempt < maxAttempts-1 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			jitter := time.Duration(rand.IntN(int(backoff / 10)))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff + jitter):
			}
		}
	}
	return lastErr
}

// --- Sync state persistence ---

type syncStateFile struct {
	VaultID      string `json:"vault_id"`
	Version      int64  `json:"version"`
	Hash         string `json:"hash"`
	LastPushedAt string `json:"last_pushed_at,omitempty"`
	LastPulledAt string `json:"last_pulled_at,omitempty"`
}

func syncStatePath(vaultDBPath string) string {
	dir := filepath.Dir(vaultDBPath)
	return filepath.Join(dir, "sync_state.json")
}

func loadSyncState(vaultDBPath string) (*syncStateFile, error) {
	data, err := os.ReadFile(syncStatePath(vaultDBPath))
	if err != nil {
		return nil, fmt.Errorf("sync: read sync state: %w", err)
	}
	var state syncStateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("sync: parse sync state: %w", err)
	}
	return &state, nil
}

func saveSyncState(vaultDBPath string, state *syncStateFile) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("sync: marshal sync state: %w", err)
	}
	if err := os.WriteFile(syncStatePath(vaultDBPath), data, 0600); err != nil {
		return fmt.Errorf("sync: write sync state: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("sync: open source file: %w", err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("sync: create destination file: %w", err)
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("sync: copy file: %w", err)
	}
	return nil
}

