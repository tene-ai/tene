package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// QueueEntry represents a pending sync operation queued for offline retry.
type QueueEntry struct {
	Action    string    `json:"action"`    // "push" or "pull"
	VaultID   string    `json:"vault_id"`
	Project   string    `json:"project"`
	Env       string    `json:"env"`
	QueuedAt  time.Time `json:"queued_at"`
	VaultPath string    `json:"vault_path"`
}

// SyncQueue manages offline sync operations in .tene/sync_queue.json.
type SyncQueue struct {
	path string
}

// NewSyncQueue creates a queue backed by the project's .tene directory.
func NewSyncQueue(projectDir string) *SyncQueue {
	return &SyncQueue{
		path: filepath.Join(projectDir, ".tene", "sync_queue.json"),
	}
}

// Enqueue adds a sync operation to the offline queue.
func (q *SyncQueue) Enqueue(entry QueueEntry) error {
	entries, _ := q.List()
	entry.QueuedAt = time.Now().UTC()
	entries = append(entries, entry)
	return q.save(entries)
}

// List returns all pending queue entries.
func (q *SyncQueue) List() ([]QueueEntry, error) {
	data, err := os.ReadFile(q.path)
	if err != nil {
		return nil, nil // no queue file = empty queue
	}
	var entries []QueueEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, nil
	}
	return entries, nil
}

// Dequeue removes the first entry from the queue.
func (q *SyncQueue) Dequeue() (*QueueEntry, error) {
	entries, _ := q.List()
	if len(entries) == 0 {
		return nil, nil
	}
	first := entries[0]
	return &first, q.save(entries[1:])
}

// Clear removes all entries from the queue.
func (q *SyncQueue) Clear() error {
	return os.Remove(q.path)
}

// IsEmpty returns true if the queue has no pending entries.
func (q *SyncQueue) IsEmpty() bool {
	entries, _ := q.List()
	return len(entries) == 0
}

func (q *SyncQueue) save(entries []QueueEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(q.path, data, 0600)
}
