package primitive

// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

import "encoding/json"

// MarshalJSON implements the json.Marshaler interface for BaseKeyring.
func (t *BaseKeyring) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"pathexp":`)...)
	ob, err = json.Marshal(t.PathExp)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"project_id":`)...)
	ob, err = json.Marshal(t.ProjectID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"version":`)...)
	ob, err = json.Marshal(t.KeyringVersion)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of User.
func (t *User) Type() byte {
	return 0x01
}

// Type returns the enumerated byte representation of UserV1.
func (t *UserV1) Type() byte {
	return 0x01
}

// Type returns the enumerated byte representation of Service.
func (t *Service) Type() byte {
	return 0x03
}

// Type returns the enumerated byte representation of Project.
func (t *Project) Type() byte {
	return 0x04
}

// Type returns the enumerated byte representation of Environment.
func (t *Environment) Type() byte {
	return 0x05
}

// MarshalJSON implements the json.Marshaler interface for PublicKey.
func (t *PublicKey) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"alg":`)...)
	ob, err = json.Marshal(t.Algorithm)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"expires_at":`)...)
	ob, err = json.Marshal(t.Expires)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"key":`)...)
	ob, err = json.Marshal(t.Key)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"type":`)...)
	ob, err = json.Marshal(t.KeyType)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of PublicKey.
func (t *PublicKey) Type() byte {
	return 0x06
}

// MarshalJSON implements the json.Marshaler interface for PrivateKey.
func (t *PrivateKey) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"key":`)...)
	ob, err = json.Marshal(t.Key)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"pnonce":`)...)
	ob, err = json.Marshal(t.PNonce)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"public_key_id":`)...)
	ob, err = json.Marshal(t.PublicKeyID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of PrivateKey.
func (t *PrivateKey) Type() byte {
	return 0x07
}

// MarshalJSON implements the json.Marshaler interface for Claim.
func (t *Claim) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"public_key_id":`)...)
	ob, err = json.Marshal(t.PublicKeyID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"type":`)...)
	ob, err = json.Marshal(t.ClaimType)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of Claim.
func (t *Claim) Type() byte {
	return 0x08
}

// MarshalJSON implements the json.Marshaler interface for Keyring.
func (t *Keyring) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"pathexp":`)...)
	ob, err = json.Marshal(t.PathExp)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"project_id":`)...)
	ob, err = json.Marshal(t.ProjectID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"version":`)...)
	ob, err = json.Marshal(t.KeyringVersion)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of Keyring.
func (t *Keyring) Type() byte {
	return 0x09
}

// MarshalJSON implements the json.Marshaler interface for KeyringV1.
func (t *KeyringV1) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"pathexp":`)...)
	ob, err = json.Marshal(t.PathExp)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"project_id":`)...)
	ob, err = json.Marshal(t.ProjectID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"version":`)...)
	ob, err = json.Marshal(t.KeyringVersion)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of KeyringV1.
func (t *KeyringV1) Type() byte {
	return 0x09
}

// MarshalJSON implements the json.Marshaler interface for KeyringMember.
func (t *KeyringMember) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"encrypting_key_id":`)...)
	ob, err = json.Marshal(t.EncryptingKeyID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_id":`)...)
	ob, err = json.Marshal(t.KeyringID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"public_key_id":`)...)
	ob, err = json.Marshal(t.PublicKeyID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of KeyringMember.
func (t *KeyringMember) Type() byte {
	return 0x0a
}

// MarshalJSON implements the json.Marshaler interface for KeyringMemberV1.
func (t *KeyringMemberV1) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"encrypting_key_id":`)...)
	ob, err = json.Marshal(t.EncryptingKeyID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"key":`)...)
	ob, err = json.Marshal(t.Key)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_id":`)...)
	ob, err = json.Marshal(t.KeyringID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"project_id":`)...)
	ob, err = json.Marshal(t.ProjectID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"public_key_id":`)...)
	ob, err = json.Marshal(t.PublicKeyID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of KeyringMemberV1.
func (t *KeyringMemberV1) Type() byte {
	return 0x0a
}

// MarshalJSON implements the json.Marshaler interface for Credential.
func (t *Credential) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"credential":`)...)
	ob, err = json.Marshal(t.Credential)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_id":`)...)
	ob, err = json.Marshal(t.KeyringID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"name":`)...)
	ob, err = json.Marshal(t.Name)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"nonce":`)...)
	ob, err = json.Marshal(t.Nonce)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"pathexp":`)...)
	ob, err = json.Marshal(t.PathExp)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"project_id":`)...)
	ob, err = json.Marshal(t.ProjectID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"state":`)...)
	ob, err = json.Marshal(t.State)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"version":`)...)
	ob, err = json.Marshal(t.CredentialVersion)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of Credential.
