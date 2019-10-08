package payload_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/RobotsAndPencils/buford/payload"
)

func ExampleBrowser() {
	p := payload.Browser{
		Alert: payload.BrowserAlert{
			Title:  "Flight A998 Now Boarding",
			Body:   "Boarding has begun for Flight A998.",
			Action: "View",
		},
		URLArgs: []string{"boarding", "A998"},
	}

	b, err := json.Marshal(p)
	if err != nil {
		// handle error
	}
	fmt.Printf("%s", b)
	// Output: {"aps":{"alert":{"title":"Flight A998 Now Boarding","body":"Boarding has begun for Flight A998.","action":"View"},"url-args":["boarding","A998"]}}
}

func TestBrowser(t *testing.T) {
	p := payload.Browser{
		Alert: payload.BrowserAlert{
			Title:  "Flight A998 Now Boarding",
			Body:   "Boarding has begun for Flight A998.",
			Action: "View",
		},
		URLArgs: []string{"boarding", "A998"},
	}
	expected := []byte(`{"aps":{"alert":{"title":"Flight A998 Now Boarding","body":"Boarding has begun for Flight A998.","action":"View"},"url-args":["boarding","A998"]}}`)
	testPayload(t, p, expected)
}

func TestValidBrowser(t *testing.T) {
	p := payload.Browser{
		Alert: payload.BrowserAlert{
			Title: "Flight A998 Now Boarding",
			Body:  "Boarding has begun for Flight A998.",
		},
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Expected no error, got %v.", err)
	}
}

func TestInvalidBrowser(t *testing.T) {
	tests := []*payload.Browser{
		{
			Alert: payload.BrowserAlert{Action: "View"},
		},
		{},
		nil,
	}

	for _, p := range tests {
		if err := p.Validate(); err != payload.ErrIncomplete {
			t.Errorf("Expected err %v, got %v.", payload.ErrIncomplete, err)
		}
	}
}
