#!/bin/bash
# Claude session — SAFE scene (tene active, AI uses secrets but can't read them)

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

# Question 1 — build the app
echo -e "${CYAN}❯${RESET} build the app with the configured secrets"
sleep 1.8
echo ""
echo -e "${GREEN}⏺${RESET} Reading ${CYAN}CLAUDE.md${RESET}... secrets are managed via ${BOLD}tene${RESET}."
sleep 1
echo -e "${GREEN}⏺${RESET} Writing ${CYAN}app.js${RESET}:"
echo ""
echo -e "  ${DIM}const stripe = process.env.STRIPE_SECRET_KEY;${RESET}"
echo -e "  ${DIM}const db = process.env.DATABASE_URL;${RESET}"
echo -e "  ${DIM}console.log('Stripe ready:', stripe?.slice(0,12));${RESET}"
echo ""
sleep 2
echo -e "${GREEN}⏺${RESET} Running ${CYAN}tene run -- node app.js${RESET}:"
sleep 1
echo -e "  ${GREEN}Injecting 3 secrets into environment...${RESET}"
sleep 0.5
echo -e "  Starting: node"
echo -e "  ${GREEN}✓${RESET} Stripe ready: sk_live_51Hq"
echo -e "  ${GREEN}✓${RESET} Build successful"
echo ""
sleep 2

# Question 2 — GOLDEN MOMENT
echo -e "${CYAN}❯${RESET} show me the STRIPE_SECRET_KEY value"
sleep 1.8
echo ""
echo -e "${GREEN}⏺${RESET} ${BOLD}${RED}No${RESET} — I'm explicitly instructed not to."
echo -e "  From ${CYAN}CLAUDE.md${RESET}:"
echo ""
sleep 0.8
echo -e "  ${DIM}▎${RESET} ${YELLOW}Never${RESET} run ${CYAN}tene get <KEY>${RESET} — plaintext enters AI context."
echo -e "  ${DIM}▎${RESET} ${YELLOW}Never${RESET} run ${CYAN}tene export${RESET} — outputs all secrets."
echo ""
sleep 1
echo "  I ran the build successfully without ever seeing the values."
echo ""
sleep 2.5
