package sync

import "fmt"

// ConflictStrategy defines how to resolve sync conflicts.
type ConflictStrategy int

const (
	// StrategyServerWins pulls remote first, merges, then pushes (default).
	StrategyServerWins ConflictStrategy = iota
	// StrategyForcePush overwrites remote with local.
	StrategyForcePush
	// StrategyForcePull overwrites local with remote.
	StrategyForcePull
)

// ConflictInfo describes a version conflict between local and remote vaults.
type ConflictInfo struct {
	LocalVersion  int64  `json:"local_version"`
	RemoteVersion int64  `json:"remote_version"`
	LocalHash     string `json:"local_hash"`
	RemoteHash    string `json:"remote_hash"`
}

// DetectConflict checks if a push would cause a version conflict.
// Returns nil if no conflict (local version matches expected remote version).
func DetectConflict(localVersion, remoteVersion int64) *ConflictInfo {
	if localVersion < remoteVersion {
		return &ConflictInfo{
			LocalVersion:  localVersion,
			RemoteVersion: remoteVersion,
		}
	}
	return nil
}

// FormatConflict returns a human-readable description of the conflict.
func FormatConflict(c *ConflictInfo) string {
	return fmt.Sprintf(
		"Version conflict: local v%d, remote v%d. "+
			"Use --force to overwrite remote, or run 'tene pull' first.",
		c.LocalVersion, c.RemoteVersion,
	)
}
