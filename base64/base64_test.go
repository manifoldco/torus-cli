package base64

import (
	"bytes"
	"testing"
)

func TestBase64(t *testing.T) {
	b := []byte{'t', 'e', 's', 't', 0x0}
	v := NewValue(b)

	out, err := v.MarshalJSON()
	if err != nil {
		t.Fatal("failed to marshal value")
	}

	if string(out) != `"dGVzdAA"` {
		t.Error("marshaled value did not match expected.")
	}

	in := Value{}
	err = in.UnmarshalJSON(out)
	if err != nil {
		t.Fatal("failed to unmarshal value")
	}

	if !bytes.Equal(b, in) {
		t.Error("unmarshal failed")
	}
}

func TestBase64UnmarshalErrs(t *testing.T) {
	for _, tc := range []string{`"`, "15", `"Not base 64"`} {
		t.Run(tc, func(t *testing.T) {
			in := Value{}
			err := in.UnmarshalJSON([]byte(tc))
			if err == nil {
				t.Error(tc, "did not error")
			}
		})
	}
}
