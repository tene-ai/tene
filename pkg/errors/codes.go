package errors

// --- Exit Code 0: success / warning ---

var ErrVaultAlreadyExists = &TeneError{
	Code: "VAULT_ALREADY_EXISTS", Message: "Vault already exists. Use existing vault.", Exit: 0,
}
var ErrKeychainError = &TeneError{
	Code: "KEYCHAIN_ERROR", Message: "Keychain access failed.", Exit: 0,
}

// --- Exit Code 1: general errors ---

var ErrVaultNotFound = &TeneError{
	Code: "VAULT_NOT_FOUND", Message: "Not in a Tene project. Run \"tene init\" first.", Exit: 1,
}

func ErrSecretNotFound(key, env string) *TeneError {
	return Newf("SECRET_NOT_FOUND", 1, "Secret %q not found in %q environment.", key, env)
}

func ErrSecretAlreadyExists(key string) *TeneError {
	return Newf("SECRET_ALREADY_EXISTS", 1, "Secret %q already exists. Use --overwrite to replace.", key)
}

func ErrEnvironmentNotFound(env string) *TeneError {
	return Newf("ENVIRONMENT_NOT_FOUND", 1, "Environment %q not found. Create it with \"tene env create %s\".", env, env)
}

func ErrEnvironmentAlreadyExists(env string) *TeneError {
	return Newf("ENVIRONMENT_ALREADY_EXISTS", 1, "Environment %q already exists.", env)
}

func ErrEnvironmentProtected(env, reason string) *TeneError {
	return Newf("ENVIRONMENT_PROTECTED", 1, "Cannot delete the %q environment. %s", env, reason)
}

func ErrInvalidKeyName(key string) *TeneError {
	return Newf("INVALID_KEY_NAME", 1, "Invalid key name %q. Keys must match [A-Z][A-Z0-9_]*.", key)
}

func ErrReservedKeyName(key string) *TeneError {
	return Newf("RESERVED_KEY_NAME", 1, "Key name %q is reserved.", key)
}

var ErrInvalidEnvName = &TeneError{
	Code: "INVALID_ENV_NAME", Message: "Invalid environment name. Must match [a-z][a-z0-9-]*.", Exit: 1,
}
var ErrEmptyValue = &TeneError{
	Code: "EMPTY_VALUE", Message: "Value cannot be empty.", Exit: 1,
}
var ErrValueTooLarge = &TeneError{
	Code: "VALUE_TOO_LARGE", Message: "Value exceeds maximum size (64KB).", Exit: 1,
}
var ErrEncryptFailed = &TeneError{
	Code: "ENCRYPT_FAILED", Message: "Encryption failed.", Exit: 1,
}

func ErrFileNotFound(path string) *TeneError {
	return Newf("FILE_NOT_FOUND", 1, "File %q not found.", path)
}

func ErrFileParse(path string, line int, detail string) *TeneError {
	return Newf("FILE_PARSE_ERROR", 1, "Failed to parse %q at line %d: %s.", path, line, detail)
}

var ErrPermissionDenied = &TeneError{
	Code: "PERMISSION_DENIED", Message: "Permission denied.", Exit: 1,
}
var ErrDiskFull = &TeneError{
	Code: "DISK_FULL", Message: "Cannot create vault: insufficient disk space.", Exit: 1,
}
var ErrInteractiveRequired = &TeneError{
	Code: "INTERACTIVE_REQUIRED", Message: "This command requires an interactive terminal.", Exit: 1,
}
var ErrInvalidBackupFile = &TeneError{
	Code: "INVALID_BACKUP_FILE", Message: "Invalid encrypted backup file format.", Exit: 1,
}

// --- Exit Code 2: authentication errors ---

var ErrPasswordMismatch = &TeneError{
	Code: "PASSWORD_MISMATCH", Message: "Passwords do not match. Try again.", Exit: 2,
}
var ErrPasswordTooShort = &TeneError{
	Code: "PASSWORD_TOO_SHORT", Message: "Master Password must be at least 8 characters.", Exit: 2,
}
var ErrInvalidPassword = &TeneError{
	Code: "INVALID_PASSWORD", Message: "Invalid Master Password.", Exit: 2,
}
var ErrInvalidRecoveryKey = &TeneError{
	Code: "INVALID_RECOVERY_KEY", Message: "Invalid Recovery Key.", Exit: 2,
}
var ErrDecryptFailed = &TeneError{
	Code: "DECRYPT_FAILED", Message: "Failed to decrypt secret. Master Password may have changed.", Exit: 2,
}

// --- Exit Code 127: command not found ---

func ErrCommandNotFound(cmd string) *TeneError {
	return Newf("COMMAND_NOT_FOUND", 127, "Command %q not found.", cmd)
}
