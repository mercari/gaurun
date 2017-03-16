package gcm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testResponse struct {
	StatusCode int
	Response   *Response
}

func startTestServer(t *testing.T, responses []*testResponse) *httptest.Server {
	i := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if i >= len(responses) {
			t.Fatalf("server received %d requests, expected %d", i+1, len(responses))
		}
		resp := responses[i]
		status := resp.StatusCode
		if status == 0 || status == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			respBytes, _ := json.Marshal(resp.Response)
			fmt.Fprint(w, string(respBytes))
		} else {
			w.WriteHeader(status)
		}
		i++
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	return server
}

func TestNewClient(t *testing.T) {
	if _, err := NewClient("", ""); err == nil {
		t.Fatalf("expect to be faied (missing GCM/FCM endpoint)")
	}

	if _, err := NewClient(GCMSendEndpoint, ""); err == nil {
		t.Fatalf("expect to be faied (missing API Key)")
	}
}

func TestSend(t *testing.T) {
	cases := []struct {
		serverResponses []*testResponse
		retry           int
		failure         int
		success         bool
	}{
		{
			[]*testResponse{
				{Response: &Response{}},
			},
			0,
			0,
			true,
		},

		{
			[]*testResponse{
				{StatusCode: http.StatusBadRequest},
			},
			0,
			0,
			false,
		},

		// Should succeed after one retry.
		{
			[]*testResponse{
				{Response: &Response{Failure: 1, Results: []Result{{Error: "Unavailable"}}}},
				{Response: &Response{Success: 1, Results: []Result{{MessageID: "id"}}}},
			},
			1,
			0,
			true,
		},

		// Should return response with one failure.
		{
			[]*testResponse{
				{Response: &Response{Failure: 1, Results: []Result{{Error: "Unavailable"}}}},
				{Response: &Response{Failure: 1, Results: []Result{{Error: "Unavailable"}}}},
			},
			1,
			1,
			true,
		},

		// Should send should fail after one retry.
		{
			[]*testResponse{
				{Response: &Response{Failure: 1, Results: []Result{{Error: "Unavailable"}}}},
				{StatusCode: http.StatusBadRequest},
			},
			1,
			0,
			false,
		},
	}

	for i, tc := range cases {
		server := startTestServer(t, tc.serverResponses)
		sender, err := NewClient(server.URL, "testAPIKey")
		if err != nil {
			t.Fatalf("Failed to setup sender client: %s", err)
		}

		var resp *Response
		msg := NewMessage(map[string]interface{}{"key": "value"}, "1")
		if tc.retry == 0 {
			resp, err = sender.SendNoRetry(msg)
		} else {
			resp, err = sender.Send(msg, tc.retry)
		}

		if err != nil {
			if tc.success {
				t.Fatalf("#%d expect to be success: %s", i, err)
			}

			server.Close()
			continue
		}

		if !tc.success {
			t.Fatalf("#%d expect to be failed", i)
		}

		if resp.Failure != tc.failure {
			t.Fatalf("#%d number of failure %d, want %d", i, resp.Failure, tc.failure)
		}

		server.Close()
	}
}
