package gaurun

import (
	"fmt"
	"testing"
	"time"

	"github.com/uber-go/zap"
)

func init() {
	encoder := zap.NewJSONEncoder(
		zap.MessageKey("message"),
		zap.TimeFormatter(func(t time.Time) zap.Field {
			return zap.String("time", t.Local().Format("2006/01/02 15:04:05 MST"))
		}),
	)
	LogAccess = zap.New(encoder, zap.DiscardOutput)
	LogError = zap.New(encoder, zap.DiscardOutput)
}

func BenchmarkLogPushIOSOmitempty(b *testing.B) {
	req := RequestGaurunNotification{
		Platform: PlatFormIos,
	}
	errPush := fmt.Errorf("error")
	for i := 0; i < b.N; i++ {
		LogPush(uint64(100), StatusAcceptedPush, "xxx", 0.123, req, errPush)
	}
}

func BenchmarkLogPushIOSFull(b *testing.B) {
	req := RequestGaurunNotification{
		Platform:         PlatFormIos,
		Badge:            1,
		Sound:            "foo",
		ContentAvailable: true,
		Expiry:           100,
	}
	errPush := fmt.Errorf("error")
	for i := 0; i < b.N; i++ {
		LogPush(uint64(100), StatusAcceptedPush, "xxx", 0.123, req, errPush)
	}
}
