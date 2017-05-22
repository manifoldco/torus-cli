// Package aws provides Gatekeeper authentication for AWS
package aws

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/fullsailor/pkcs7"
	"github.com/manifoldco/torus-cli/data"
	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/gatekeeper/client"
)

const (
	// MetadataURL is the metadata URL for AWS
	MetadataURL = "http://169.254.169.254/latest/"
)

// AWS provides AWS bootstrapping and verification
type AWS struct {
	// Identity is the []bytes of the identity metadata document
	Identity []byte

	// Signature is the []bytes of the identity signature
	Signature []byte

	// VerifyInstance verifies that the instance is running in the account
	// provided by the document in Identity
	VerifyInstance bool
}

// New returns a new AWS bootstrap authentication Provider
func New() *AWS {
	return &AWS{
		VerifyInstance: false,
	}
}

// Bootstrap bootstraps a the AWS instance into a role to a given Gatekeeper instance
func (aws *AWS) Bootstrap(url, name, org, role string) (*apitypes.BootstrapResponse, error) {
	var err error
	client := client.NewClient(url)

	aws.Identity, err = aws.metadata()
	if err != nil {
		return nil, err
	}

	aws.Signature, err = aws.signature()
	if err != nil {
		return nil, err
	}

	if err := aws.Verify(); err != nil {
		return nil, err
	}

	var identityDoc identityDocument
	json.Unmarshal(aws.Identity, &identityDoc)

	if name == "" {
		name = identityDoc.InstanceID
	}

	bootreq := apitypes.AWSBootstrapRequest{
		Identity:      aws.Identity,
		Signature:     aws.Signature,
		ProvisionTime: identityDoc.PendingTime,

		Machine: apitypes.MachineBootstrap{
			Name: name,
			Org:  org,
			Team: role,
		},
	}

	return client.Bootstrap("aws", bootreq)
}

// Verify verifies the instance metadata and isntance identity documents to provide a safe way
// to provision a machine
// TODO: Verify further that the instance actually exists in AWS
func (aws *AWS) Verify() error {
	if aws.Identity == nil {
		var err error
		if aws.Identity, err = aws.metadata(); err != nil {
			return err
		}
	}

	if aws.Signature == nil {
		var err error
		if aws.Signature, err = aws.signature(); err != nil {
			return err
		}
	}

	signature := fmt.Sprintf("-----BEGIN PKCS7-----\n%s\n-----END PKCS7-----", aws.Signature)
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

	return nil
}

func (aws *AWS) metadata() ([]byte, error) {
	var err error
	aws.Identity, err = fetchMetadata("dynamic/instance-identity/document")
	if err != nil {
		return nil, err
	}

	return aws.Identity, nil
}

func (aws *AWS) signature() ([]byte, error) {
	var err error
	aws.Signature, err = fetchMetadata("dynamic/instance-identity/pkcs7")
	if err != nil {
		return nil, err
	}
	return aws.Signature, nil
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

func fetchMetadata(endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s%s", MetadataURL, endpoint)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type identityDocument struct {
	InstanceID  string    `json:"instanceId"`
	PendingTime time.Time `json:"pendingTime"`
}
