#!/usr/bin/env bash
#
# Phase 1 — Track 1: GitHub metadata enrichment for AI discoverability.
#
# What this script does (idempotent):
#   1. Replaces the repo's Topics with the 20-item list defined below.
#   2. Replaces the repo's short Description with an AI-agent-friendly phrasing.
#
# Prerequisites:
#   - `gh` CLI authenticated (`gh auth status` must show logged-in).
#   - Repo write permission on github.com/tene-ai/tene.
#
# Usage:
#   ./scripts/ai-discoverability/update-github-metadata.sh          # apply
#   ./scripts/ai-discoverability/update-github-metadata.sh --dry    # print only
#
# Verify after running:
#   gh api repos/tene-ai/tene --jq '{topics, description}'
#
# Source: docs/02-design/features/ai-discoverability.design.md §2.1

set -euo pipefail

REPO="tene-ai/tene"

# Final 20-topic list (GitHub max = 20). Lowercase, alphanumeric + dash only.
# Rationale:
#   - AI discovery keywords: ai-agents, claude-code, cursor, codex, gemini,
#     windsurf, vibe-coding
#   - Category keywords:     secret-management, devsecops, developer-tools,
#                            cli, api-key-management, encryption, vault, dotenv
#   - Language / platform:   go, linux, macos, windows, opensource
# Dropped from old list: "secrets" (superseded by secret-management),
#                        "tene" (project name is not a useful discovery topic).
TOPICS=(
  ai-agents
  api-key-management
  claude-code
  cli
  codex
  cursor
  developer-tools
  devsecops
  dotenv
  encryption
  gemini
  go
  linux
  macos
  opensource
  secret-management
  vault
  vibe-coding
  windows
  windsurf
)

DESCRIPTION="AI-safe secret manager CLI for Claude Code, Cursor, and other AI agents. Local-first, encrypted, no cloud."

dry_run=false
if [[ "${1:-}" == "--dry" || "${1:-}" == "-n" ]]; then
  dry_run=true
fi

echo "Target repo:     $REPO"
echo "Description:     $DESCRIPTION"
echo "Topics (${#TOPICS[@]}):   ${TOPICS[*]}"
echo

if $dry_run; then
  echo "--dry: skipping API calls"
  exit 0
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh CLI not found on PATH" >&2
  exit 1
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "error: gh is not authenticated. Run: gh auth login" >&2
  exit 1
fi

# --- Update topics (idempotent PUT) ---------------------------------------
topic_args=()
for t in "${TOPICS[@]}"; do
  topic_args+=(-f "names[]=$t")
done

echo "Updating topics..."
gh api --method PUT "repos/$REPO/topics" "${topic_args[@]}" >/dev/null

# --- Update description (idempotent PATCH) --------------------------------
echo "Updating description..."
gh api --method PATCH "repos/$REPO" -f "description=$DESCRIPTION" >/dev/null

# --- Verify ---------------------------------------------------------------
echo
echo "Verifying:"
gh api "repos/$REPO" --jq '{topics, description, stargazers_count}'

echo
echo "Done. Commit the change in CHANGELOG or ai-discoverability stats file."
