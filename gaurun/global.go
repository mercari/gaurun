package gaurun

import (
	"net/http"
	"sync"

	"github.com/mercari/gcm"
	"github.com/uber-go/zap"
)

var (
	ConfGaurun        ConfToml
	QueueNotification chan RequestGaurunNotification
	CertificatePemIos CertificatePem
	LogAccess         zap.Logger
	LogError          zap.Logger
	StatGaurun        StatApp
	// for numbering push
	OnceNumbering sync.Once
	WgNumbering   *sync.WaitGroup
	SeqID         uint64
	GCMClient     *gcm.Sender
	APNSClient    *http.Client
)
