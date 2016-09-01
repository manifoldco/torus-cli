// Package identity defines the ID format used for uniquely identifying
// objects in Arigato.
package identity

import (
	"crypto/rand"
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
	byteLength     = 18
)

// Identifiable is the interface implemented by objects that can be given
// IDs.
type Identifiable interface {
	Version() int
	Type() byte
}

var lowerBase32 = base32.NewEncoding(base32Alphabet)

// ID is an encoded unique identifier for an object.
//
// The first byte holds the schema version of the id itself.
// The second byte holds the type of the object.
// The remaining 16 bytes hold a digest of the contents of the object for
// immutable objects, or a random value for mutable objects.
type ID [18]byte

// Mutable returns a new ID for a mutable object.
func Mutable(body Identifiable) (ID, error) {
	id := ID{idVersion, body.Type()}
	_, err := rand.Read(id[2:])
	if err != nil {
		return ID{}, err
	}
	return id, nil
}

// Immutable returns a new signed ID for an immutable object.
//
// sig should be a registry.Signature type
func Immutable(body Identifiable, sig interface{}) (ID, error) {
	h, err := blake2b.New(&blake2b.Config{Size: 16})
	if err != nil {
		return ID{}, err
	}

	h.Write([]byte(strconv.Itoa(body.Version())))

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

// DecodeFromString returns an ID that is stored in the given string.
func DecodeFromString(value string) (ID, error) {
	buf, err := decodeFromByte([]byte(value))
	if err != nil {
		return ID{}, err
	}

	id := ID{}
	copy(id[:], buf)
	return id, nil
}

// Type returns the binary encoded object type represented by this ID.
func (id *ID) Type() byte {
	return id[1]
}

func (id *ID) String() string {
	return strings.TrimRight(lowerBase32.EncodeToString(id[:]), "=")
}

// MarshalJSON implements the json.Marshaler interface for IDs.
//
// IDs are encoded in unpadded base32.
func (id *ID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + id.String() + "\""), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for IDs.
func (id *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != byte('"') || b[len(b)-1] != byte('"') {
		return errors.New("value is not a string")
	}

	return id.fillID(b[1 : len(b)-1])
}

func (id *ID) fillID(raw []byte) error {
	out, err := decodeFromByte(raw)
	if err != nil {
		return err
	}

	copy(id[:], out)
	return nil
}

func decodeFromByte(raw []byte) ([]byte, error) {
	pad := 8 - (len(raw) % 8)
	nb := make([]byte, len(raw)+pad)
	copy(nb, raw)
	for i := 0; i < pad; i++ {
		nb[len(raw)+i] = '='
	}

	out, err := lowerBase32.DecodeString(string(nb))
	if err != nil {
		return nil, err
	}
	if len(out) != 18 {
		return nil, errors.New("Incorrect length for id")
	}

	return out, nil
}
