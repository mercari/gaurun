package gaurun

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type globalsStore struct {
	ConfGaurun        ConfToml
	QueueNotification chan RequestGaurunNotification
}

var (
	store globalsStore
)

func (store globalsStore) save() {
	store.QueueNotification = QueueNotification
	store.ConfGaurun = ConfGaurun
}

func (store globalsStore) restore() {
	QueueNotification = store.QueueNotification
	ConfGaurun = store.ConfGaurun
}

func TestEnqueueNotifications(t *testing.T) {
	store.save()
	defer store.restore()

	positiveCases := []struct {
		N []RequestGaurunNotification
		C ConfToml
	}{
		{
			[]RequestGaurunNotification{
				{
					Tokens:   []string{"test token"},
					Platform: 1,
					Message:  "test message",
				},
				{
					Tokens:   []string{"test token"},
					Platform: 2,
					Message:  "test message",
				},
			},
			ConfToml{Ios: SectionIos{Enabled: true}, Android: SectionAndroid{Enabled: true}},
		},
	}

	negativeCases := []struct {
		N []RequestGaurunNotification
		C ConfToml
	}{
		// push is disabled
		{
			[]RequestGaurunNotification{
				{
					Tokens:   []string{"test token"},
					Platform: 1,
					Message:  "test message",
				},
				{
					Tokens:   []string{"test token"},
					Platform: 2,
					Message:  "test message",
				},
			},
			ConfToml{Ios: SectionIos{Enabled: false}, Android: SectionAndroid{Enabled: false}},
		},

		// config is invalid
		{
			[]RequestGaurunNotification{
				{
					Tokens: []string{""},
				},
			},
			ConfToml{Ios: SectionIos{Enabled: true}, Android: SectionAndroid{Enabled: true}},
		},
		{
			[]RequestGaurunNotification{
				{
					Tokens:   []string{"test token"},
					Platform: 100, /* neither iOS nor Android */
					Message:  "test message",
				},
			},
			ConfToml{Ios: SectionIos{Enabled: true}, Android: SectionAndroid{Enabled: true}},
		},
	}

	for _, c := range positiveCases {
		QueueNotification = make(chan RequestGaurunNotification, len(c.N))
		ConfGaurun = c.C

		enqueueNotifications(c.N)

		assert.Equal(t, len(QueueNotification), len(c.N))
	}

	for _, c := range negativeCases {
		QueueNotification = make(chan RequestGaurunNotification, len(c.N))
		ConfGaurun = c.C

		enqueueNotifications(c.N)

		assert.Equal(t, len(QueueNotification), 0)
	}
}

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

func TestPushNotificationHandler(t *testing.T) {
	store.save()
	defer store.restore()

	cases := []struct {
		Method       string
		RequestBody  string
		Conf         ConfToml
		ExpectedBody string
		ExpectedCode int
	}{
		{
			"GET",
			"",
			ConfToml{},
			"{\"message\":\"method must be POST\"}\n",
			http.StatusBadRequest,
		},
		{
			"POST",
			"",
			ConfToml{},
			"{\"message\":\"request body is empty\"}\n",
			http.StatusBadRequest,
		},
		{
			"POST",
			"invalid json",
			ConfToml{},
			"{\"message\":\"Request-body is malformed\"}\n",
			http.StatusBadRequest,
		},
		/* NOTE It gets "empty notification" actually ...
		   {
		     "POST",
		     "invalid json",
		     ConfToml{Log: SectionLog{Level: "debug"}},
		     "{\"message\":\"Request-body is malformed\"}\n",
		     http.StatusBadRequest,
		   },
		*/
		{
			"POST",
			"{\"notifications\":[]}",
			ConfToml{},
			"{\"message\":\"empty notification\"}\n",
			http.StatusBadRequest,
		},
		{
			"POST",
			"{\"notifications\":[{}]}",
			ConfToml{Core: SectionCore{NotificationMax: 0}},
			"{\"message\":\"number of notifications(1) over limit(0)\"}\n",
			http.StatusBadRequest,
		},
		// NOTE It will cause goroutine leak ...
		{
			"POST",
			"{\"notifications\":[{}]}",
			ConfToml{Core: SectionCore{NotificationMax: 10}},
			"{\"message\":\"ok\"}\n",
			http.StatusOK,
		},
	}

	for _, c := range cases {
		s := httptest.NewServer(http.HandlerFunc(PushNotificationHandler))
		defer s.Close()

		ConfGaurun = c.Conf

		client := http.Client{}
		req, _ := http.NewRequest(c.Method, s.URL, bytes.NewBufferString(c.RequestBody))
		res, err := client.Do(req)

		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, c.ExpectedCode)

		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		assert.Equal(t, string(body), c.ExpectedBody)
	}
}
