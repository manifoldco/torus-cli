package aws

import (
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/fullsailor/pkcs7"

	"github.com/manifoldco/torus-cli/data"
)

const (
	// BootstrapTime indicates the amount of time that a machine
	// is able to register with Torus after being booted
	BootstrapTime = 5 // minutes
)

// Verifier verifies the AWS instance metadata
type Verifier struct {
	// Identity is the []bytes of the identity metadata document
	Identity []byte

	// Signature is the []bytes of the identity signature
	Signature []byte
}

// Verify verifies the instance metadata and instance identity documents to provide a safe way
// to provision a machine
// TODO: Verify further that the instance actually exists in AWS
func (v *Verifier) Verify() error {
	if v.Identity == nil {
		return fmt.Errorf("no identity document to verify")
	}

	if v.Signature == nil {
		return fmt.Errorf("no signature found for verification")
	}

	signature := fmt.Sprintf("-----BEGIN PKCS7-----\n%s\n-----END PKCS7-----", v.Signature)
	block, rest := pem.Decode([]byte(signature))
	if len(rest) != 0 {
		return fmt.Errorf("Unable to decode signature")
	}

	sigData, err := pkcs7.Parse(block.Bytes)
	if err != nil {
		return fmt.Errorf("Unable to parse signature")
	}

	awsCert, err := awsPublicCert()
	if err != nil {
		return nil
	}

	sigData.Certificates = []*x509.Certificate{awsCert}

	if err := sigData.Verify(); err != nil {
		return fmt.Errorf("failed to verify")
	}

	if subtle.ConstantTimeCompare(v.Identity, sigData.Content) != 1 {
		return fmt.Errorf("failed to validate instance metadata")
	}

	var signedIdentityDoc identityDocument
	json.Unmarshal(sigData.Content, &signedIdentityDoc)
	elapsed := time.Since(signedIdentityDoc.PendingTime)
	if elapsed.Minutes() > BootstrapTime {
		return fmt.Errorf(
			"failed validation - this server has beens started more than %d minutes ago",
			BootstrapTime,
		)
	}

	return nil
}

func awsPublicCert() (*x509.Certificate, error) {
	awsCert, err := data.Asset("data/aws_identity_cert.pem")
	if err != nil {
		return nil, fmt.Errorf("unable to find AWS Public certificate")
	}

	block, rest := pem.Decode(awsCert)
	if len(rest) != 0 {
		return nil, fmt.Errorf("unable to decode AWS Public certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		return nil, errors.New("invalid cerficicate")
	}

	return cert, nil
}
