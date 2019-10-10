package payload_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
)

func ExampleAPS() {
	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
		Sound: "bingbong.aiff",
	}

	b, err := json.Marshal(p)
	if err != nil {
		// handle error
	}
	fmt.Printf("%s", b)
	// Output: {"aps":{"alert":"Hello HTTP/2","badge":42,"sound":"bingbong.aiff"}}
}

// Use Map to add custom values to the payload.
func ExampleAPS_Map() {
	p := payload.APS{
		Alert: payload.Alert{Body: "Topic secret message"},
	}
	pm := p.Map()
	pm["acme2"] = []string{"bang", "whiz"}

	b, err := json.Marshal(pm)
	if err != nil {
		// handle error
	}
	fmt.Printf("%s", b)
	// Output: {"acme2":["bang","whiz"],"aps":{"alert":"Topic secret message"}}
}

func ExampleAPS_Validate() {
	p := payload.APS{
		Badge: badge.Preserve,
		Sound: "bingbong.aiff",
	}
	if err := p.Validate(); err != nil {
		fmt.Println(err)
	}
	// Output: payload does not contain necessary fields
}

func TestPayload(t *testing.T) {
	var tests = []struct {
		input    payload.APS
		expected []byte
	}{
		{
			payload.APS{
				Alert: payload.Alert{Body: "Message received from Bob"},
			},
			[]byte(`{"aps":{"alert":"Message received from Bob"}}`),
		},
		{
			payload.APS{
				Alert: payload.Alert{Body: "You got your emails."},
				Badge: badge.New(9),
				Sound: "bingbong.aiff",
			},
			[]byte(`{"aps":{"alert":"You got your emails.","badge":9,"sound":"bingbong.aiff"}}`),
		},
		{
			payload.APS{ContentAvailable: true},
			[]byte(`{"aps":{"content-available":1}}`),
		},
		{
			payload.APS{
				Alert: payload.Alert{
					Title:    "Message",
					Subtitle: "This is important",
					Body:     "Message received from Bob",
				},
			},
			[]byte(`{"aps":{"alert":{"title":"Message","subtitle":"This is important","body":"Message received from Bob"}}}`),
		},
		{
			payload.APS{
				Alert:          payload.Alert{Body: "Change is coming"},
				MutableContent: true,
			},
			[]byte(`{"aps":{"alert":"Change is coming","mutable-content":1}}`),
		},
		{
			payload.APS{
				Alert:    payload.Alert{Body: "Grouped notification"},
				ThreadID: "thread-id-1",
			},
			[]byte(`{"aps":{"alert":"Grouped notification","thread-id":"thread-id-1"}}`),
		},
	}

	for _, tt := range tests {
		testPayload(t, tt.input, tt.expected)
	}
}

func TestCustomArray(t *testing.T) {
	p := payload.APS{Alert: payload.Alert{Body: "Message received from Bob"}}
	pm := p.Map()
	pm["acme2"] = []string{"bang", "whiz"}
	expected := []byte(`{"acme2":["bang","whiz"],"aps":{"alert":"Message received from Bob"}}`)
	testPayload(t, pm, expected)
}

func TestValidAPS(t *testing.T) {
	tests := []payload.APS{
		{Alert: payload.Alert{Body: "You got your emails."}},
		{Badge: badge.New(9)},
		{Badge: badge.Clear},
	}

	for _, p := range tests {
		if err := p.Validate(); err != nil {
			t.Errorf("Expected no error, got %v.", err)
		}
	}
}

func TestInvalidAPS(t *testing.T) {
	tests := []*payload.APS{
		{Sound: "bingbong.aiff"},
		{},
		nil,
	}

	for _, p := range tests {
		if err := p.Validate(); err != payload.ErrIncomplete {
			t.Errorf("Expected err %v, got %v.", payload.ErrIncomplete, err)
		}
	}
}
