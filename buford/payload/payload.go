// Package payload prepares a JSON payload to push.
package payload

import "errors"

// Validation errors.
var (
	ErrIncomplete = errors.New("payload does not contain necessary fields")
)
