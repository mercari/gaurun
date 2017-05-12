package gcm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	// FCMSendEndpoint is the endpoint for sending message to the Firebase Cloud Messaging (FCM) server.
	// See more on https://firebase.google.com/docs/cloud-messaging/server
	FCMSendEndpoint = "https://fcm.googleapis.com/fcm/send"

	// GCMSendEndpoint is the endpoint for sending messages to the Google Cloud Messaging (GCM) server.
	// Firebase Cloud Messaging (FCM) is the new version of GCM. Should use new endpoint.
	// See more on https://firebase.google.com/support/faq/#gcm-fcm
	GCMSendEndpoint = "https://gcm-http.googleapis.com/gcm/send"
)

const (
	// Initial delay before first retry, without jitter.
	backoffInitialDelay = 1000

	// Maximum delay before a retry.
	maxBackoffDelay = 1024000

	// maxRegistrationIDs are max number of registration IDs in one message.
	maxRegistrationIDs = 1000

	// maxTimeToLive is max time GCM storage can store messages when the device is offline
	maxTimeToLive = 2419200 // 4 weeks
)

// Client abstracts the interaction between the application server and the
// GCM server. The developer must obtain an API key from the Google APIs
// Console page and pass it to the Client so that it can perform authorized
// requests on the application server's behalf. To send a message to one or
// more devices use the Client's Send or SendNoRetry methods.
type Client struct {
	ApiKey string
	URL    string
	Http   *http.Client
}

// NewClient returns a new sender with the given URL and apiKey.
// If one of input is empty or URL is malformed, returns error.
// It sets http.DefaultHTTP client for http connection to server.
// If you need our own configuration overwrite it.
func NewClient(urlString, apiKey string) (*Client, error) {
	if len(urlString) == 0 {
		return nil, fmt.Errorf("missing GCM/FCM endpoint url")
	}

	if len(apiKey) == 0 {
		return nil, fmt.Errorf("missing API Key")
	}

	if _, err := url.Parse(urlString); err != nil {
		return nil, fmt.Errorf("failed to parse URL %q: %s", urlString, err)
	}

	return &Client{
		URL:    urlString,
		ApiKey: apiKey,
		Http:   http.DefaultClient,
	}, nil
}

// SendNoRetry sends a message to the GCM server without retrying in case of
// service unavailability. A non-nil error is returned if a non-recoverable
// error occurs (i.e. if the response status is not "200 OK").
func (c *Client) SendNoRetry(msg *Message) (*Response, error) {
	if err := msg.validate(); err != nil {
		return nil, err
	}

	return c.send(msg)
}

// Send sends a message to the GCM server, retrying in case of service
// unavailability. A non-nil error is returned if a non-recoverable
// error occurs (i.e. if the response status is not "200 OK").
//
// Note that messages are retried using exponential backoff, and as a
// result, this method may block for several seconds.
func (c *Client) Send(msg *Message, retries int) (*Response, error) {
	if err := msg.validate(); err != nil {
		return nil, err
	}

	if retries < 0 {
		return nil, errors.New("'retries' must not be negative.")
	}

	// Send the message for the first time.
	resp, err := c.send(msg)
	if err != nil {
		return nil, err
	} else if resp.Failure == 0 || retries == 0 {
		return resp, nil
	}

	// One or more messages failed to send.
	regIDs := msg.RegistrationIDs
	allResults := make(map[string]Result, len(regIDs))
	backoff := backoffInitialDelay
	for i := 0; updateStatus(msg, resp, allResults) > 0 && i < retries; i++ {
		sleepTime := backoff/2 + rand.Intn(backoff)
		time.Sleep(time.Duration(sleepTime) * time.Millisecond)
		backoff = min(2*backoff, maxBackoffDelay)
		if resp, err = c.send(msg); err != nil {
			msg.RegistrationIDs = regIDs
			return nil, err
		}
	}

	// Bring the message back to its original state.
	msg.RegistrationIDs = regIDs

	// Create a Response containing the overall results.
	finalResults := make([]Result, len(regIDs))
	var success, failure, canonicalIDs int
	for i := 0; i < len(regIDs); i++ {
		result, _ := allResults[regIDs[i]]
		finalResults[i] = result
		if result.MessageID != "" {
			if result.RegistrationID != "" {
				canonicalIDs++
			}
			success++
		} else {
			failure++
		}
	}

	return &Response{
		// Return the most recent multicast id.
		MulticastID:  resp.MulticastID,
		Success:      success,
		Failure:      failure,
		CanonicalIDs: canonicalIDs,
		Results:      finalResults,
	}, nil
}

func (c *Client) send(msg *Message) (*Response, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(msg); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.URL, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("key=%s", c.ApiKey))
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code %d: %s", resp.StatusCode, resp.Status)
	}

	var response Response
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	return &response, err
}

// updateStatus updates the status of the messages sent to devices and
// returns the number of recoverable errors that could be retried.
func updateStatus(msg *Message, resp *Response, allResults map[string]Result) int {
	unsentRegIDs := make([]string, 0, resp.Failure)
	for i := 0; i < len(resp.Results); i++ {
		regID := msg.RegistrationIDs[i]
		allResults[regID] = resp.Results[i]
		if resp.Results[i].Error == "Unavailable" {
			unsentRegIDs = append(unsentRegIDs, regID)
		}
	}
	msg.RegistrationIDs = unsentRegIDs
	return len(unsentRegIDs)
}

// min returns the smaller of two integers. For exciting religious wars
// about why this wasn't included in the "math" package, see this thread:
// https://groups.google.com/d/topic/golang-nuts/dbyqx_LGUxM/discussion
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
