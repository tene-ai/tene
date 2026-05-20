# Security Policy

Tene takes security seriously. It encrypts secrets locally with
XChaCha20-Poly1305 and makes zero network calls by default.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, email **security@tene.sh** (or contact `@agent-kay-it` privately on
GitHub if the alias is not yet live).

Please include:

- A clear description of the vulnerability
- Steps to reproduce (PoC code welcome)
- Affected version(s)
- Your proposed remediation, if any

We aim to acknowledge within 72 hours and fix critical issues within 14 days.

## Supported Versions

| Version | Supported |
| ------- | :-------: |
| 1.x     | ✅        |
| < 1.0   | ❌        |

## Security Model

- **Encryption**: XChaCha20-Poly1305 (256-bit keys, 192-bit nonces,
  secret name bound as AAD)
- **Key Derivation**: Argon2id (64 MiB memory, 3 iterations)
- **Key Storage**: OS native keychain (macOS Keychain, Linux libsecret,
  Windows Credential Vault)
- **Recovery**: 12-word BIP-39 mnemonic
- **Network**: Zero network calls from the CLI by default
- **Audit**: All encryption primitives live in `pkg/crypto/` and can be
  inspected under the MIT license.

## AI-Safe Design Properties

The CLI actively defends against AI agent secret exfiltration:

- `tene run -- <cmd>` is the primary workflow — secrets are injected as
  environment variables and never printed to stdout.
- `tene list` returns secret **names** only, never values.
- `tene get <KEY>` refuses non-TTY stdout by default to prevent accidental
  leakage into AI agent context windows, log aggregators, and shell history.
  Opt in with `--unsafe-stdout` or `TENE_ALLOW_STDOUT_SECRETS=1` when you
  truly need piped output.
- `.tene/` contents are never served by the landing site (robots.txt
  disallow).

## Vault DB Preview Column (Schema v2)

Starting with schema v2 (sprint cli-ux-permission-model, Q2 user decision)
`vault.db` stores a small plaintext "preview" alongside each encrypted
secret value. This is an explicit, opt-out trade-off that gives `tene list`
a useful one-shot summary without forcing a master-password prompt on every
listing.

### What the preview is

A substring of the original plaintext, governed by two configuration keys
stored in the `vault_meta` table:

- `preview.front` — characters from the start of the value (default **0**, max 8)
- `preview.back`  — characters from the end of the value (default **4**, max 8)
- `front + back ≤ 8` (hard cap, enforced by `pkg/crypto.DerivePreview`)

At the default settings the preview is the last 4 characters only — for
example `…aBc1` for an OpenAI token, `…xyz9` for a Stripe key. It is not
possible to derive the original secret from the preview.

### Threat model

**What an attacker who exfiltrates your `vault.db` learns:**

| `preview` setting | API key prefix exposed? | Last-N chars exposed? |
|---|:---:|:---:|
| `preview.enabled=false` | no | no |
| `preview.front=0, preview.back=4` (default) | **no** | yes |
| `preview.front=4, preview.back=4` (explicit opt-in) | **yes** (sk-, ghp_, AKIA, …) | yes |

At the default, an attacker cannot tell whether `STRIPE_KEY` is actually a
Stripe key, an OpenAI key, or a GitHub token — the prefix is hidden. The
last 4 characters alone are not sufficient to mount targeted social
engineering. With `preview.front>0` you explicitly accept that API key
prefixes (which identify the issuing service) become readable from a
leaked database file.

### Mitigations

1. **Default is safe.** Out of the box tene exposes only the last 4 chars.
2. **`vault.db` file permissions** are chmod 0600 on creation.
3. **`.gitignore`** in newly initialised projects excludes `.tene/`.
4. **Opt-out:** `tene config preview.enabled=false` turns the column into
   an empty string for every future write. Existing rows are not
   retroactively cleared — re-set the secrets you care about, or run
   `tene migrate fill-previews` after re-enabling to repopulate.
5. **Opt-in carries a confirm prompt.** `tene config preview.front=N`
   for `N>0` requires interactive confirmation:

   ```
   WARNING: setting preview.front > 0 will expose API key prefixes (sk-, ghp_, AKIA...) in vault.db.
   This makes service identification possible if vault.db leaks. Continue? [y/N]
   ```

   Pass `--force` to skip the prompt in scripts.

### Boundary

tene protects you against **passive** attackers: someone who later finds
a copied `vault.db` file (in a screenshot, a backup, an old laptop, a
ticket attachment). tene does **not** protect you against **active**
attackers who already have your unlocked vault, your keychain, or your
machine. If your threat model includes an active adversary with your
disk and keychain, you need defenses beyond a local secret manager.

## Security Disclosures Log

_None yet — first valid disclosure will be recorded here (and a CVE
requested if applicable)._

## Bug Bounty

There is no formal program at this time. Security researchers are credited
here when they submit valid reports.
