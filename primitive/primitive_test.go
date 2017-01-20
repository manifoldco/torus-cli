package primitive

import (
	"encoding/json"
	"reflect"
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

func TestPolicyActionUnmarshalJSON(t *testing.T) {
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

func TestKeyringMemberRevocationTypeMarshalJSON(t *testing.T) {
	type tc struct {
		in  KeyringMemberRevocationType
		out string
	}

	testCases := []tc{
		{in: OrgRemovalRevocationType, out: `"org_removal"`},
		{in: KeyRevocationRevocationType, out: `"key_revocation"`},
		{in: MachineDestroyRevocationType, out: `"machine_destroy"`},
		{in: MachineTokenDestroyRevocationType, out: `"machine_token_destroy"`},
	}

	for _, test := range testCases {
		t.Run(test.in.String(), func(t *testing.T) {
			out, err := json.Marshal(test.in)
			if err != nil {
				t.Fatal("Error while marshaling:", err)
			}

			if string(out) != test.out {
				t.Error("Expected type:", test.out, "Got:", out)
			}
		})
	}
}
func TestKeyringMemberClaimReasonUnarshalJSON(t *testing.T) {
	type tc struct {
		in  string
		out KeyringMemberClaimReason
	}

	testCases := []tc{
		{in: `{"type": "org_removal"}`, out: KeyringMemberClaimReason{
			Type:   OrgRemovalRevocationType,
			Params: nil,
		}},
		{in: `{"type": "key_revocation", "params": {}}`, out: KeyringMemberClaimReason{
			Type:   KeyRevocationRevocationType,
			Params: &KeyRevocationRevocationParams{},
		}},
		{in: `{"type": "machine_destroy", "params": {}}`, out: KeyringMemberClaimReason{
			Type:   MachineDestroyRevocationType,
			Params: &MachineDestroyRevocationParams{},
		}},
		{in: `{"type": "machine_token_destroy", "params": {}}`, out: KeyringMemberClaimReason{
			Type:   MachineTokenDestroyRevocationType,
			Params: &MachineTokenDestroyRevocationParams{},
		}},
	}

	for _, test := range testCases {
		t.Run(test.out.Type.String(), func(t *testing.T) {
			var reason KeyringMemberClaimReason
			err := json.Unmarshal([]byte(test.in), &reason)
			if err != nil {
				t.Fatal("Error while unmarshaling:", err)
			}

			if reason.Type != test.out.Type {
				t.Error("Expected type:", test.out.Type, "Got", reason.Type)
			}

			if reflect.TypeOf(reason.Params) != reflect.TypeOf(test.out.Params) {
				t.Error("param type mismatch")
			}
		})
	}
}
