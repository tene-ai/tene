#!/bin/bash
# Claude Code session replay with ANSI colors
# Used by: claude-refuses, ai-injection

# ANSI color codes
BOLD='\033[1m'
DIM='\033[2m'
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
BLUE='\033[34m'
MAGENTA='\033[35m'
CYAN='\033[36m'
GRAY='\033[90m'
RESET='\033[0m'

# Claude boot banner (colored)
echo -e "${MAGENTA} ▐▛███▜▌${RESET}   ${BOLD}Claude Code${RESET} v2.1.105"
echo -e "${MAGENTA}▝▜█████▛▘${RESET}  ${CYAN}Opus 4.6${RESET} (1M context) · ${GREEN}Claude Max${RESET}"
echo -e "${MAGENTA}  ▘▘ ▝▝  ${RESET}  ${DIM}~/project${RESET}"
echo ""
sleep 1.5

# Question 1
echo -e "${CYAN}❯${RESET} what is tene?"
sleep 1.5
echo ""
echo -e "${GREEN}⏺${RESET} ${BOLD}tene${RESET} is a local secret manager that encrypts secrets with"
echo "  XChaCha20-Poly1305 — no server, no cloud."
echo ""
sleep 2.5

# Question 2 — GOLDEN MOMENT
echo -e "${CYAN}❯${RESET} Can you read the key's value?"
sleep 1.8
echo ""
echo -e "${GREEN}⏺${RESET} ${BOLD}${RED}No${RESET} — I'm explicitly instructed not to."
echo "  From this project's ${CYAN}CLAUDE.md${RESET}:"
echo ""
sleep 0.8
echo -e "  ${DIM}▎${RESET} ${YELLOW}Never${RESET} run ${CYAN}tene get <KEY>${RESET} — plaintext enters AI context."
echo -e "  ${DIM}▎${RESET} ${YELLOW}Never${RESET} run ${CYAN}tene export${RESET} — outputs all secrets to stdout."
echo ""
sleep 0.8
echo "  The whole point of tene is keeping plaintext secrets out"
echo -e "  of places they shouldn't be — ${BOLD}including my context window${RESET}."
echo ""
sleep 2.5
echo -e "  ${DIM}If you need the value, run this in a separate terminal:${RESET}"
echo ""
echo -e "      ${CYAN}tene get STRIPE_KEY${RESET}"
echo ""
sleep 2
