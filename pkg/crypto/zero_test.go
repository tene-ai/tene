package crypto

import (
	"testing"
)

func TestZeroBytes(t *testing.T) {
	data := []byte{0x41, 0x42, 0x43, 0x44, 0x45}
	ZeroBytes(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("data[%d] = 0x%02x, want 0x00", i, b)
		}
	}
}

func TestZeroBytes_Empty(t *testing.T) {
	// Should not panic on nil or empty slices
	ZeroBytes(nil)
	ZeroBytes([]byte{})
}

func TestZeroBytes_Large(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 0xFF
	}

	ZeroBytes(data)

	for i, b := range data {
		if b != 0 {
			t.Errorf("data[%d] = 0x%02x, want 0x00", i, b)
			break
		}
	}
}
