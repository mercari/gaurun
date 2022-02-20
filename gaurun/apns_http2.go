package gaurun

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/mercari/gaurun/buford/payload"
	"github.com/mercari/gaurun/buford/payload/badge"
	"github.com/mercari/gaurun/buford/push"
	"github.com/mercari/gaurun/buford/token"

)

type APNsClient struct {
	HTTPClient *http.Client
	// Token is set only for token-based provider connection trust
	Token *token.Token
}

func NewTransportHttp2(cert tls.Certificate) (*http.Transport, error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	config.BuildNameToCertificate()

	transport := &http.Transport{
		TLSClientConfig:     config,
		MaxIdleConnsPerHost: ConfGaurun.Ios.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Ios.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Ios.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout:   time.Duration(ConfGaurun.Ios.KeepAliveTimeout) * time.Second,
		ForceAttemptHTTP2: true,
	}

	return transport, nil
}

func NewApnsClientHttp2(certPath, keyPath, keyPassphrase string) (APNsClient, error) {
	cert, err := loadX509KeyPairWithPassword(certPath, keyPath, keyPassphrase)
	if err != nil {
		return APNsClient{}, err
	}

	transport, err := NewTransportHttp2(cert)
	if err != nil {
		return APNsClient{}, err
	}

	return APNsClient{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(ConfGaurun.Ios.Timeout) * time.Second,
		},
	}, nil
}

func NewApnsClientHttp2ForToken(authKey *ecdsa.PrivateKey, keyID, teamID string) (APNsClient, error) {
	authToken := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Ios.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Ios.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Ios.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout:   time.Duration(ConfGaurun.Ios.KeepAliveTimeout) * time.Second,
		ForceAttemptHTTP2: true,
	}

	return APNsClient{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(ConfGaurun.Ios.Timeout) * time.Second,
		},
		Token: authToken,
	}, nil
}

func loadX509KeyPairWithPassword(certPath, keyPath, keyPassphrase string) (tls.Certificate, error) {
	keyPEMBlock, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return tls.Certificate{}, err
	}
	if keyPassphrase != "" {
		pemBlock, _ := pem.Decode(keyPEMBlock)
		if !x509.IsEncryptedPEMBlock(pemBlock) {
			err = fmt.Errorf("%s is not encrypted. passphrase is not required", keyPath)
			return tls.Certificate{}, err
		}
		keyPEMBlock, err = x509.DecryptPEMBlock(pemBlock, []byte(keyPassphrase))
		if err != nil {
			return tls.Certificate{}, err
		}
		keyPEMBlock = pem.EncodeToMemory(&pem.Block{Type: pemBlock.Type, Bytes: keyPEMBlock})
	}
	certPEMBlock, err := ioutil.ReadFile(certPath)
	if err != nil {
		return tls.Certificate{}, err
	}
	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}

func NewApnsServiceHttp2(apnsClient APNsClient) *push.Service {
	var host string
	if ConfGaurun.Ios.Sandbox {
		host = push.Development
	} else {
		host = push.Production
	}
	return &push.Service{
		Client: apnsClient.HTTPClient,
		Host:   host,
	}
}

func NewApnsPayloadHttp2(req *RequestGaurunNotification) map[string]interface{} {
	p := payload.APS{
		Alert:            payload.Alert{Title: req.Title, Body: req.Message, Subtitle: req.Subtitle},
		Badge:            badge.New(uint(req.Badge)),
		Category:         req.Category,
		Sound:            req.Sound,
		ContentAvailable: req.ContentAvailable,
		MutableContent:   req.MutableContent,
	}

	pm := p.Map()

	if len(req.Extend) > 0 {
		for _, extend := range req.Extend {
			pm[extend.Key] = extend.Value
		}
	}

	return pm
}

func NewApnsHeadersHttp2(req *RequestGaurunNotification) *push.Headers {
	var pushType push.PushType

	// Required when delivering notifications to devices running iOS 13 and later, or watchOS 6 and later. Ignored on earlier system versions.
	// cf: https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/sending_notification_requests_to_apns
	if req.PushType == ApnsPushTypeBackground {
		pushType = push.PushTypeBackground
	} else {
		pushType = push.PushTypeAlert
	}

	headers := &push.Headers{
		Topic:    ConfGaurun.Ios.Topic,
		ID:	  ConfGaurun.Ios.ApnsId,
		PushType: pushType,
	}

	LogError.Debug("apns_https2.go topic:" + ConfGaurun.Ios.Topic +", apns-id:" + ConfGaurun.Ios.ApnsId)

	if req.Expiry > 0 {
		headers.Expiration = time.Now().Add(time.Duration(int64(req.Expiry)) * time.Second).UTC()
	}

	return headers
}

func NewApnsHeadersHttp2WithToken(req *RequestGaurunNotification, t *token.Token) *push.Headers {
	headers := NewApnsHeadersHttp2(req)
	headers.AuthToken = t

	return headers
}

func ApnsPushHttp2(token string, service *push.Service, headers *push.Headers, payload map[string]interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = service.Push(token, headers, b)
	return err
}
