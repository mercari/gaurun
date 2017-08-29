package gaurun

import (
	"errors"
	"testing"
	"time"

	"github.com/RobotsAndPencils/buford/push"
	"github.com/stretchr/testify/assert"
)

// Test utility
func InitPusher(count *int, errorNum int, err error) func(req RequestGaurunNotification) error {
	return func(req RequestGaurunNotification) error {
		*count += 1
		if *count >= errorNum {
			return nil
		} else {
			return err
		}
	}
}

func TestStartPushWorkers(t *testing.T) {
	workerNum := int64(1)
	queueNum := int64(10)

	StartPushWorkers(workerNum, queueNum)

	assert.Equal(t, int64(cap(QueueNotification)), queueNum)
}

func TestIsExternalServerError(t *testing.T) {
	cases := []struct {
		Err      error
		Platform int
		Expected bool
	}{
		{push.ErrIdleTimeout, PlatFormIos, true},
		{push.ErrShutdown, PlatFormIos, true},
		{push.ErrInternalServerError, PlatFormIos, true},
		{push.ErrServiceUnavailable, PlatFormIos, true},
		{errors.New("no error"), PlatFormIos, false},

		{errors.New("Unavailable"), PlatFormAndroid, true},
		{errors.New("InternalServerError"), PlatFormAndroid, true},
		{errors.New("Timeout"), PlatFormAndroid, true},
		{errors.New("no error"), PlatFormAndroid, false},

		{errors.New("no error"), 100 /* neither iOS nor Android */, false},
	}

	for _, c := range cases {
		actual := isExternalServerError(c.Err, c.Platform)
		assert.Equal(t, actual, c.Expected)
	}
}

func TestPushSync(t *testing.T) {
	cases := []struct {
		ErrorNum int
		RetryMax int
		Expected int
	}{
		{0, 10, 1},
		{5, 10, 5},
		{100, 10, 11},
	}

	for _, c := range cases {
		count := 0
		pusher := InitPusher(&count, c.ErrorNum, push.ErrIdleTimeout)
		req := RequestGaurunNotification{Platform: PlatFormIos}

		pushSync(pusher, req, c.RetryMax)

		assert.Equal(t, count, c.Expected)
	}
}

func TestPushAsync(t *testing.T) {
	cases := []struct {
		ErrorNum int
		RetryMax int
		Expected int
	}{
		{0, 10, 1},
		{5, 10, 5},
		{100, 10, 11},
	}

	for _, c := range cases {
		count := 0
		pusher := InitPusher(&count, c.ErrorNum, push.ErrIdleTimeout)
		req := RequestGaurunNotification{Platform: PlatFormIos}

		pusherCount := int64(1)
		PusherCountAll = int64(1)
		PusherWg.Add(1)
		go pushAsync(pusher, req, c.RetryMax, &pusherCount)
		PusherWg.Wait()

		assert.Equal(t, count, c.Expected)
		assert.Equal(t, pusherCount, int64(0))
		assert.Equal(t, PusherCountAll, int64(0))
	}
}

func TestPushNotificationWorker(t *testing.T) {
	cases := []struct {
		Platform  int
		PusherMax int64
		Expected  int
	}{
		{PlatFormIos, 0, 1},
		{PlatFormIos, 1, 1},
		{PlatFormIos, 2, 1},

		{PlatFormAndroid, 0, 1},
		{PlatFormAndroid, 1, 1},
		{PlatFormAndroid, 2, 1},

		{100 /* Neither iOS nor Android */, 0, 0},
	}

	// Save globals
	originalPusherMax := ConfGaurun.Core.PusherMax
	originalPushNotificationIos := PushNotificationIosFunc
	originalPushNotificationAndroid := PushNotificationAndroidFunc

	// Inject dummy pusher
	var count int
	pusher := InitPusher(&count, 0, push.ErrIdleTimeout)
	PushNotificationIosFunc = pusher
	PushNotificationAndroidFunc = pusher

	go pushNotificationWorker()

	for _, c := range cases {
		count = 0

		ConfGaurun.Core.PusherMax = c.PusherMax
		QueueNotification <- RequestGaurunNotification{Platform: c.Platform}

		// Wait processing pusher
		time.Sleep(1 * time.Second)

		assert.Equal(t, count, c.Expected)
	}

	// Restore globals
	ConfGaurun.Core.PusherMax = originalPusherMax
	PushNotificationIosFunc = originalPushNotificationIos
	PushNotificationAndroidFunc = originalPushNotificationAndroid
}
