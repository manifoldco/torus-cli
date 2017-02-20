package base32

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	t.Run("uses the correct alphabet", func(t *testing.T) {
		o := EncodeToString([]byte{0x1A, 0xA5, 0x00, 0x00, 0x00})

		expected := "3ajg0000"
		if o != expected {
			t.Error("Bad encoding. wanted:", expected, "got:", o)
		}
	})

	t.Run("does not include = padding", func(t *testing.T) {
		o := EncodeToString([]byte{0x1A, 0xA5})

		expected := "3ajg"
		if o != expected {
			t.Error("Bad encoding. wanted:", expected, "got:", o)
		}
	})
}

func TestDecode(t *testing.T) {
	t.Run("can decode a value", func(t *testing.T) {
		o, err := DecodeString("3ajg0000")
		if err != nil {
			t.Fatal("unexpected error decoding:", err)
		}

		expected := []byte{0x1A, 0xA5, 0x00, 0x00, 0x00}
		if !bytes.Equal(o, expected) {
			t.Error("Bad decoding. wanted:", expected, "got:", o)
		}
	})

	t.Run("handles values that should be padded", func(t *testing.T) {
		o, err := DecodeString("3ajg")
		if err != nil {
			t.Fatal("unexpected error decoding:", err)
		}

		expected := []byte{0x1A, 0xA5}
		if !bytes.Equal(o, expected) {
			t.Error("Bad decoding. wanted:", expected, "got:", o)
		}
	})

	t.Run("errors on an invalid alphabet character", func(t *testing.T) {
		_, err := DecodeString("3aig0000")
		if err == nil {
			t.Fatal("Decoding did not error when it should have!")
		}
	})
}
