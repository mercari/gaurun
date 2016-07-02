package gaurun

import (
	"github.com/RobotsAndPencils/buford/push"
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
		if err != nil && notification.Retry < retryMax {
			// gaurun does not retry to push notification
			// when token is invalid.
			switch notification.Platform {
			case PlatFormIos:
				if err == push.ErrUnregistered || err == push.ErrDeviceTokenNotForTopic {
					continue
				}
			case PlatFormAndroid:
				if err.Error() == "NotRegistered" {
					continue
				}
			default:
				// not through
				continue
			}

			notification.Retry++
			goto Retry
		}
	}
}
