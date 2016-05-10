package gaurun

import (
	"net/http"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/mercari/gcm"
)

var (
	ConfGaurun        ConfToml
	QueueNotification chan RequestGaurunNotification
	CertificatePemIos CertificatePem
	LogAccess         *logrus.Logger
	LogError          *logrus.Logger
	StatGaurun        StatApp
	// for numbering push
	OnceNumbering   sync.Once
	WgNumbering     *sync.WaitGroup
	SeqID           uint64
	GCMClient       *gcm.Sender
	APNSClient      *http.Client
	TransportGaurun *http.Transport
)
