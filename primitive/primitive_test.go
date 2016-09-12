package primitive

import (
	"encoding/json"
	"testing"
)

const (
	single   = `"create"`
	multiple = `["create","delete"]`
)

func TestPolicyActionMarshalJSON(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		pa := PolicyAction(PolicyActionCreate)

		out, err := json.Marshal(&pa)
		if err != nil {
			t.Error("Error while marshaling:", err)
		}

		if string(out) != single {
			t.Error("Expected:", single, "Got:", string(out))
		}
	})

	t.Run("multiple", func(t *testing.T) {
		pa := PolicyAction(PolicyActionCreate | PolicyActionDelete)

		out, err := json.Marshal(&pa)
		if err != nil {
			t.Error("Error while marshaling:", err)
		}

		if string(out) != multiple {
			t.Error("Expected:", multiple, "Got:", string(out))
		}
	})
}

func TestPolicyActionUnarshalJSON(t *testing.T) {
	type tc struct {
		in  string
		out byte
	}

	testCases := []tc{
		{in: single, out: PolicyActionCreate},
		{in: multiple, out: PolicyActionCreate | PolicyActionDelete},
	}

	for _, test := range testCases {
		t.Run(test.in, func(t *testing.T) {
			var pa PolicyAction

			err := json.Unmarshal([]byte(test.in), &pa)
			if err != nil {
				t.Error("Error while Unmarshaling:", err)
			}

			if byte(pa) != test.out {
				t.Error("Expected:", test.out, "Got:", pa)
			}
		})
	}
}
