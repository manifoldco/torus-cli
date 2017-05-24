package aws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/manifoldco/torus-cli/gatekeeper/apitypes"
	"github.com/manifoldco/torus-cli/gatekeeper/client"
)

const (
	// MetadataURL is the metadata URL for AWS
	MetadataURL = "http://169.254.169.254/latest/"
)

// Bootstrap bootstraps a the AWS instance into a role to a given Gatekeeper instance
func Bootstrap(url, name, org, role string) (*apitypes.BootstrapResponse, error) {
	var err error
	client := client.NewClient(url)

	identity, err := metadata()
	if err != nil {
		return nil, err
	}

	sig, err := signature()
	if err != nil {
		return nil, err
	}

	v := Verifier{
		Identity:  identity,
		Signature: sig,
	}

	if err := v.Verify(); err != nil {
		return nil, err
	}

	var identityDoc identityDocument
	json.Unmarshal(identity, &identityDoc)

	if name == "" {
		name = identityDoc.InstanceID
	}

	bootreq := apitypes.AWSBootstrapRequest{
		Identity:      identity,
		Signature:     sig,
		ProvisionTime: identityDoc.PendingTime,

		Machine: apitypes.MachineBootstrap{
			Name: name,
			Org:  org,
			Team: role,
		},
	}

	return client.Bootstrap("aws", bootreq)
}

func metadata() ([]byte, error) {
	identity, err := fetchMetadata("dynamic/instance-identity/document")
	if err != nil {
		return nil, err
	}

	return identity, nil
}

func signature() ([]byte, error) {
	signature, err := fetchMetadata("dynamic/instance-identity/pkcs7")
	if err != nil {
		return nil, err
	}
	return signature, nil
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
