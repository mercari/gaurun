package gaurun

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"
)

func NewApnsClientHttp2(certPath, keyPath string) (*http.Client, error) {
	var client *http.Client
	var cert tls.Certificate
	var err error
	cert, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return client, err
	}

	client, err = push.NewClient(cert)
	if err != nil {
		return client, err
	}

	return client, nil
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
		Alert: payload.Alert{Body: req.Message},
		Badge: badge.New(uint(req.Badge)),
		Sound: req.Sound,
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
	return &push.Headers{
		Expiration: time.Unix(int64(req.Expiry), 0),
	}
}

func ApnsPushHttp2(token string, service *push.Service, headers *push.Headers, payload map[string]interface{}) error {
	_, err := service.Push(token, headers, payload)
	return err
}
