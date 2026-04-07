package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectConflict_NoConflict(t *testing.T) {
	assert.Nil(t, DetectConflict(1, 1), "same version = no conflict")
	assert.Nil(t, DetectConflict(2, 1), "local ahead = no conflict")
}

func TestDetectConflict_Conflict(t *testing.T) {
	c := DetectConflict(1, 3)
	assert.NotNil(t, c, "remote ahead = conflict")
	assert.Equal(t, int64(1), c.LocalVersion)
	assert.Equal(t, int64(3), c.RemoteVersion)
}

func TestFormatConflict(t *testing.T) {
	c := &ConflictInfo{LocalVersion: 2, RemoteVersion: 5}
	msg := FormatConflict(c)
	assert.Contains(t, msg, "v2")
	assert.Contains(t, msg, "v5")
	assert.Contains(t, msg, "--force")
}
