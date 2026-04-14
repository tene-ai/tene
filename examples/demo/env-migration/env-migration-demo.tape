Output env-migration-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type `printf 'STRIPE_SECRET_KEY=sk_live_51HqY0K8F4j\nSTRIPE_PUBLISHABLE_KEY=pk_live_51HqY0K\nOPENAI_API_KEY=sk-proj-abc123xyz789\nANTHROPIC_API_KEY=sk-ant-api03-xxxx\nDATABASE_URL=postgres://admin:hunter2@db\nREDIS_URL=redis://default:pass@redis\nAWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE\nAWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI\nSENTRY_DSN=https://abc123@sentry.io/123\nGOOGLE_CLIENT_ID=12345.apps.googleusercontent.com\n' > .env`
Enter
Type "clear"
Enter

Type `r "# Your .env in plaintext. AI can read every line."`
Show
Enter
Sleep 1500ms

Type "wc -l .env"
Enter
Sleep 2s

Hide
Type `y "# Migrate in one command."`
Show
Enter
Sleep 1s

Type "tene init --claude"
Enter
Sleep 3s

Type "tene import .env"
Enter
Sleep 2s

Type "rm .env"
Enter
Sleep 800ms

Type "tene list"
Enter
Sleep 3s

Hide
Type "clear"
Enter
Type `g "# .env is gone. Secrets are encrypted."`
Show
Enter
Sleep 1s
Hide
Type `c "# https://tene.sh"`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
