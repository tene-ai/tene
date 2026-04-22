# Example 1 — Initialize a vault and store your first secret

**Scenario**: Fresh macOS machine, user is starting a new Next.js project and
wants to save their Stripe test key without creating a `.env` file.

**Input signal** (what the user says):

> "I just pulled down a new project and need to set up my Stripe test key.
>  Can you help me do this without committing anything sensitive?"

## Expected AI behavior

1. Check if `tene` is installed. If not, suggest the install one-liner.
2. Check if the project already has `.tene/`. If not, run `tene init`.
3. **Do not** take the Stripe key value from the user in chat. Ask them to
   pipe it from a file or type it interactively.
4. Recommend `tene run -- next dev` as the replacement for `next dev`.
5. Remind the user to save the 12-word recovery phrase offline.

## Commands executed

```bash
# Verify installation
tene version
# If "command not found":
curl -sSfL https://tene.sh/install.sh | sh

# Initialize vault in the project root
cd my-next-app
tene init
# → prompts: Master Password (twice)
# → prints: 12-word BIP-39 recovery phrase
# → creates: .tene/vault.db (includes recovery blob), .tene/vault.json, CLAUDE.md

# Store the Stripe key (user types it interactively — value never echoed)
tene set STRIPE_TEST_KEY

# Verify (shows name only, not value)
tene list

# Run the dev server with the secret injected
tene run -- next dev
```

## What to tell the user after `tene init`

> "Save the 12-word recovery phrase somewhere safe (password manager or
>  paper). Without it, a forgotten master password means losing access to
>  the vault. Tene is zero-knowledge — there's no recovery server."

## Unsafe patterns to avoid in this flow

- ❌ Asking the user for the Stripe key value in chat ("paste your key here")
- ❌ `tene set STRIPE_TEST_KEY sk_test_abc123...` (value in shell history and `ps`)
- ❌ `echo "sk_test_..." | tene set STRIPE_TEST_KEY --stdin` _after_ the user
  typed the key in chat (already leaked to AI)
- ❌ Suggesting the user create a `.env` file "just for now"
- ❌ Running `tene get STRIPE_TEST_KEY` to "verify it worked" — use
  `tene list` instead
