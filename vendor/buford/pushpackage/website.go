package pushpackage

// Website JSON for creating a push package.
type Website struct {
	// Website Name shown in the Notification Center.
	Name string `json:"websiteName"`

	// Website Push ID (eg. web.com.domain)
	PushID string `json:"websitePushID"`

	// Websites that can request permission from the user.
	AllowedDomains []string `json:"allowedDomains"`

	// http(s) URL for clicked notifications with %@ placeholders.
	URLFormatString string `json:"urlFormatString"`

	// A 16+ character string to identify the user.
	AuthenticationToken string `json:"authenticationToken"`

	// Location of your web service. Must be HTTPS.
	// Don't include a trailing slash.
	WebServiceURL string `json:"webServiceURL"`
}
