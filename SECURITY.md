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
  Windows Credential Vault). See "Master Key Storage Modes" below for the
  fallback hierarchy and the security stance of each mode.
- **Recovery**: 12-word BIP-39 mnemonic
- **Network**: Zero network calls from the CLI by default
- **Audit**: All encryption primitives live in `pkg/crypto/` and can be
  inspected under the MIT license.

## Master Key Storage Modes

The master key (derived from your password via Argon2id) is held in one of
four modes. Tene picks one at every CLI invocation based on the
`--no-keychain` flag and the `TENE_KEYFILE` environment variable. The
trade-offs differ — read this section before automating tene in CI/CD.

| Mode | Selected when | Persistence | When to use | Trust model |
| --- | --- | --- | --- | --- |
| **OS Keychain** | default (no flag) | OS-managed encrypted store | Local dev, single-user laptop | OS keychain ACL gates `Load()` |
| **Auto file fallback** | OS keychain probe failed (Docker, headless Linux without libsecret) | `~/.tene/keyfile` (mode 0600) | CI hosts without a keychain | File mode 0600; user must trust the host |
| **Explicit file (`TENE_KEYFILE=path`)** | `--no-keychain` + `TENE_KEYFILE` set | user-chosen path (mode 0600) | Daemons that cannot pipe a password every call | User picks the path; isolate per project |
| **`NullStore` (no persistence)** | `--no-keychain` alone, since v1.0.14 | none | CI/CD with `TENE_MASTER_PASSWORD` per call | Every invocation re-derives from the env var |

The fourth mode — `NullStore` — exists because the previous shared
`~/.tene/keyfile` fallback caused a real cross-project key-bleed (filed
as B1 in the v1.0.14-rc1 QA cycle): a second project's `tene init
--no-keychain` overwrote the first project's master key in the shared
file, after which any subsequent decrypt from the first project succeeded
with the *second* project's password. The fix restores the obvious
contract: `--no-keychain` means no persistent key on disk, period.

If your automation relied on the old shared file (so it did not have to
re-supply the password on every call), you have two safe replacements:

1. Pipe `TENE_MASTER_PASSWORD` on every invocation (recommended for CI/CD).
2. Set `TENE_KEYFILE=/secure/per-project/path` to opt back into a
   file-backed store at a path **you** control. Pick a path that is not
   shared across projects.

Whichever you pick, the file mode is `0600` and the directory is created
with mode `0700`. Tene never grants other users access to your key file;
that part of the contract is unchanged.

### OS Keychain probe vs. master-key storage (sprint `keychain-probe-fixed`)

On macOS / Linux libsecret / Windows Credential Manager, tene uses
two distinct service names against the OS keychain:

| Service name | Purpose | Lifetime | Per-project? |
| --- | --- | --- | --- |
| `tene-probe` | One-shot Set/Delete to verify the OS keychain accepts writes (selects KeyringStore vs. FileStore fallback) | Created on first `tene` invocation per host, deleted immediately after, but the service registration / ACL grant persists | **No** — single fixed name shared across all projects |
| `tene-<hashPath(projectDir)>` | Stores the master key for one specific project's vault | Persists for the life of the vault (cleared on `tene` uninstall) | **Yes** — one per project |

The probe service is shared because its only job is "does the
keychain work?" — a single registration is enough. The storage
service is per-project because a leaked master key for project A
must not unlock project B's vault.

Prior to v1.0.15 the probe wrote to the per-project service, which
made macOS Keychain accumulate one ACL-registered entry per project
directory the user touched. Inert leftovers from that period can be
removed via the snippet in `CHANGELOG.md` for [Unreleased].

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
