package payload

import "encoding/json"

// Browser for Safari Push Notifications.
type Browser struct {
	Alert   BrowserAlert
	URLArgs []string
}

// BrowserAlert for Safari Push Notifications.
type BrowserAlert struct {
	// Title and Body are required
	Title string `json:"title"`
	Body  string `json:"body"`
	// Action button label (defaults to "Show")
	Action string `json:"action,omitempty"`
}

// MarshalJSON allows you to json.Marshal(browser) directly.
func (p Browser) MarshalJSON() ([]byte, error) {
	aps := map[string]interface{}{"alert": p.Alert, "url-args": p.URLArgs}
	return json.Marshal(map[string]interface{}{"aps": aps})
}

// Validate browser payload.
func (p *Browser) Validate() error {
	if p == nil {
		return ErrIncomplete
	}

	// must have both a title and body. action and url-args are optional.
	if len(p.Alert.Title) == 0 || len(p.Alert.Body) == 0 {
		return ErrIncomplete
	}
	return nil
}
