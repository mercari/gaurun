// Package pushpackage creates website push packages and wallet pass packages.
package pushpackage

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/aai/gocrypto/pkcs7"
)

// Package for website push package or wallet pass package.
type Package struct {
	z *zip.Writer

	// manifest is a map of relative file paths to their SHA checksums
	manifest map[string]string

	err error
}

// New push package will be written to w.
func New(w io.Writer) Package {
	return Package{
		z:        zip.NewWriter(w),
		manifest: make(map[string]string),
	}
}

// EncodeJSON to the push package.
func (p *Package) EncodeJSON(name string, e interface{}) {
	if p.err != nil {
		return
	}

	b, err := json.Marshal(e)
	if err != nil {
		p.err = err
		return
	}
	r := bytes.NewReader(b)

	p.Copy(name, r)
}

// Copy reader to the push package.
func (p *Package) Copy(name string, r io.Reader) {
	if p.err != nil {
		return
	}

	zf, err := p.z.Create(name)
	if err != nil {
		p.err = err
		return
	}

	checksum, err := copyAndChecksum(zf, r)
	if err != nil {
		p.err = err
		return
	}

	p.manifest[name] = checksum
}

// File writes a file to the push package.
//
// NOTE: Name is a relative path. Only forward slashes are allowed.
func (p *Package) File(name, src string) {
	if p.err != nil {
		return
	}

	f, err := os.Open(src)
	if err != nil {
		p.err = err
		return
	}
	defer f.Close()
	p.Copy(name, f)
}

// Sign the package and close.
// Passbook needs Apple's intermediate WWDR certificate.
func (p *Package) Sign(cert tls.Certificate, wwdr *x509.Certificate) error {
	if p.err != nil {
		return p.err
	}

	// assert that private key is RSA
	key, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return errors.New("expected RSA private key type")
	}

	manifestBytes, err := json.Marshal(p.manifest)
	if err != nil {
		return err
	}

	zf, err := p.z.Create("manifest.json")
	if err != nil {
		return err
	}
	zf.Write(manifestBytes)

	// sign manifest.json with PKCS #7
	// and add signature to the zip file
	zf, err = p.z.Create("signature")
	if err != nil {
		return err
	}
	sig, err := pkcs7.Sign2(bytes.NewReader(manifestBytes), cert.Leaf, key, wwdr)
	if err != nil {
		return err
	}
	zf.Write(sig)

	return p.z.Close()
}

// Error that occurred while adding files to the push package.
func (p *Package) Error() error {
	return p.err
}
