package gaurun

import (
	"fmt"
	"testing"

	"github.com/uber-go/zap"
)

func init() {
	LogAccess = zap.New(zap.NewJSONEncoder(zap.RFC3339Formatter("time")), zap.DiscardOutput)
	LogError = zap.New(zap.NewJSONEncoder(zap.RFC3339Formatter("time")), zap.DiscardOutput)
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
