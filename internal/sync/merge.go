package sync

// SecretEntry represents a single secret for merge comparison.
type SecretEntry struct {
	Key       string `json:"key"`
	Value     []byte `json:"value"` // encrypted value
	Env       string `json:"env"`
	UpdatedAt int64  `json:"updated_at"` // unix timestamp
}

// MergeResult describes the outcome of a 3-way merge.
type MergeResult struct {
	// Merged is the final set of secrets after merge.
	Merged []SecretEntry `json:"merged"`
	// AutoResolved lists keys that were auto-merged without conflict.
	AutoResolved []string `json:"auto_resolved,omitempty"`
	// Conflicts lists keys that require manual resolution.
	Conflicts []MergeConflict `json:"conflicts,omitempty"`
}

// MergeConflict represents a key where local and remote both changed differently.
type MergeConflict struct {
	Key         string       `json:"key"`
	Env         string       `json:"env"`
	BaseValue   *SecretEntry `json:"base,omitempty"`
	LocalValue  *SecretEntry `json:"local"`
	RemoteValue *SecretEntry `json:"remote"`
}

// ThreeWayMerge performs a 3-way merge of secrets.
//
// Rules (per design spec §5.4):
//
//	Base=A, Local=B, Remote=A → use Local (only local changed)
//	Base=A, Local=A, Remote=B → use Remote (only remote changed)
//	Base=A, Local=B, Remote=C → CONFLICT (both changed differently)
//	Base=nil, Local=A, Remote=nil → use Local (local addition)
//	Base=nil, Remote=A, Local=nil → use Remote (remote addition)
//	Base=A, Local=nil, Remote=B → CONFLICT (local deleted, remote modified)
//	Base=A, Local=B, Remote=nil → CONFLICT (local modified, remote deleted)
func ThreeWayMerge(base, local, remote []SecretEntry) *MergeResult {
	baseMap := indexByKeyEnv(base)
	localMap := indexByKeyEnv(local)
	remoteMap := indexByKeyEnv(remote)

	result := &MergeResult{}
	seen := make(map[string]bool)

	// Process all keys from local
	for key, localEntry := range localMap {
		seen[key] = true
		baseEntry, inBase := baseMap[key]
		remoteEntry, inRemote := remoteMap[key]

		switch {
		case !inBase && !inRemote:
			// Local addition, no conflict
			result.Merged = append(result.Merged, localEntry)
			result.AutoResolved = append(result.AutoResolved, localEntry.Key)

		case !inBase && inRemote:
			// Both added — check if same value
			if entriesEqual(localEntry, remoteEntry) {
				result.Merged = append(result.Merged, localEntry)
			} else {
				result.Conflicts = append(result.Conflicts, MergeConflict{
					Key: localEntry.Key, Env: localEntry.Env,
					LocalValue: &localEntry, RemoteValue: &remoteEntry,
				})
			}

		case inBase && !inRemote:
			// Remote deleted
			if entriesEqual(baseEntry, localEntry) {
				// Local unchanged, accept remote deletion (don't add)
				result.AutoResolved = append(result.AutoResolved, localEntry.Key)
			} else {
				// Local modified + remote deleted = conflict
				b := baseEntry
				result.Conflicts = append(result.Conflicts, MergeConflict{
					Key: localEntry.Key, Env: localEntry.Env,
					BaseValue: &b, LocalValue: &localEntry,
				})
			}

		case inBase && inRemote:
			localChanged := !entriesEqual(baseEntry, localEntry)
			remoteChanged := !entriesEqual(baseEntry, remoteEntry)

			switch {
			case !localChanged && !remoteChanged:
				result.Merged = append(result.Merged, localEntry)
			case localChanged && !remoteChanged:
				result.Merged = append(result.Merged, localEntry)
				result.AutoResolved = append(result.AutoResolved, localEntry.Key)
			case !localChanged && remoteChanged:
				result.Merged = append(result.Merged, remoteEntry)
				result.AutoResolved = append(result.AutoResolved, remoteEntry.Key)
			default: // both changed
				if entriesEqual(localEntry, remoteEntry) {
					result.Merged = append(result.Merged, localEntry)
				} else {
					b := baseEntry
					result.Conflicts = append(result.Conflicts, MergeConflict{
						Key: localEntry.Key, Env: localEntry.Env,
						BaseValue: &b, LocalValue: &localEntry, RemoteValue: &remoteEntry,
					})
				}
			}
		}
	}

	// Process remote-only keys (not in local)
	for key, remoteEntry := range remoteMap {
		if seen[key] {
			continue
		}
		baseEntry, inBase := baseMap[key]
		if !inBase {
			// Remote addition
			result.Merged = append(result.Merged, remoteEntry)
			result.AutoResolved = append(result.AutoResolved, remoteEntry.Key)
		} else {
			// Local deleted
			if entriesEqual(baseEntry, remoteEntry) {
				// Remote unchanged, accept local deletion
				result.AutoResolved = append(result.AutoResolved, remoteEntry.Key)
			} else {
				// Local deleted + remote modified = conflict
				b := baseEntry
				result.Conflicts = append(result.Conflicts, MergeConflict{
					Key: remoteEntry.Key, Env: remoteEntry.Env,
					BaseValue: &b, RemoteValue: &remoteEntry,
				})
			}
		}
	}

	return result
}

func indexByKeyEnv(entries []SecretEntry) map[string]SecretEntry {
	m := make(map[string]SecretEntry, len(entries))
	for _, e := range entries {
		m[e.Key+"\x00"+e.Env] = e
	}
	return m
}

func entriesEqual(a, b SecretEntry) bool {
	if len(a.Value) != len(b.Value) {
		return false
	}
	for i := range a.Value {
		if a.Value[i] != b.Value[i] {
			return false
		}
	}
	return true
}
