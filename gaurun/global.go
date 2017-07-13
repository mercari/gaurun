package gaurun

import (
	"net/http"

	"github.com/mercari/gaurun/gcm"

	"go.uber.org/zap"
)

var (
	ConfGaurun        ConfToml
	QueueNotification chan RequestGaurunNotification
	CertificatePemIos CertificatePem
	LogAccess         *zap.Logger
	LogError          *zap.Logger
	StatGaurun        StatApp
	// for numbering push
	SeqID         uint64
	GCMClient     *gcm.Client
	APNSClient    *http.Client
)
