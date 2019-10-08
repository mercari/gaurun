package push

import "encoding/hex"

// IsDeviceTokenValid checks if s is a hexadecimal token of the correct length.
func IsDeviceTokenValid(s string) bool {
	// TODO: In 2016, they may be growing up to 100 bytes (200 hexadecimal digits).
	if len(s) < 64 || len(s) > 200 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
