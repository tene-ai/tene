Output ai-injection-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# AI writes code that uses secrets — never sees them."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Type "tene set STRIPE_KEY sk_live_secret_xyz"
Enter
Sleep 1500ms

Type `printf 'const stripe = process.env.STRIPE_KEY;\nconsole.log("Using key:", stripe?.slice(0, 12) + "...");\nconsole.log("Ready for Stripe SDK.");\n' > app.js`
Enter
Sleep 1500ms

Type "cat app.js"
Enter
Sleep 3s

Hide
Type `c "# Run with tene — injects STRIPE_KEY as env var."`
Show
Enter
Sleep 1s

Type "tene run -- node app.js"
Enter
Sleep 3s

Hide
Type "clear"
Enter
Type `g "# Secret never appeared in code."`
Show
Enter
Sleep 600ms
Hide
Type `g "# Never in Claude's context."`
Show
Enter
Sleep 2s

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore app.js 2>/dev/null"
Enter
