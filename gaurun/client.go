package gaurun

import (
	"net"
	"net/http"
	"time"

	"github.com/mercari/gcm"
)

func keepAliveInterval(keepAliveTimeout int) int {
	const minInterval = 30
	if keepAliveTimeout <= minInterval {
		return keepAliveTimeout
	}
	result := keepAliveTimeout / 3
	if result < minInterval {
		return minInterval
	}
	return result
}

func InitHttpClient() error {
	TransportGCM := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout),
		}).Dial,
		IdleConnTimeout: time.Duration(ConfGaurun.Android.KeepAliveTimeout) * time.Second,
	}
	GCMClient = &gcm.Sender{
		ApiKey: ConfGaurun.Android.ApiKey,
		Http: &http.Client{
			Transport: TransportGCM,
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
		},
	}

	var err error
	APNSClient, err = NewApnsClientHttp2(
		ConfGaurun.Ios.PemCertPath,
		ConfGaurun.Ios.PemKeyPath,
	)
	if err != nil {
		return err
	}
	return nil
}
