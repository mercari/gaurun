// Package push sends notifications over HTTP/2 to
// Apple's Push Notification Service.
package push

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

// Apple host locations for configuring Service.
const (
	Development     = "https://api.development.push.apple.com"
	Development2197 = "https://api.development.push.apple.com:2197"
	Production      = "https://api.push.apple.com"
	Production2197  = "https://api.push.apple.com:2197"
)

const maxPayload = 4096 // 4KB at most

// Service is the Apple Push Notification Service that you send notifications to.
type Service struct {
	Host   string
	Client *http.Client
}

// NewService creates a new service to connect to APN.
func NewService(client *http.Client, host string) *Service {
	return &Service{
		Client: client,
		Host:   host,
	}
}

// NewClient sets up an HTTP/2 client for a certificate.
func NewClient(cert tls.Certificate) (*http.Client, error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: config}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	return &http.Client{Transport: transport}, nil
}

// Push sends a notification and waits for a response.
func (s *Service) Push(deviceToken string, headers *Headers, payload []byte) (string, error) {
	// check payload length before even hitting Apple.
	if len(payload) > maxPayload {
		return "", &Error{
			Reason: ErrPayloadTooLarge,
			Status: http.StatusRequestEntityTooLarge,
		}
	}

	urlStr := fmt.Sprintf("%v/3/device/%v", s.Host, deviceToken)

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	headers.set(req.Header)

	resp, err := s.Client.Do(req)

	if err != nil {
		if e, ok := err.(*url.Error); ok {
			if e, ok := e.Err.(http2.GoAwayError); ok {
				// parse DebugData as JSON. no status code known (0)
				return "", parseErrorResponse(strings.NewReader(e.DebugData), 0)
			}
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return resp.Header.Get("apns-id"), nil
	}

	return "", parseErrorResponse(resp.Body, resp.StatusCode)
}

func parseErrorResponse(body io.Reader, statusCode int) error {
	var response struct {
		// Reason for failure
		Reason string `json:"reason"`
		// Timestamp for 410 StatusGone (ErrUnregistered)
		Timestamp int64 `json:"timestamp"`
	}
	err := json.NewDecoder(body).Decode(&response)
	if err != nil {
		return err
	}

	es := &Error{
		Reason: mapErrorReason(response.Reason),
		Status: statusCode,
	}

	if response.Timestamp != 0 {
		// the response.Timestamp is Milliseconds, but time.Unix() requires seconds
		es.Timestamp = time.Unix(response.Timestamp/1000, 0).UTC()
	}
	return es
}
