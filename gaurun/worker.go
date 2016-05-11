package gaurun

func StartPushWorkers(workerNum, queueNum int) {
	QueueNotification = make(chan RequestGaurunNotification, queueNum)
	for i := 0; i < workerNum; i++ {
		go pushNotificationWorker()
	}
}

func pushNotificationWorker() {
	var (
		success  bool
		retryMax int
	)

	for {
		notification := <-QueueNotification

		switch notification.Platform {
		case PlatFormIos:
			success = pushNotificationIos(notification)
			retryMax = ConfGaurun.Ios.RetryMax
		case PlatFormAndroid:
			success = pushNotificationAndroid(notification)
			retryMax = ConfGaurun.Android.RetryMax
		default:
			LogError.Warnf("invalid platform: %d", notification.Platform)
			continue
		}
		if !success && notification.Retry < retryMax {
			if len(QueueNotification) < cap(QueueNotification) {
				notification.Retry++
				QueueNotification <- notification
			}
		}
	}
}
