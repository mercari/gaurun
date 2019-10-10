package pushpackage

import (
	"bytes"
	"strings"
	"testing"
)

func TestChecksum(t *testing.T) {
	content := `{"websiteName": "Bay Airlines"}`
	r := strings.NewReader(content)

	// confirmed matches openssl sha1 <filename>
	expected := "82c436ae8f6702859cd4dd4c5461c71b77a586a3"

	buf := new(bytes.Buffer)
	c, err := copyAndChecksum(buf, r)
	if err != nil {
		t.Fatal(err)
	}
	if c != expected {
		t.Errorf("Expected checksum %q, got %q", expected, c)
	}
	if buf.String() != content {
		t.Errorf("Expected content %q, got %q", content, buf.String())
	}
}
