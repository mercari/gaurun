package certificate_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/certificate"
)

func TestValidCert(t *testing.T) {
	const name = "../testdata/cert.p12"

	_, err := certificate.Load(name, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMissingFile(t *testing.T) {
	_, err := certificate.Load("hide-and-seek.p12", "")
	if err == nil {
		t.Fatal("Expected file not found, got", err)
	}
}

func TestTopicFromCert(t *testing.T) {
	const name = "../testdata/cert.p12"

	cert, err := certificate.Load(name, "")
	if err != nil {
		t.Fatal(err)
	}

	// TODO: need a test cert with a CommonName
	const expected = ""
	actual := certificate.TopicFromCert(cert)
	if actual != expected {
		t.Errorf("Expected topic %q, got %q.", expected, actual)
	}
}
