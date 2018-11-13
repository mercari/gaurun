package gaurun

import (
	"net/http"

	"github.com/mercari/gaurun/fcm"
	"github.com/mercari/gaurun/gcm"

	"go.uber.org/zap"
)

var (
	// Toml configuration for Gaurun
	ConfGaurun ConfToml
	// push notification Queue
	QueueNotification chan RequestGaurunNotification
	// TLS certificate and key for APNs
	CertificatePemIos CertificatePem
	// Stat for Gaurun
	StatGaurun StatApp
	// http client for APNs and GCM/FCM
	APNSClient *http.Client
	GCMClient  *gcm.Client
	FCMClient  *fcm.Client
	// access and error logger
	LogAccess *zap.Logger
	LogError  *zap.Logger
	// sequence ID for numbering push
	SeqID uint64
)
