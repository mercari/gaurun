package gaurun

import (
	"net/http"

	firebase "firebase.google.com/go"
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
	APNSClient  *http.Client
	GCMClient   *gcm.Client
	FirebaseApp *firebase.App
	// access and error logger
	LogAccess *zap.Logger
	LogError  *zap.Logger
	// sequence ID for numbering push
	SeqID uint64
)
