package push

import (
	"net/http"
	"strconv"
	"time"
)

// Headers sent with a push to control the notification (optional)
type Headers struct {
	// ID for the notification. Apple generates one if omitted.
	// This should be a UUID with 32 lowercase hexadecimal digits.
	ID string

	// CollapseID is used to update an existing notification that has the same
	// identifier (Notification Management in iOS 10).
	CollapseID string

	// Apple will retry delivery until this time. The default behavior only tries once.
	Expiration time.Time

	// Allow Apple to group messages together to reduce power consumption.
	// By default messages are sent immediately.
	LowPriority bool

	// Topic for certificates with multiple topics.
	Topic string

	PushType PushType
}

type PushType string

const (
	PushTypeAlert      PushType = "alert"
	PushTypeBackground PushType = "background"
)

// set headers for an HTTP request
func (h *Headers) set(reqHeader http.Header) {
	// headers are optional
	if h == nil {
		return
	}

	if h.ID != "" {
		reqHeader.Set("apns-id", h.ID)
	} // when omitted, Apple will generate a UUID for you

	if h.CollapseID != "" {
		reqHeader.Set("apns-collapse-id", h.CollapseID)
	}

	if !h.Expiration.IsZero() {
		reqHeader.Set("apns-expiration", strconv.FormatInt(h.Expiration.Unix(), 10))
	}

	if h.LowPriority {
		reqHeader.Set("apns-priority", "5")
	} // when omitted, the default priority is 10

	if h.Topic != "" {
		reqHeader.Set("apns-topic", h.Topic)
	}

	if h.PushType != "" {
		reqHeader.Set("apns-push-type", string(h.PushType))
	}
}
