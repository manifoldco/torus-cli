package apitypes

import (
	"errors"

	"github.com/dchest/blake2b"

	"github.com/arigatomachine/cli/base32"
)

// WorklogType is the enumerated byte type of WorklogItems
type WorklogType byte

// The enumberated byte types of WorklogItems
const (
	SecretRotateWorklogType WorklogType = 1 << iota
)

// WorklogResult result states.
const (
	SuccessWorklogResult = "success"
	FailureWorklogResult = "failure"
	ErrorWorklogResult   = "error"
	ManualWorklogResult  = "manual"
)

// ErrIncorrectWorklogIDLen is returned when a base32 encoded worklog id is the
// wrong length.
var ErrIncorrectWorklogIDLen = errors.New("Incorrect worklog ID length")

const worklogIDLen = 9

// WorklogID is the unique content-based identifier for worklog entries
type WorklogID [worklogIDLen]byte

// DecodeWorklogIDFromString decodes a WorklogID from the given base32 encoded
// representation.
func DecodeWorklogIDFromString(raw string) (WorklogID, error) {
	id := WorklogID{}

	buf, err := base32.DecodeString(raw)
	if err != nil {
		return id, err
	}

	if len(buf) != worklogIDLen {
		return id, ErrIncorrectWorklogIDLen
	}

	copy(id[:], buf)
	return id, nil

}

func (id WorklogID) String() string {
	return base32.EncodeToString(id[:])
}

// WorklogItem is an item that the daemon has identified as needing to be done
// to ensure system correctness, or security in the face of stale secrets.
type WorklogItem struct {
	ID      *WorklogID `json:"id"`
	Subject string     `json:"subject"`
	Summary string     `json:"summary"`
}

// Type returns this item's type
func (w *WorklogItem) Type() WorklogType {
	return WorklogType(w.ID[0])
}

// String returns a human reable string for this worklog item type.
func (t WorklogType) String() string {
	switch t {
	case SecretRotateWorklogType:
		return "secret"
	default:
		return "n/a"
	}
}

// CreateID creates and populates a WorklogID for the WorklogItem based on the
// given type and its subject.
func (w *WorklogItem) CreateID(worklogType WorklogType) {
	h, err := blake2b.New(&blake2b.Config{Size: worklogIDLen - 1})
	if err != nil { // this only happens with a bad config
		panic(err)
	}

	h.Write([]byte{byte(worklogType)})
	h.Write([]byte(w.Subject))

	id := WorklogID{byte(worklogType)}
	copy(id[1:], h.Sum(nil))
	w.ID = &id
}

// WorklogResult is the result, either positive or negative, of attempting to
// resolve a WorklogItem
type WorklogResult struct {
	ID      *WorklogID `json:"id"`
	State   string     `json:"state"`
	Message string     `json:"message"`
}