func (t *Credential) Type() byte {
	return 0x0b
}

// MarshalJSON implements the json.Marshaler interface for CredentialV1.
func (t *CredentialV1) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"credential":`)...)
	ob, err = json.Marshal(t.Credential)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_id":`)...)
	ob, err = json.Marshal(t.KeyringID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"name":`)...)
	ob, err = json.Marshal(t.Name)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"nonce":`)...)
	ob, err = json.Marshal(t.Nonce)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"pathexp":`)...)
	ob, err = json.Marshal(t.PathExp)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"project_id":`)...)
	ob, err = json.Marshal(t.ProjectID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"version":`)...)
	ob, err = json.Marshal(t.CredentialVersion)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of CredentialV1.
func (t *CredentialV1) Type() byte {
	return 0x0b
}

// Type returns the enumerated byte representation of Org.
func (t *Org) Type() byte {
	return 0x0d
}

// Type returns the enumerated byte representation of Membership.
func (t *Membership) Type() byte {
	return 0x0e
}

// Type returns the enumerated byte representation of Team.
func (t *Team) Type() byte {
	return 0x0f
}

// Type returns the enumerated byte representation of Token.
func (t *Token) Type() byte {
	return 0x10
}

// Type returns the enumerated byte representation of Policy.
func (t *Policy) Type() byte {
	return 0x11
}

// Type returns the enumerated byte representation of PolicyAttachment.
func (t *PolicyAttachment) Type() byte {
	return 0x12
}

// Type returns the enumerated byte representation of OrgInvite.
func (t *OrgInvite) Type() byte {
	return 0x13
}

// MarshalJSON implements the json.Marshaler interface for KeyringMemberClaim.
func (t *KeyringMemberClaim) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_id":`)...)
	ob, err = json.Marshal(t.KeyringID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_member_id":`)...)
	ob, err = json.Marshal(t.KeyringMemberID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"previous":`)...)
	ob, err = json.Marshal(t.Previous)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"reason":`)...)
	ob, err = json.Marshal(t.Reason)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"type":`)...)
	ob, err = json.Marshal(t.ClaimType)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of KeyringMemberClaim.
func (t *KeyringMemberClaim) Type() byte {
	return 0x15
}

// MarshalJSON implements the json.Marshaler interface for MEKShare.
func (t *MEKShare) MarshalJSON() ([]byte, error) {
	var ob []byte
	var err error
	b := []byte{'{'}

	b = append(b, []byte(`"created_at":`)...)
	ob, err = json.Marshal(t.Created)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"key":`)...)
	ob, err = json.Marshal(t.Key)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_id":`)...)
	ob, err = json.Marshal(t.KeyringID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"keyring_member_id":`)...)
	ob, err = json.Marshal(t.KeyringMemberID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"org_id":`)...)
	ob, err = json.Marshal(t.OrgID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, ',')
	b = append(b, []byte(`"owner_id":`)...)
	ob, err = json.Marshal(t.OwnerID)
	if err != nil {
		return nil, err
	}
	b = append(b, ob...)

	b = append(b, '}')

	return b, nil
}

// Type returns the enumerated byte representation of MEKShare.
func (t *MEKShare) Type() byte {
	return 0x16
}

// Type returns the enumerated byte representation of Machine.
func (t *Machine) Type() byte {
	return 0x17
}

// Type returns the enumerated byte representation of MachineToken.
func (t *MachineToken) Type() byte {
	return 0x18
}
