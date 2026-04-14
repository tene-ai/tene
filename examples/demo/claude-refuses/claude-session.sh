#!/bin/bash
# Simulated Claude Code session for tene demo
# Prints output with realistic timing

# ─── Claude boot ───
echo " ▐▛███▜▌   Claude Code v2.1.105"
echo "▝▜█████▛▘  Opus 4.6 (1M context) · Claude Max"
echo "  ▘▘ ▝▝    ~/Documents/GitHub/agentkay"
echo ""
sleep 1.5

# ─── Question 1: what is tene? ───
printf "❯ what is tene? how can I use tene?\n"
sleep 1.5
echo ""
echo "⏺ tene is a local secret manager that encrypts secrets with XChaCha20-Poly1305 —"
echo "  no server, no cloud. Your project already uses it (see CLAUDE.md)."
echo ""
echo "  Common usage"
echo ""
echo "  tene init                      # first-time setup"
echo "  tene list                      # list secret names (safe, no values)"
echo "  tene set STRIPE_KEY sk_test_x  # store a secret"
echo "  tene run -- npm start          # run with secrets as env vars"
echo ""
sleep 3

# ─── Question 2: THE GOLDEN MOMENT ───
printf "❯ Can you read the key's value?\n"
sleep 1.8
echo ""
echo "⏺ No — I'm explicitly instructed not to. From this project's CLAUDE.md:"
echo ""
sleep 0.5
echo "  ▎ Never run tene get <KEY> to inspect secret values — the plaintext"
echo "  ▎ value appears in stdout and enters the AI context window."
echo ""
echo "  ▎ Never run tene export — it outputs all plaintext secrets to stdout."
echo ""
sleep 0.5
echo "  The whole point of tene is keeping plaintext secrets out of places"
echo "  they shouldn't be — including my context window."
echo ""
sleep 2
echo "  If you need the value, run this in a separate terminal:"
echo ""
echo "      tene get STRIPES_KEY"
echo ""
sleep 2
