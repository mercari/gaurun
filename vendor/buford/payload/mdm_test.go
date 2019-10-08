package payload_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
)

func TestMDM(t *testing.T) {
	p := payload.MDM{Token: "00000000-1111-3333-4444-555555555555"}
	expected := []byte(`{"mdm":"00000000-1111-3333-4444-555555555555"}`)
	testPayload(t, p, expected)
}

func TestValidMDM(t *testing.T) {
	p := payload.MDM{Token: "00000000-1111-3333-4444-555555555555"}
	if err := p.Validate(); err != nil {
		t.Errorf("Expected no error, got %v.", err)
	}
}

func TestInvalidMDM(t *testing.T) {
	tests := []*payload.MDM{
		{},
		nil,
	}

	for _, p := range tests {
		if err := p.Validate(); err != payload.ErrIncomplete {
			t.Errorf("Expected err %v, got %v.", payload.ErrIncomplete, err)
		}
	}
}
