package gaurun

import (
	"github.com/RobotsAndPencils/buford/push"
	"strings"
)

func StartPushWorkers(workerNum, queueNum int) {
	QueueNotification = make(chan RequestGaurunNotification, queueNum)
	for i := 0; i < workerNum; i++ {
		go pushNotificationWorker()
	}
}

func pushNotificationWorker() {
	var (
		err      error
		retryMax int
	)

	for {
		notification := <-QueueNotification
	Retry:
		switch notification.Platform {
		case PlatFormIos:
			err = pushNotificationIos(notification)
			retryMax = ConfGaurun.Ios.RetryMax
		case PlatFormAndroid:
			err = pushNotificationAndroid(notification)
			retryMax = ConfGaurun.Android.RetryMax
		default:
			LogError.Warnf("invalid platform: %d", notification.Platform)
			continue
		}
		// retry when server error is occurred.
		if err != nil && notification.Retry < retryMax {
			switch notification.Platform {
			case PlatFormIos:
				if err == push.ErrIdleTimeout || err == push.ErrShutdown || err == push.ErrInternalServerError || err == push.ErrServiceUnavailable {
					notification.Retry++
					goto Retry
				}
			case PlatFormAndroid:
				if err.Error() == "Unavailable" || err.Error() == "InternalServerError" || strings.Contains(err.Error(), "Timeout") {
					notification.Retry++
					goto Retry
				}
			default:
				// not through
				continue
			}

		}
	}
}
