package gaurun

import (
	"github.com/Sirupsen/logrus"
	"sync"
)

var (
	ConfGaurun        ConfToml
	QueueNotification chan RequestGaurunNotification
	CertificatePemIos CertificatePem
	LogAccess         *logrus.Logger
	LogError          *logrus.Logger
	StatGaurun        StatApp
	// for numbering push
	OnceNumbering sync.Once
	WgNumbering   *sync.WaitGroup
	SeqID         uint64
)
