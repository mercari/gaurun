package gaurun

import (
	"context"
	"net"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"github.com/mercari/gaurun/gcm"
	"google.golang.org/api/option"
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

// InitGCMClient initializes GCMClient which is globally declared.
func InitGCMClient() error {
	// By default, use FCM endpoint. If UseFCM is explicitly disabled via configuration,
	// use GCM endpoint.
	url := gcm.FCMSendEndpoint
	if !ConfGaurun.Android.UseFCM {
		url = gcm.GCMSendEndpoint
	}

	var err error
	GCMClient, err = gcm.NewClient(url, ConfGaurun.Android.ApiKey)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout: time.Duration(ConfGaurun.Android.KeepAliveTimeout) * time.Second,
	}

	GCMClient.Http = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
	}

	return nil
}

func InitFirebaseApp() error {
	transport := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.FCMV1.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.FCMV1.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.FCMV1.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout: time.Duration(ConfGaurun.FCMV1.KeepAliveTimeout) * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(ConfGaurun.FCMV1.Timeout) * time.Second,
	}

	opts := make([]option.ClientOption, 2)
	opts[0] = option.WithCredentialsFile(ConfGaurun.FCMV1.CredentialsFile)
	opts[1] = option.WithHTTPClient(client)

	// if ConfGaurun.Android.Project is empty string, it is acquired from the contents of ConfGaurun.Android.CredentialsFile
	config := &firebase.Config{ProjectID: ConfGaurun.FCMV1.Project}

	var err error

	FirebaseApp, err = firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		return err
	}

	return nil
}

func InitAPNSClient() error {
	var err error
	APNSClient, err = NewApnsClientHttp2(
		ConfGaurun.Ios.PemCertPath,
		ConfGaurun.Ios.PemKeyPath,
		ConfGaurun.Ios.PemKeyPassphrase,
	)
	if err != nil {
		return err
	}
	return nil
}
