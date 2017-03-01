package gaurun

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/RobotsAndPencils/buford/push"
)

var (
	// PusherCountAll is the shared value between workers
	PusherCountAll int64

	// PusherWg is global wait group for pusher worker.
	// It increments when new pusher is swapned and decrements when job is done.
	//
	// This is used to block main process to shutdown while pusher is still working.
	PusherWg sync.WaitGroup
)

func init() {
	PusherCountAll = 0
}

func StartPushWorkers(workerNum, queueNum int64) {
	QueueNotification = make(chan RequestGaurunNotification, queueNum)
	for i := int64(0); i < workerNum; i++ {
		go pushNotificationWorker()
	}
}

func isExternalServerError(err error, platform int) bool {
	switch platform {
	case PlatFormIos:
		if err == push.ErrIdleTimeout || err == push.ErrShutdown || err == push.ErrInternalServerError || err == push.ErrServiceUnavailable {
			return true
		}
	case PlatFormAndroid:
		if err.Error() == "Unavailable" || err.Error() == "InternalServerError" || strings.Contains(err.Error(), "Timeout") {
			return true
		}
	default:
		// not through
	}
	return false
}

func pushSync(pusher func(req RequestGaurunNotification) error, req RequestGaurunNotification, retryMax int) {
	PusherWg.Add(1)
	defer PusherWg.Done()
Retry:
	err := pusher(req)
	if err != nil && req.Retry < retryMax && isExternalServerError(err, req.Platform) {
		req.Retry++
		goto Retry
	}
}

func pushAsync(pusher func(req RequestGaurunNotification) error, req RequestGaurunNotification, retryMax int, pusherCount *int64) {
	defer PusherWg.Done()
Retry:
	err := pusher(req)
	if err != nil && req.Retry < retryMax && isExternalServerError(err, req.Platform) {
		req.Retry++
		goto Retry
	}

	atomic.AddInt64(pusherCount, -1)
	atomic.AddInt64(&PusherCountAll, -1)
}

func pushNotificationWorker() {
	var (
		retryMax    int
		pusher      func(req RequestGaurunNotification) error
		pusherCount int64
	)

	// pusherCount is the independent value between workers
	pusherCount = 0

	for {
		notification := <-QueueNotification

		switch notification.Platform {
		case PlatFormIos:
			pusher = pushNotificationIos
			retryMax = ConfGaurun.Ios.RetryMax
		case PlatFormAndroid:
			pusher = pushNotificationAndroid
			retryMax = ConfGaurun.Android.RetryMax
		default:
			LogError.Warn(fmt.Sprintf("invalid platform: %d", notification.Platform))
			continue
		}

		if atomic.LoadInt64(&ConfGaurun.Core.PusherMax) <= 0 {
			pushSync(pusher, notification, retryMax)
			continue
		}

		if atomic.LoadInt64(&pusherCount) < atomic.LoadInt64(&ConfGaurun.Core.PusherMax) {
			// Do not increment pusherCount and PusherCountAll in pushAsync().
			// Because pusherCount and PusherCountAll are sometimes over pusherMax
			// as the increment in goroutine runs asynchronously.
			atomic.AddInt64(&pusherCount, 1)
			atomic.AddInt64(&PusherCountAll, 1)
			PusherWg.Add(1)
			go pushAsync(pusher, notification, retryMax, &pusherCount)
			continue
		} else {
			pushSync(pusher, notification, retryMax)
			continue
		}
	}
}
