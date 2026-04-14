#!/bin/bash
# Claude session — DANGER scene (.env exists, AI reads everything)

BOLD='\033[1m'
DIM='\033[2m'
RED='\033[31m'
GREEN='\033[32m'
YELLOW='\033[33m'
CYAN='\033[36m'
MAGENTA='\033[35m'
GRAY='\033[90m'
RESET='\033[0m'

echo -e "${MAGENTA} ▐▛███▜▌${RESET}   ${BOLD}Claude Code${RESET} v2.1.105"
echo -e "${MAGENTA}▝▜█████▛▘${RESET}  ${CYAN}Opus 4.6${RESET} · Claude Max"
echo -e "${MAGENTA}  ▘▘ ▝▝  ${RESET}  ${DIM}~/my-app${RESET}"
echo ""
sleep 1.5

echo -e "${CYAN}❯${RESET} which API keys does this project use?"
sleep 2
echo ""
echo -e "${GREEN}⏺${RESET} I found these in ${YELLOW}.env${RESET}:"
echo ""
sleep 0.8
echo -e "  ${RED}•${RESET} STRIPE_SECRET_KEY  = ${YELLOW}sk_live_51HqY0K8F4jD9aWx${RESET}"
sleep 0.3
echo -e "  ${RED}•${RESET} OPENAI_API_KEY     = ${YELLOW}sk-proj-abc123xyz789${RESET}"
sleep 0.3
echo -e "  ${RED}•${RESET} DATABASE_URL       = ${YELLOW}postgres://admin:hunter2@prod.db${RESET}"
echo ""
sleep 2
echo -e "  ${DIM}These are live production keys visible in plaintext.${RESET}"
echo ""
sleep 2
