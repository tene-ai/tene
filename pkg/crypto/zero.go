package crypto

// ZeroBytes fills all bytes in the slice with zeros.
// Uses a volatile pattern to prevent compiler optimization removal.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
	// Prevent compiler from optimizing away the zeroing
	keepAlive(b)
}

// keepAlive prevents the compiler from optimizing away zero operations.
//
//go:noinline
func keepAlive(b []byte) {
	if len(b) > 0 {
		_ = b[0]
	}
}
