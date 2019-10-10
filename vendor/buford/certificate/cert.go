// Package certificate loads Push Services certificates exported from your
// Keychain in Personal Information Exchange format (*.p12).
//
// If you prefer to use *.PEM files, you can of course use tls.LoadX509KeyPair
// or tls.X509KeyPair from the standard library.
package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/pkcs12"
)

// Certificate errors
var (
	ErrExpired = errors.New("certificate has expired or is not yet valid")
)

// Load a .p12 certificate from disk.
func Load(filename, password string) (tls.Certificate, error) {
	p12, err := ioutil.ReadFile(filename)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("Unable to load %s: %v", filename, err)
	}
	return Decode(p12, password)
}

// Decode and verify an in memory .p12 certificate (DER binary format).
func Decode(p12 []byte, password string) (tls.Certificate, error) {
	// decode an x509.Certificate to verify
	privateKey, cert, err := pkcs12.Decode(p12, password)
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := verify(cert); err != nil {
		return tls.Certificate{}, err
	}

	// wraps x509 certificate as a tls.Certificate:
	return tls.Certificate{
		Certificate: [][]byte{cert.Raw},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}, nil
}

// TopicFromCert extracts topic from a certificate's common name.
func TopicFromCert(cert tls.Certificate) string {
	commonName := cert.Leaf.Subject.CommonName

	var topic string
	// Apple Push Services: {bundle}
	// Apple Development IOS Push Services: {bundle}
	n := strings.Index(commonName, ":")
	if n != -1 {
		topic = strings.TrimSpace(commonName[n+1:])
	}
	return topic
}

// verify checks if a certificate has expired
func verify(cert *x509.Certificate) error {
	_, err := cert.Verify(x509.VerifyOptions{})
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case x509.CertificateInvalidError:
		switch e.Reason {
		case x509.Expired:
			return ErrExpired
		case x509.IncompatibleUsage:
			// Apple cert fail on go 1.10
			return nil
		default:
			return err
		}
	case x509.UnknownAuthorityError:
		// Apple cert isn't in the cert pool
		// ignoring this error
		return nil
	default:
		return err
	}
}
