package gaurun

import (
	"github.com/mercari/gaurun/gcm"

	"go.uber.org/zap"
)

var (
	// Toml configuration for Gaurun
	ConfGaurun ConfToml
	// push notification Queue
	QueueNotification chan RequestGaurunNotification
	// Stat for Gaurun
	StatGaurun StatApp
	// http client for APNs and GCM/FCM
	APNSClient APNsClient
	GCMClient  *gcm.Client
	// access and error logger
	LogAccess *zap.Logger
	LogError  *zap.Logger
	// sequence ID for numbering push
	SeqID uint64
)
