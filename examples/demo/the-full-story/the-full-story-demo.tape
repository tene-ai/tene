# DEMO 13: The Full Story (Hide-Type-Show-Enter pattern, real-looking commands)
# Trick: a fake `claude` executable on PATH dispatches to danger or safe replay
#        based on whether .tene/vault.db exists. Viewer sees `claude` / `tene run -- claude`.

Output demo.gif

Set FontSize 20
Set FontFamily "Menlo"
Set Width 1280
Set Height 720
Set Theme "Catppuccin Mocha"
Set TypingSpeed 100ms
Set Padding 30
Set WindowBar Colorful
Set BorderRadius 10
Set Margin 30
Set MarginFill "#1e1e2e"
Set Shell "bash"

# ───────── Setup (fully hidden) ─────────
Hide
Type `export PS1="\[\033[38;5;147m\]$\[\033[0m\] "`
Enter
Type `export TENE_MASTER_PASSWORD="demopass123"`
Enter

# Color-comment helpers (self-erase)
Type `r() { printf '\033[1A\033[2K\r\033[1;31m%s\033[0m\n' "$*"; }`
Enter
Type `y() { printf '\033[1A\033[2K\r\033[1;33m%s\033[0m\n' "$*"; }`
Enter
Type `g() { printf '\033[1A\033[2K\r\033[1;32m%s\033[0m\n' "$*"; }`
Enter
Type `c() { printf '\033[1A\033[2K\r\033[1;36m%s\033[0m\n' "$*"; }`
Enter
Type `m() { printf '\033[1A\033[2K\r\033[1;35m%s\033[0m\n' "$*"; }`
Enter

# Install fake `claude` wrapper on PATH (dispatches based on vault presence)
Type `mkdir -p /tmp/tene-demo-bin && cp /Users/popup-kay/Documents/GitHub/agentkay/tene-biz/tene/examples/demo/the-full-story/claude-danger.sh /tmp/tene-demo-bin/.danger.sh && cp /Users/popup-kay/Documents/GitHub/agentkay/tene-biz/tene/examples/demo/the-full-story/claude-safe.sh /tmp/tene-demo-bin/.safe.sh && chmod +x /tmp/tene-demo-bin/.danger.sh /tmp/tene-demo-bin/.safe.sh`
Enter
Type `printf '#!/bin/bash\nif [ -f .tene/vault.db ]; then exec /tmp/tene-demo-bin/.safe.sh; else exec /tmp/tene-demo-bin/.danger.sh; fi\n' > /tmp/tene-demo-bin/claude && chmod +x /tmp/tene-demo-bin/claude`
Enter
Type `export PATH=/tmp/tene-demo-bin:$HOME/.local/bin:$PATH`
Enter

# Working directory + initial .env
Type "cd $(mktemp -d)"
Enter
Type `printf 'STRIPE_SECRET_KEY=sk_live_51HqY0K8F4jD9aWx\nOPENAI_API_KEY=sk-proj-abc123xyz789\nDATABASE_URL=postgres://admin:hunter2@prod.db\n' > .env`
Enter
Type "clear"
Enter

# ───────── Act 1: The Danger ─────────
Type `r "# Before tene: .env is visible to every AI agent."`
Show
Enter

Sleep 1500ms

Type "cat .env"
Enter
Sleep 3s

Hide
Type `y "# What does Claude actually see?"`
Show
Enter

Sleep 1s

Type "claude"
Enter
Sleep 13s

# ───────── Act 2: The Migration ─────────
Hide
Type "clear"
Enter
Type `y "# Time to migrate. tene takes care of it."`
Show
Enter

Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 4s

Type "tene import .env"
Enter
Sleep 2s

Type "rm .env && ls -la | grep -E '\\.env|CLAUDE'"
Enter
Sleep 3s

Hide
Type `g "# .env deleted. Secrets encrypted. Claude has new rules."`
Show
Enter

Sleep 2500ms

# ───────── Act 3: The Proof ─────────
Hide
Type "clear"
Enter
Type `c "# Run with tene — AI uses secrets but never sees their values."`
Show
Enter

Sleep 1500ms

Type "tene run -- claude"
Enter
Sleep 22s

# ───────── End card ─────────
Hide
Type "clear"
Enter
Type `m "# AI uses your secrets."`
Show
Enter

Sleep 600ms

Hide
Type `m "# AI never sees them."`
Show
Enter

Sleep 800ms

Hide
Type `c "#   curl -sSfL https://tene.sh/install.sh | sh"`
Show
Enter

Sleep 600ms

Hide
Type `c "#   tene --help"`
Show
Enter

Sleep 4s

# ───────── Cleanup (hidden) ─────────
Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore /tmp/tene-demo-bin 2>/dev/null"
Enter
