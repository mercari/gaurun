package payload

import (
	"encoding/json"

	"github.com/RobotsAndPencils/buford/payload/badge"
)

// APS is Apple's reserved namespace.
// Use it for payloads destined to mobile devices (iOS).
type APS struct {
	// Alert dictionary.
	Alert Alert

	// Badge to display on the app icon.
	// Set to badge.Preserve (default), badge.Clear
	// or a specific value with badge.New(n).
	Badge badge.Badge

	// The name of a sound file to play as an alert.
	Sound string

	// Content available is for silent notifications
	// with no alert, sound, or badge.
	ContentAvailable bool

	// Category identifier for custom actions in iOS 8 or newer.
	Category string

	// Mutable is used for Service Extensions introduced in iOS 10.
	MutableContent bool

	// Thread identifier to create notification groups in iOS 12 or newer.
	ThreadID string
}

// Alert dictionary.
type Alert struct {
	// Title is a short string shown briefly on Apple Watch in iOS 8.2 or newer.
	Title        string   `json:"title,omitempty"`
	TitleLocKey  string   `json:"title-loc-key,omitempty"`
	TitleLocArgs []string `json:"title-loc-args,omitempty"`

	// Subtitle added in iOS 10
	Subtitle string `json:"subtitle,omitempty"`

	// Body text of the alert message.
	Body    string   `json:"body,omitempty"`
	LocKey  string   `json:"loc-key,omitempty"`
	LocArgs []string `json:"loc-args,omitempty"`

	// Key for localized string for "View" button.
	ActionLocKey string `json:"action-loc-key,omitempty"`

	// Image file to be used when user taps or slides the action button.
	LaunchImage string `json:"launch-image,omitempty"`
}

// isSimple alert with only Body set.
func (a *Alert) isSimple() bool {
	return len(a.Title) == 0 && len(a.Subtitle) == 0 &&
		len(a.LaunchImage) == 0 &&
		len(a.TitleLocKey) == 0 && len(a.TitleLocArgs) == 0 &&
		len(a.LocKey) == 0 && len(a.LocArgs) == 0 && len(a.ActionLocKey) == 0
}

// isZero if no Alert fields are set.
func (a *Alert) isZero() bool {
	return len(a.Body) == 0 && a.isSimple()
}

// Map returns the payload as a map that you can customize
// before serializing it to JSON.
func (a *APS) Map() map[string]interface{} {
	aps := make(map[string]interface{}, 5)

	if !a.Alert.isZero() {
		if a.Alert.isSimple() {
			aps["alert"] = a.Alert.Body
		} else {
			aps["alert"] = a.Alert
		}
	}
	if n, ok := a.Badge.Number(); ok {
		aps["badge"] = n
	}
	if a.Sound != "" {
		aps["sound"] = a.Sound
	}
	if a.ContentAvailable {
		aps["content-available"] = 1
	}
	if a.Category != "" {
		aps["category"] = a.Category
	}
	if a.MutableContent {
		aps["mutable-content"] = 1
	}
	if a.ThreadID != "" {
		aps["thread-id"] = a.ThreadID
	}

	// wrap in "aps" to form the final payload
	return map[string]interface{}{"aps": aps}
}

// MarshalJSON allows you to json.Marshal(aps) directly.
func (a APS) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Map())
}

// Validate that a payload has the correct fields.
func (a *APS) Validate() error {
	if a == nil {
		return ErrIncomplete
	}

	// must have a body or a badge (or custom data)
	if len(a.Alert.Body) == 0 && a.Badge == badge.Preserve {
		return ErrIncomplete
	}
	return nil
}
