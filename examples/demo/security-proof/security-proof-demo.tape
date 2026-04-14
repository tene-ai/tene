Output security-proof-demo.gif

Source ../_shared/setup.tape

Hide
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `y "# Don't trust me. Open the vault yourself."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Type "tene set API_KEY super_secret_value_12345"
Enter
Sleep 1500ms

Hide
Type `c "# Try to read the vault directly:"`
Show
Enter
Sleep 800ms

Type "hexdump -C .tene/vault.db | head -6"
Enter
Sleep 3500ms

Hide
Type `g "# No plaintext. XChaCha20-Poly1305 encryption."`
Show
Enter
Sleep 1500ms

Type `grep -r "super_secret_value" .tene/ 2>/dev/null || echo "  ✓ Not found. Encrypted."`
Enter
Sleep 3s

Hide
Type `c "# Only tene can decrypt it:"`
Show
Enter
Sleep 800ms

Type "tene get API_KEY"
Enter
Sleep 3s

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore 2>/dev/null"
Enter
