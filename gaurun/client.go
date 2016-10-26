package gaurun

import (
	"net"
	"net/http"
	"time"

	"github.com/mercari/gcm"
)

func keepAliveInterval(keepAliveTimeout int) int {
	const minInterval = 30
	const maxInterval = 90
	if keepAliveTimeout <= minInterval {
		return keepAliveTimeout
	}
	result := keepAliveTimeout / 3
	if result < minInterval {
		return minInterval
	}
	if result > maxInterval {
		return maxInterval
	}
	return result
}

func InitHttpClient() error {
	TransportGCM := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout)) * time.Second,
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
