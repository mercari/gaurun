package payload_test

import (
	"encoding/json"
	"reflect"
	"testing"
)

func testPayload(t *testing.T, p interface{}, expected []byte) {
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	if !reflect.DeepEqual(b, expected) {
		t.Errorf("Expected %s, got %s", expected, b)
	}
}
