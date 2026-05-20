package cli

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/agent-kay-it/tene/pkg/crypto"
	teneerr "github.com/agent-kay-it/tene/pkg/errors"
	"golang.org/x/term"
)

var (
	keyNameRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	envNameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

	reservedKeys = map[string]bool{
		"PATH": true, "HOME": true, "USER": true,
		"SHELL": true, "TENE_MASTER_PASSWORD": true,
	}
)

func validateKeyName(name string) error {
	if len(name) == 0 || len(name) > 256 {
		return teneerr.ErrInvalidKeyName(name)
	}
	if !keyNameRegex.MatchString(name) {
		return teneerr.ErrInvalidKeyName(name)
	}
	if reservedKeys[name] {
		return teneerr.ErrReservedKeyName(name)
	}
	return nil
}

func validateEnvName(name string) error {
	if len(name) == 0 || len(name) > 64 {
		return teneerr.ErrInvalidEnvName
	}
	if !envNameRegex.MatchString(name) {
		return teneerr.ErrInvalidEnvName
	}
	return nil
}

func promptPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(password), nil
}

func promptPasswordConfirm(prompt string) (string, error) {
	// Check for env var first
	if pw := os.Getenv("TENE_MASTER_PASSWORD"); pw != "" {
		if len(pw) < 8 {
			return "", teneerr.ErrPasswordTooShort
		}
		return pw, nil
	}

	if !isTerminal() {
		return "", teneerr.ErrInteractiveRequired
	}

	for attempts := 0; attempts < 3; attempts++ {
		pw, err := promptPassword(prompt)
		if err != nil {
			return "", err
		}
		if len(pw) < 8 {
			fmt.Fprintln(os.Stderr, "Master Password must be at least 8 characters.")
			continue
		}

		confirm, err := promptPassword("Confirm Master Password: ")
		if err != nil {
			return "", err
		}
		if pw != confirm {
			fmt.Fprintln(os.Stderr, "Passwords do not match. Try again.")
			continue
		}
		return pw, nil
	}
	return "", teneerr.ErrPasswordMismatch
}

func promptConfirm(msg string) bool {
	if !isTerminal() {
		// Sprint v1014-rc1-qa-fixes / FX2 (invariant I-12).
		//
		// Before v1.0.14 this branch returned true, which meant any
		// destructive verb invoked from a non-TTY shell (CI/CD, an AI
		// agent piping through bash, `tene env delete | tee log`)
		// silently consented to data loss. QA filed it as B2: typing
		// `tene env delete prod` in a non-interactive context wiped
		// the prod environment with no prompt at all.
		//
		// Fail-closed restores the principle that destructive ops
		// require explicit human (or scripted-with-intent) consent.
		// Callers that legitimately want unattended execution opt in
		// with --force; the verb-specific call site checks that flag
		// BEFORE entering promptConfirm, so the only way to reach
		// here is "no flag set, no TTY available" — which is the
		// shape of a mistake, not the shape of intent.
		fmt.Fprintln(os.Stderr, "Refusing to confirm a destructive operation on a non-interactive shell.")
		fmt.Fprintln(os.Stderr, "Pass --force to skip the prompt, or run in a terminal.")
		return false
	}
	fmt.Fprintf(os.Stderr, "%s (y/N) ", msg)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func deriveMasterKey(password string, salt []byte) ([]byte, error) {
	return crypto.DeriveKey(password, salt)
}

func decodeBase64(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// apiErrMsg builds a user-friendly error message from API error responses.
func apiErrMsg(code, message string, status int) string {
	if message != "" {
		return message
	}
	if code != "" {
		return fmt.Sprintf("%s (HTTP %d)", code, status)
	}
	return fmt.Sprintf("API error (HTTP %d)", status)
}

// checkProPlan decodes the JWT payload (without signature verification) and
// returns an error if the plan claim is not "pro". If decoding fails, nil is
// returned so the server can perform the authoritative check.
func checkProPlan(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil // cannot decode; let server decide
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var claims struct {
		Plan string `json:"plan"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}
	if claims.Plan != "" && claims.Plan != "pro" {
		return fmt.Errorf("cloud sync requires a Pro plan ($5/month).\nUpgrade: tene billing upgrade")
	}
	return nil
}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func maskValue(value string) string {
	if len(value) < 5 {
		return "*****"
	}
	return value[:5] + "*****"
}

// writeGitignore creates .tene/.gitignore with content "*"
func writeGitignore(path string) error {
	return os.WriteFile(path, []byte("*\n"), 0600)
}

// addToRootGitignore adds .tene/ to the project root .gitignore
func addToRootGitignore(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(gitignorePath, []byte(".tene/\n"), 0644)
		}
		return err
	}

	if strings.Contains(string(content), ".tene/") {
		return nil // already present
	}

	// Append
	separator := "\n"
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		separator = "\n"
	}
	updated := string(content) + separator + ".tene/\n"
	return os.WriteFile(gitignorePath, []byte(updated), 0644)
}
