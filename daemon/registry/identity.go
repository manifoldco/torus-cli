package registry

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/dchest/blake2b"
)

const (
	base32Alphabet = "0123456789abcdefghjkmnpqrtuvwxyz"
	idVersion      = 0x01
)

type AgObject interface {
	Type() byte
}

var LowerBase32 = base32.NewEncoding(base32Alphabet)

type ID [18]byte

func NewID(version uint8, body AgObject, sig *Signature) (ID, error) {
	h, err := blake2b.New(&blake2b.Config{Size: 16})
	if err != nil {
		return ID{}, err
	}

	h.Write([]byte(strconv.Itoa(int(version))))

	b, err := json.Marshal(body)
	if err != nil {
		return ID{}, err
	}
	h.Write(b)

	b, err = json.Marshal(sig)
	if err != nil {
		return ID{}, err
	}
	h.Write(b)

	id := ID{idVersion, body.Type()}

	copy(id[2:], h.Sum(nil))

	return id, nil
}

func NewIDFromString(raw string) (*ID, error) {
	id := &ID{}
	err := id.fillID([]byte(raw))
	return id, err
}

func (id *ID) MarshalJSON() ([]byte, error) {
	b32 := LowerBase32.EncodeToString(id[:])
	return []byte("\"" + strings.TrimRight(b32, "=") + "\""), nil
}

func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != byte('"') || b[len(b)-1] != byte('"') {
		return errors.New("value is not a string")
	}

	return id.fillID(b[1 : len(b)-1])
}

func (id *ID) fillID(raw []byte) error {
	pad := 8 - (len(raw) % 8)
	nb := make([]byte, len(raw)+pad)
	copy(nb, raw)
	for i := 0; i < pad; i++ {
		nb[len(raw)+i] = '='
	}

	out, err := LowerBase32.DecodeString(string(nb))
	if err != nil {
		return err
	}
	if len(out) != 18 {
		return errors.New("Incorrect length for id")
	}

	copy(id[:], out)
	return nil
}
