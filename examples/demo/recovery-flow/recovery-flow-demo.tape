Output recovery-flow-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# Lost your master password? 12 words save you."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 4s

Hide
Type `r "# ... time passes. Master password forgotten."`
Show
Enter
Sleep 1500ms

Hide
Type `c "# Use the Recovery Key (12-word mnemonic):"`
Show
Enter
Sleep 1s

Type "tene recover"
Enter
Sleep 1500ms
Type "finish south dentist oxygen mercy seven verb stock walnut joy advice hold"
Enter
Sleep 1500ms

Hide
Type `g "✓ Recovery Key verified. New master password set."`
Show
Enter
Sleep 1500ms
Hide
Type `g "✓ Vault restored. All secrets intact."`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
