package gaurun

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNotification(t *testing.T) {
	cases := []struct {
		Notification RequestGaurunNotification
		Expected     error
	}{
		// positive cases
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 1,
				Message:  "test message",
			},
			nil,
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 2,
				Message:  "test message",
			},
			nil,
		},
		{
			RequestGaurunNotification{
				Tokens:     []string{"test token"},
				Platform:   1,
				Message:    "test message with identifier",
				Identifier: "identifier",
			},
			nil,
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 3,
				Title:    "test title",
				Body:     "test body",
			},
			nil,
		},

		// negative cases
		{
			RequestGaurunNotification{
				Tokens: []string{""},
			},
			errors.New("empty token"),
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 100, /* neither iOS nor Android */
			},
			errors.New("invalid platform"),
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 1,
				Message:  "",
			},
			errors.New("empty message"),
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 3,
				Message:  "abcd",
			},
			errors.New("empty title or body"),
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 3,
				Title:    "abcd",
			},
			errors.New("empty title or body"),
		},
		{
			RequestGaurunNotification{
				Tokens:   []string{"test token"},
				Platform: 3,
				Body:     "abcd",
			},
			errors.New("empty title or body"),
		},
	}

	for _, c := range cases {
		actual := validateNotification(&c.Notification)
		assert.Equal(t, actual, c.Expected)
	}
}

func TestSendResponse(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sendResponse(w, "valid message", http.StatusOK)
		return
	}))
	defer s.Close()

	res, err := http.Get(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, string(body), "{\"message\":\"valid message\"}\n")
}
