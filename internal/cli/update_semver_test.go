package cli

import "testing"

// Sprint v1014-rc1-qa-fixes / FX3.
//
// These tests pin invariant I-13: `tene update` and `tene update --check`
// never recommend a target whose SemVer precedence is lower than the
// current version, and never auto-recommend a pre-release tag to a
// stable-channel user.
//
// The QA reproduction (B3) showed v1.0.14-rc1 → "updateAvailable: true"
// pointing at v1.0.13. shouldOfferUpdate is the entry point that
// returns the boolean rc1 published incorrectly. By unit-testing the
// function directly we cover the decision table without needing the
// network path (fetchFromS3 / fetchFromGitHub) to be live.

func TestShouldOfferUpdate(t *testing.T) {
	// Each case is "what we want users to see" expressed as a row.
	// "name" doubles as a t.Run label so failures point at the rule.
	cases := []struct {
		name               string
		current            string
		latest             string
		includePrerelease  bool
		want               bool
	}{
		{
			name:    "stable upgrade — normal happy path",
			current: "v1.0.13", latest: "v1.0.14",
			includePrerelease: false,
			want:              true,
		},
		{
			name:    "B3 fix — RC must not be downgraded to older stable",
			current: "v1.0.14-rc1", latest: "v1.0.13",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "RC → final stable upgrade is offered automatically",
			current: "v1.0.14-rc1", latest: "v1.0.14",
			includePrerelease: false,
			want:              true,
		},
		{
			name:    "stable user is NOT auto-pushed to a pre-release",
			current: "v1.0.13", latest: "v1.0.14-rc1",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "stable user CAN opt in to pre-release channel",
			current: "v1.0.13", latest: "v1.0.14-rc1",
			includePrerelease: true,
			want:              true,
		},
		{
			name:    "already up to date — never offers",
			current: "v1.0.13", latest: "v1.0.13",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "RC user opted in: newer RC → offer",
			current: "v1.0.14-rc1", latest: "v1.0.14-rc2",
			includePrerelease: true,
			want:              true,
		},
		{
			name:    "RC user without opt-in: newer RC blocked",
			current: "v1.0.14-rc1", latest: "v1.0.14-rc2",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "dev baseline has no comparable version — refuse",
			current: "vdev", latest: "v1.0.14",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "malformed latest — fail-closed",
			current: "v1.0.13", latest: "not-a-semver",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "malformed current — fail-closed",
			current: "garbage", latest: "v1.0.14",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "empty current — fail-closed",
			current: "", latest: "v1.0.14",
			includePrerelease: false,
			want:              false,
		},
		{
			name:    "major bump stable",
			current: "v1.5.0", latest: "v2.0.0",
			includePrerelease: false,
			want:              true,
		},
		{
			name:    "patch bump pre-release lineage",
			current: "v1.0.14-rc1", latest: "v1.0.14-rc1.1",
			includePrerelease: true,
			want:              true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldOfferUpdate(tc.current, tc.latest, tc.includePrerelease)
			if got != tc.want {
				t.Errorf("shouldOfferUpdate(%q, %q, includePrerelease=%v) = %v, want %v",
					tc.current, tc.latest, tc.includePrerelease, got, tc.want)
			}
		})
	}
}

// TestUpdateFlagIncludePrereleaseIsWired is a guard against accidentally
// deleting the new --include-prerelease flag during a future refactor of
// update.go's cobra wiring. The cobra Flags() lookup fails if the flag
// is missing, which fails the test fast and points at the broken init().
func TestUpdateFlagIncludePrereleaseIsWired(t *testing.T) {
	flag := updateCmd.Flags().Lookup("include-prerelease")
	if flag == nil {
		t.Fatal("--include-prerelease flag missing from updateCmd; FX3 wiring was removed")
	}
	if flag.DefValue != "false" {
		t.Errorf("--include-prerelease default should be false (safe default), got %q", flag.DefValue)
	}
}
