package push_test

import (
	"testing"

	"github.com/RobotsAndPencils/buford/push"
)

func TestInvalidDeviceTokens(t *testing.T) {
	tokens := []string{
		"invalid-token",
		"f00f",
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijkl",
		"50c4afb5 9197d2ba d1794be4 e63f2532 ee18c660 0ee655fa 38b0b380 94fd8847",
		"<50c4afb5 9197d2ba d1794be4 e63f2532 ee18c660 0ee655fa 38b0b380 94fd8847>",
	}

	for _, token := range tokens {
		if push.IsDeviceTokenValid(token) {
			t.Errorf("Expected device token %q to be invalid.", token)
		}
	}
}

func TestValidDeviceToken(t *testing.T) {
	tokens := []string{
		"c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433",
		"c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954ac2732227a1d8021cfaf781d71fb2f908c61f5861079a00954ac2732227a1d8021cfaf781d71fb2f908c61f5861079a00954ac2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a",
	}

	for _, token := range tokens {
		if !push.IsDeviceTokenValid(token) {
			t.Errorf("Expected device token %q to be valid.", token)
		}
	}
}
