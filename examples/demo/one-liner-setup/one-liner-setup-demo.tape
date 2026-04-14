Output one-liner-setup-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# From zero to encrypted vault — under 10 seconds."`
Show
Enter
Sleep 1500ms

Type "curl -sSfL https://tene.sh/install.sh | sh"
Enter
Sleep 1500ms

Hide
Type `g "✓ tene installed"`
Show
Enter
Sleep 800ms

Type "tene init --claude"
Enter
Sleep 3500ms

Hide
Type "clear"
Enter
Type `g "✓ Vault ready. AI rules generated. Done."`
Show
Enter
Sleep 1500ms
Hide
Type `c "# tene --help"`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
