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

func startTestServer(t *testing.T, resp *testResponse) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		status := resp.StatusCode
		if status == 0 || status == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			respBytes, _ := json.Marshal(resp.Response)
			fmt.Fprint(w, string(respBytes))
		} else {
			w.WriteHeader(status)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	return server
}

func TestNewClient(t *testing.T) {
	if _, err := NewClient("", ""); err == nil {
		t.Fatalf("expect to be faied (missing GCM/FCM endpoint)")
	}

	if _, err := NewClient(FCMSendEndpoint, ""); err == nil {
		t.Fatalf("expect to be faied (missing API Key)")
	}
}

func TestSend(t *testing.T) {
	cases := []struct {
		serverResponse *testResponse
		success        bool
	}{
		{
			&testResponse{
				Response: &Response{},
			},
			true,
		},

		{
			&testResponse{
				StatusCode: http.StatusBadRequest,
			},
			false,
		},

		{
			&testResponse{
				Response: &Response{
					Results: []Result{
						{
							Error: "Unavailable",
						},
					},
				},
			},
			true,
		},

		{
			&testResponse{
				Response: &Response{
					Results: []Result{
						{
							MessageID: "id",
						},
					},
				},
			},
			true,
		},
	}

	for i, tc := range cases {
		server := startTestServer(t, tc.serverResponse)
		sender, err := NewClient(server.URL, "testAPIKey")
		if err != nil {
			t.Fatalf("Failed to setup sender client: %s", err)
		}

		msg := NewMessage(map[string]interface{}{"key": "value"}, "1")
		_, err = sender.Send(msg)

		if err != nil {
			if tc.success {
				t.Fatalf("#%d expect to be success: %v", i, err)
			}

			server.Close()
			continue
		}

		if !tc.success {
			t.Fatalf("#%d expect to be failed", i)
		}

		server.Close()
	}
}
