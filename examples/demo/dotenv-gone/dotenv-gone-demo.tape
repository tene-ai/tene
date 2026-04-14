Output dotenv-gone-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type `printf 'STRIPE_KEY=sk_live_secret_xyz\nDB_URL=postgres://admin:pass@prod\n' > .env`
Enter
Type `printf 'console.log("STRIPE_KEY:", process.env.STRIPE_KEY);\nconsole.log("DB_URL:", process.env.DB_URL);\n' > app.js`
Enter
Type "clear"
Enter

Type `r "# Before: .env has plaintext secrets. AI reads it."`
Show
Enter
Sleep 1500ms

Type "ls -la"
Enter
Sleep 2s

Type "cat .env"
Enter
Sleep 3s

Hide
Type `y "# Migrate to tene."`
Show
Enter
Sleep 1s

Type "tene init --claude && tene import .env"
Enter
Sleep 4s

Hide
Type `r "# Now: delete .env."`
Show
Enter
Sleep 800ms

Type "rm .env && ls -la | grep env"
Enter
Sleep 2500ms

Hide
Type `g "# .env gone. App still works."`
Show
Enter
Sleep 1s

Type "tene run -- node app.js"
Enter
Sleep 3500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore app.js 2>/dev/null"
Enter
