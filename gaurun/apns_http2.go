package gaurun

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"

	"golang.org/x/net/http2"
)

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
		IdleConnTimeout: time.Duration(ConfGaurun.Ios.KeepAliveTimeout) * time.Second,
	}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, err
	}

	return transport, nil
}

func NewApnsClientHttp2(certPath, keyPath, keyPassphrase string) (*http.Client, error) {
	cert, err := loadX509KeyPairWithPassword(certPath, keyPath, keyPassphrase)
	if err != nil {
		return nil, err
	}

	transport, err := NewTransportHttp2(cert)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(ConfGaurun.Ios.Timeout) * time.Second,
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

func NewApnsServiceHttp2(client *http.Client) *push.Service {
	var host string
	if ConfGaurun.Ios.Sandbox {
		host = push.Development
	} else {
		host = push.Production
	}
	return &push.Service{
		Client: client,
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
		ThreadID:         req.ThreadID,
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
	headers := &push.Headers{
		Topic: ConfGaurun.Ios.Topic,
	}

	if req.Expiry > 0 {
		headers.Expiration = time.Now().Add(time.Duration(int64(req.Expiry)) * time.Second).UTC()
	}

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
