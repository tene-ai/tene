Output multi-env-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# Same KEY. Different value per environment."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Type "tene env create dev"
Enter
Sleep 800ms
Type "tene env create prod"
Enter
Sleep 1500ms

Type "tene set DATABASE_URL postgres://localhost:5432/dev --env dev"
Enter
Sleep 1500ms

Type "tene set DATABASE_URL postgres://prod.db.example.com/main --env prod"
Enter
Sleep 1500ms

Type "tene env list"
Enter
Sleep 2500ms

Hide
Type `c "# Run in dev:"`
Show
Enter
Sleep 800ms
Type `tene run --env dev -- bash -c 'echo DATABASE_URL=$DATABASE_URL'`
Enter
Sleep 2500ms

Hide
Type `c "# Run in prod:"`
Show
Enter
Sleep 800ms
Type `tene run --env prod -- bash -c 'echo DATABASE_URL=$DATABASE_URL'`
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Type `m "# One key. Any environment. Zero config files."`
Show
Enter
Sleep 1500ms
Hide
Type `c "#   tene --help"`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
