import { TrackedGithubLink } from "./tracked-github-link";

export function Security() {
  return (
    <section id="security" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          No server.{" "}
          <span className="text-accent">Nothing to hack.</span>
        </h2>
        <p className="mx-auto mt-4 max-w-xl text-center text-muted">
          While .env files expose secrets to every AI agent in your project, Tene keeps them encrypted on your device. No server to breach, no database to leak, no API to exploit.
        </p>

        <div className="mt-16 overflow-hidden rounded-xl border border-border bg-surface">
          <div className="border-b border-border px-6 py-4">
            <h3 className="font-mono text-sm text-accent">Encryption architecture</h3>
          </div>
          <div className="p-6 font-mono text-sm leading-7">
            <div className="text-muted">{"// Your device — secrets exist only here"}</div>
            <div className="mt-2" />
            <div>
              <span className="text-muted">Master Password</span>
            </div>
            <div>
              {"  └─ "}
              <span className="text-accent">Argon2id</span>
              <span className="text-muted"> (64MB memory, 3 iterations)</span>
            </div>
            <div>
              {"     └─ "}
              <span className="text-accent">Master Key</span>
              <span className="text-muted"> (256-bit) → OS Keychain</span>
            </div>
            <div>
              {"        └─ "}
              <span className="text-accent">XChaCha20-Poly1305</span>
              <span className="text-muted"> (192-bit nonce)</span>
            </div>
            <div>
              {"           └─ "}
              <span className="text-foreground">SQLite vault</span>
              <span className="text-muted"> (.tene/vault.db)</span>
            </div>
            <div className="mt-4 border-t border-border pt-4 text-muted">
              {"// Network calls:  none"}
              <br />
              {"// Server:         none"}
              <br />
              {"// Attack surface: none"}
            </div>
          </div>
        </div>

        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          <div className="rounded-xl border border-border bg-surface p-5">
            <div className="text-2xl font-bold text-accent">0</div>
            <div className="mt-1 text-sm text-muted">Network calls in free tier</div>
          </div>
          <div className="rounded-xl border border-border bg-surface p-5">
            <div className="text-2xl font-bold text-accent">256-bit</div>
            <div className="mt-1 text-sm text-muted">XChaCha20-Poly1305 encryption</div>
          </div>
          <div className="rounded-xl border border-border bg-surface p-5">
            <div className="text-2xl font-bold text-accent">12 words</div>
            <div className="mt-1 text-sm text-muted">BIP-39 recovery key</div>
          </div>
        </div>

        <div className="mt-8 rounded-xl border border-accent/20 bg-accent/5 p-6">
          <div className="flex items-start gap-4">
            <svg className="mt-0.5 h-5 w-5 shrink-0 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
              <path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <p className="text-sm font-medium">Open source — verify it yourself</p>
              <p className="mt-1 text-sm text-muted">
                Every line of encryption code is open source. Don&apos;t trust us — read the code.{" "}
                <TrackedGithubLink
                  href="https://github.com/tene-ai/tene"
                  location="security"
                  className="text-accent underline underline-offset-2 hover:text-accent-dim"
                >
                  github.com/tene-ai/tene
                </TrackedGithubLink>
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
