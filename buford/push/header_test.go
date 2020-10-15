package push

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mercari/gaurun/buford/token"
)

func TestHeaders(t *testing.T) {
	headers := Headers{
		ID:          "uuid",
		CollapseID:  "game1.score.identifier",
		Expiration:  time.Unix(12622780800, 0),
		LowPriority: true,
		Topic:       "bundle-id",
		PushType:    PushTypeAlert,
	}

	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "uuid")
	testHeader(t, reqHeader, "apns-collapse-id", "game1.score.identifier")
	testHeader(t, reqHeader, "apns-expiration", "12622780800")
	testHeader(t, reqHeader, "apns-priority", "5")
	testHeader(t, reqHeader, "apns-topic", "bundle-id")
	testHeader(t, reqHeader, "apns-push-type", "alert")
	testHeader(t, reqHeader, "authorization", "")
}

func TestHeadersAuthToken(t *testing.T) {
	ak, err := token.AuthKeyFromFile("testdata/authkey-valid.p8")
	if err != nil {
		t.Fatal(err)
	}

	headers := Headers{
		ID:          "uuid",
		CollapseID:  "game1.score.identifier",
		Expiration:  time.Unix(12622780800, 0),
		LowPriority: true,
		Topic:       "bundle-id",
		PushType:    PushTypeAlert,
		AuthToken: &token.Token{
			AuthKey: ak,
			KeyID:   "key_id",
			TeamID:  "team_id",
		},
	}

	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "uuid")
	testHeader(t, reqHeader, "apns-collapse-id", "game1.score.identifier")
	testHeader(t, reqHeader, "apns-expiration", "12622780800")
	testHeader(t, reqHeader, "apns-priority", "5")
	testHeader(t, reqHeader, "apns-topic", "bundle-id")
	testHeader(t, reqHeader, "apns-push-type", "alert")

	actual := reqHeader.Get("authorization")
	if !strings.HasPrefix(actual, "bearer ") {
		t.Errorf("expected authorization header is the beginning of `beaer`, but got %s", actual)
	}
}

func TestNilHeader(t *testing.T) {
	var headers *Headers
	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "")
	testHeader(t, reqHeader, "apns-collapse-id", "")
	testHeader(t, reqHeader, "apns-expiration", "")
	testHeader(t, reqHeader, "apns-priority", "")
	testHeader(t, reqHeader, "apns-topic", "")
	testHeader(t, reqHeader, "apns-push-type", "")
	testHeader(t, reqHeader, "authorization", "")
}

func TestEmptyHeaders(t *testing.T) {
	headers := Headers{}
	reqHeader := http.Header{}
	headers.set(reqHeader)

	testHeader(t, reqHeader, "apns-id", "")
	testHeader(t, reqHeader, "apns-collapse-id", "")
	testHeader(t, reqHeader, "apns-expiration", "")
	testHeader(t, reqHeader, "apns-priority", "")
	testHeader(t, reqHeader, "apns-topic", "")
	testHeader(t, reqHeader, "apns-push-type", "")
	testHeader(t, reqHeader, "authorization", "")
}

func testHeader(t *testing.T, reqHeader http.Header, key, expected string) {
	actual := reqHeader.Get(key)
	if actual != expected {
		t.Errorf("Expected %s %q, got %q.", key, expected, actual)
	}
}
