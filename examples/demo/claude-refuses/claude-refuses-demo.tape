Output claude-refuses-demo.gif

Source ../_shared/setup.tape

Hide
Type `mkdir -p /tmp/tene-demo-bin && cp /Users/popup-kay/Documents/GitHub/agentkay/tene-biz/tene/examples/demo/_shared/claude-replay.sh /tmp/tene-demo-bin/.claude.sh && chmod +x /tmp/tene-demo-bin/.claude.sh`
Enter
Type `printf '#!/bin/bash\nexec /tmp/tene-demo-bin/.claude.sh\n' > /tmp/tene-demo-bin/claude && chmod +x /tmp/tene-demo-bin/claude`
Enter
Type `export PATH=/tmp/tene-demo-bin:$HOME/.local/bin:$PATH`
Enter
Type "cd $(mktemp -d)"
Enter
Type "clear"
Enter

Type `c "# Step 1: encrypt your secrets with tene."`
Show
Enter
Sleep 1500ms

Type "tene init --claude"
Enter
Sleep 3s

Type "tene set STRIPE_KEY sk_live_51HqY0K8F4jD9aWx"
Enter
Sleep 1500ms

Type "tene list"
Enter
Sleep 2500ms

Hide
Type `c "# Step 2: ask Claude to read the key value."`
Show
Enter
Sleep 1500ms

Type "tene run -- claude"
Enter
Sleep 14s

Hide
Type "clear"
Enter
Type `m "# AI uses secrets. Never sees them."`
Show
Enter
Sleep 800ms
Hide
Type `c "# curl -sSfL https://tene.sh/install.sh | sh"`
Show
Enter
Sleep 2500ms

Hide
Type "rm -rf .tene CLAUDE.md .cursor .windsurfrules GEMINI.md AGENTS.md .gitignore /tmp/tene-demo-bin 2>/dev/null"
Enter
