Output full-tour-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `c "━━ 1. Install & Init ━━"`
Show
Enter
Sleep 1s

Type "tene version"
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Hide
Type "clear"
Enter
Type `c "━━ 2. Set / List / Get ━━"`
Show
Enter
Sleep 1s

Type "tene set API_KEY sk_xxx"
Enter
Sleep 800ms
Type "tene set DB_URL postgres://localhost"
Enter
Sleep 800ms
Type "tene list"
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Type `c "━━ 3. Environments ━━"`
Show
Enter
Sleep 1s

Type "tene env create dev"
Enter
Sleep 600ms
Type "tene env create prod"
Enter
Sleep 600ms
Type "tene set DB_URL postgres://prod.db --env prod"
Enter
Sleep 1s
Type "tene env list"
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Type `c "━━ 4. Import from .env ━━"`
Show
Enter
Sleep 1s

Hide
Type `printf 'AWS_KEY=AKIA123\nREDIS_URL=redis://local\n' > .env`
Enter
Show
Sleep 600ms
Type "tene import .env --env dev"
Enter
Sleep 1500ms
Type "rm .env"
Enter
Sleep 500ms
Type "tene list --env dev"
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Type `c "━━ 5. Runtime Injection ━━"`
Show
Enter
Sleep 1s

Type `tene run --env prod -- bash -c 'echo DB_URL=$DB_URL'`
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Type `c "━━ 6. Status ━━"`
Show
Enter
Sleep 1s

Type "tene whoami"
Enter
Sleep 3s

Hide
Type "clear"
Enter
Type `m "# Encrypted. Local. AI-safe."`
Show
Enter
Sleep 600ms
Hide
Type `m "# MIT. Free forever."`
Show
Enter
Sleep 600ms
Hide
Type `c "#   curl -sSfL https://tene.sh/install.sh | sh"`
Show
Enter
Sleep 600ms
Hide
Type `c "#   tene --help"`
Show
Enter
Sleep 3s

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
