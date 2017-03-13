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

func InitGCMClient() {
	// By default, use GCM endpoint. If UseFCM is explicitly enabled via configuration,
	// use FCM endpoint.
	GCMClient, _ := gcm.NewClient(gcm.GcmSendEndpoint, ConfGaurun.Android.ApiKey)
	if ConfGaurun.Android.UseFCM {
		GCMClient, _ = gcm.NewClient(gcm.FCMSendEndpoint, ConfGaurun.Android.ApiKey)
	}

	TransportGCM := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout: time.Duration(ConfGaurun.Android.KeepAliveTimeout) * time.Second,
	}

	GCMClient.Http = &http.Client{
		Transport: TransportGCM,
		Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
	}
}

func InitAPNSClient() error {
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
