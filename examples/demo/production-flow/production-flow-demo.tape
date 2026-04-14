Output production-flow-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# Real workflow: Next.js + Stripe, no .env needed."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Type "tene env create prod"
Enter
Sleep 1s

Type "tene set STRIPE_SECRET_KEY sk_test_demo_key --env prod"
Enter
Sleep 1s
Type "tene set STRIPE_PUBLISHABLE_KEY pk_test_demo_pub --env prod"
Enter
Sleep 1s
Type "tene set DATABASE_URL postgres://prod.db/main --env prod"
Enter
Sleep 1500ms

Type "tene list --env prod"
Enter
Sleep 3s

Hide
Type `printf 'const stripeKey = process.env.STRIPE_SECRET_KEY;\nconst dbUrl = process.env.DATABASE_URL;\nconsole.log("Stripe key:", stripeKey?.slice(0, 12) + "...");\nconsole.log("Database:", dbUrl?.slice(0, 20) + "...");\nconsole.log("Checkout ready");\n' > checkout.js`
Enter
Type `c "# Run production build:"`
Show
Enter
Sleep 1s

Type "tene run --env prod -- node checkout.js"
Enter
Sleep 3500ms

Hide
Type `g "# No .env file. Encrypted vault. AI-safe."`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore checkout.js 2>/dev/null"
Enter
