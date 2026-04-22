#!/usr/bin/env bash
#
# Monthly review helper for the ai-discoverability feature.
#
# Prints the live state of every T1/T2/T3 KPI so you can update
# docs/stats/ai-discoverability.md. Does not modify any file — it is a
# read-only probe.
#
# Usage:  ./scripts/ai-discoverability/review.sh [--json]

set -uo pipefail   # intentionally no -e: partial failures (e.g. gh missing)
                   # should still let the remaining probes run.

json_mode=false
[[ "${1:-}" == "--json" ]] && json_mode=true

say() { $json_mode || echo "$@"; }
header() { $json_mode || { echo; echo "─── $1 ────────────────────────────"; }; }

# --- T1: GitHub metadata ---------------------------------------------------
header "T1 — GitHub metadata (tomo-kay/tene)"
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  topics_count=$(gh api repos/tomo-kay/tene --jq '.topics | length' 2>/dev/null || echo "?")
  stars=$(gh api repos/tomo-kay/tene --jq '.stargazers_count' 2>/dev/null || echo "?")
  description=$(gh api repos/tomo-kay/tene --jq '.description' 2>/dev/null || echo "?")
  say "topics count:  $topics_count  (target: 20)"
  say "stars:         $stars"
  say "description:   $description"
else
  say "gh CLI unavailable or not authenticated — skipping T1"
  topics_count="?"; stars="?"; description="?"
fi

# --- T2: llms.txt routes ---------------------------------------------------
header "T2 — llms.txt routes"
llms_status=$(curl -sS -o /dev/null -w '%{http_code}' https://tene.sh/llms.txt || echo 000)
llms_full_status=$(curl -sS -o /dev/null -w '%{http_code}' https://tene.sh/llms-full.txt || echo 000)
root_llms_status=$(curl -sS -o /dev/null -w '%{http_code}' https://raw.githubusercontent.com/tomo-kay/tene/main/llms.txt || echo 000)
say "tene.sh/llms.txt:          HTTP $llms_status      (want 200)"
say "tene.sh/llms-full.txt:     HTTP $llms_full_status      (want 200)"
say "github root llms.txt:      HTTP $root_llms_status      (want 200)"

# --- T3: comparison pages --------------------------------------------------
header "T3 — comparison pages"
slugs=(dotenv doppler dotenv-vault infisical vault)
indexed_count=0
for slug in "${slugs[@]}"; do
  code=$(curl -sS -o /dev/null -w '%{http_code}' "https://tene.sh/vs/$slug" || echo 000)
  if [[ "$code" == "200" ]]; then
    schema_count=$(curl -sS "https://tene.sh/vs/$slug" | grep -c '"@type":"SoftwareApplication"' || true)
    indexed_count=$((indexed_count + 1))
  else
    schema_count=0
  fi
  say "$(printf '%-14s http=%s  schema=%s' "$slug" "$code" "$schema_count")"
done
say
say "deployed pages:  $indexed_count / ${#slugs[@]}"

# --- T4: awesome-list PR watch ---------------------------------------------
header "T4 — awesome-list PRs (authored by current gh user)"
if command -v gh >/dev/null 2>&1 && gh auth status >/dev/null 2>&1; then
  for repo in mahseema/awesome-ai-tools sbilly/awesome-security devsecops/awesome-devsecops agarrharr/awesome-cli-apps avelino/awesome-go; do
    prs=$(gh pr list --repo "$repo" --author "@me" --state all --json state,number,title 2>/dev/null || echo '[]')
    count=$(echo "$prs" | jq 'length' 2>/dev/null)
    [[ -z "$count" ]] && count=0
    say "$(printf '%-35s PRs:%s' "$repo" "$count")"
  done
else
  say "gh CLI unavailable — skipping T4"
fi

exit 0

# --- JSON output ----------------------------------------------------------
if $json_mode; then
  cat <<JSON
{
  "generated_at": "$(date -u +%FT%TZ)",
  "t1": { "topics_count": "$topics_count", "stars": "$stars" },
  "t2": {
    "llms_txt": $llms_status,
    "llms_full_txt": $llms_full_status,
    "root_llms_txt": $root_llms_status
  },
  "t3": { "deployed_pages": $indexed_count, "total_pages": ${#slugs[@]} }
}
JSON
fi
