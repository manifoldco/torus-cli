package utils

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

// CreateHTTPTransport creates and configures the
func CreateHTTPTransport(caBundle *x509.CertPool, hostname string) *http.Transport {
	return &http.Transport{TLSClientConfig: &tls.Config{
		ServerName: hostname,
		RootCAs:    caBundle,
	}}
}
