package gcm

// Response represents the FCM server's response to the application
// server's sent message. See the documentation for FCM Architectural
// Overview for more information:
// https://firebase.google.com/docs/cloud-messaging/http-server-ref
type Response struct {
	MulticastID  int64    `json:"multicast_id"`
	CanonicalIDs int      `json:"canonical_ids"`
	Results      []Result `json:"results"`
}

// Result represents the status of a processed message.
type Result struct {
	MessageID      string `json:"message_id"`
	RegistrationID string `json:"registration_id"`
	Error          string `json:"error"`
}
