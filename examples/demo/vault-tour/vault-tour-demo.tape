Output vault-tour-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# Every command you need. Nothing more."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Type "tene set API_KEY sk_live_xyz"
Enter
Sleep 800ms

Type "tene set DB_URL postgres://localhost/app"
Enter
Sleep 800ms

Type "tene set REDIS_URL redis://localhost"
Enter
Sleep 1500ms

Type "tene list"
Enter
Sleep 3s

Type "tene whoami"
Enter
Sleep 3s

Hide
Type `c "# Delete a secret:"`
Show
Enter
Sleep 800ms

Type "tene delete API_KEY --force"
Enter
Sleep 1500ms

Type "tene list"
Enter
Sleep 2500ms

Hide
Type "clear"
Enter
Type `g "# init → set → list → run → delete."`
Show
Enter
Sleep 600ms
Hide
Type `g "# Encrypted. Local. Free forever."`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
