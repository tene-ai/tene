package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreeWayMerge_OnlyLocalChanged(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}}
	remote := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}

	result := ThreeWayMerge(base, local, remote)
	assert.Empty(t, result.Conflicts)
	assert.Len(t, result.Merged, 1)
	assert.Equal(t, []byte("2"), result.Merged[0].Value, "should use local value")
}

func TestThreeWayMerge_OnlyRemoteChanged(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	remote := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}}

	result := ThreeWayMerge(base, local, remote)
	assert.Empty(t, result.Conflicts)
	assert.Len(t, result.Merged, 1)
	assert.Equal(t, []byte("2"), result.Merged[0].Value, "should use remote value")
}

func TestThreeWayMerge_BothChangedDifferently(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}}
	remote := []SecretEntry{{Key: "A", Value: []byte("3"), Env: "dev"}}

	result := ThreeWayMerge(base, local, remote)
	assert.Len(t, result.Conflicts, 1, "should have conflict")
	assert.Equal(t, "A", result.Conflicts[0].Key)
}

func TestThreeWayMerge_BothChangedSame(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}}
	remote := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}}

	result := ThreeWayMerge(base, local, remote)
	assert.Empty(t, result.Conflicts, "same change = no conflict")
	assert.Len(t, result.Merged, 1)
}

func TestThreeWayMerge_LocalAddition(t *testing.T) {
	base := []SecretEntry{}
	local := []SecretEntry{{Key: "NEW", Value: []byte("val"), Env: "dev"}}
	remote := []SecretEntry{}

	result := ThreeWayMerge(base, local, remote)
	assert.Empty(t, result.Conflicts)
	assert.Len(t, result.Merged, 1)
	assert.Equal(t, "NEW", result.Merged[0].Key)
}

func TestThreeWayMerge_RemoteAddition(t *testing.T) {
	base := []SecretEntry{}
	local := []SecretEntry{}
	remote := []SecretEntry{{Key: "NEW", Value: []byte("val"), Env: "dev"}}

	result := ThreeWayMerge(base, local, remote)
	assert.Empty(t, result.Conflicts)
	assert.Len(t, result.Merged, 1)
	assert.Equal(t, "NEW", result.Merged[0].Key)
}

func TestThreeWayMerge_LocalDeleteRemoteModify(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{} // deleted
	remote := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}} // modified

	result := ThreeWayMerge(base, local, remote)
	assert.Len(t, result.Conflicts, 1, "delete + modify = conflict")
}

func TestThreeWayMerge_LocalModifyRemoteDelete(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{{Key: "A", Value: []byte("2"), Env: "dev"}} // modified
	remote := []SecretEntry{} // deleted

	result := ThreeWayMerge(base, local, remote)
	assert.Len(t, result.Conflicts, 1, "modify + delete = conflict")
}

func TestThreeWayMerge_NeitherChanged(t *testing.T) {
	base := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	local := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}
	remote := []SecretEntry{{Key: "A", Value: []byte("1"), Env: "dev"}}

	result := ThreeWayMerge(base, local, remote)
	assert.Empty(t, result.Conflicts)
	assert.Len(t, result.Merged, 1)
}

func TestThreeWayMerge_MultipleKeys(t *testing.T) {
	base := []SecretEntry{
		{Key: "A", Value: []byte("1"), Env: "dev"},
		{Key: "B", Value: []byte("2"), Env: "dev"},
		{Key: "C", Value: []byte("3"), Env: "dev"},
	}
	local := []SecretEntry{
		{Key: "A", Value: []byte("10"), Env: "dev"}, // changed
		{Key: "B", Value: []byte("2"), Env: "dev"},   // unchanged
		// C deleted
		{Key: "D", Value: []byte("4"), Env: "dev"}, // added
	}
	remote := []SecretEntry{
		{Key: "A", Value: []byte("1"), Env: "dev"},   // unchanged
		{Key: "B", Value: []byte("20"), Env: "dev"},  // changed
		{Key: "C", Value: []byte("30"), Env: "dev"},  // changed
		{Key: "E", Value: []byte("5"), Env: "dev"},   // added
	}

	result := ThreeWayMerge(base, local, remote)

	// A: local changed, remote same → use local (10) — auto
	// B: remote changed, local same → use remote (20) — auto
	// C: local deleted, remote modified → conflict
	// D: local added → merged
	// E: remote added → merged

	assert.Len(t, result.Conflicts, 1)
	assert.Equal(t, "C", result.Conflicts[0].Key)

	mergedKeys := make(map[string][]byte)
	for _, m := range result.Merged {
		mergedKeys[m.Key] = m.Value
	}
	assert.Equal(t, []byte("10"), mergedKeys["A"])
	assert.Equal(t, []byte("20"), mergedKeys["B"])
	assert.Equal(t, []byte("4"), mergedKeys["D"])
	assert.Equal(t, []byte("5"), mergedKeys["E"])
}
