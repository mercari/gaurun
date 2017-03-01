package gaurun

import (
	"fmt"
	"testing"
)

func init() {
	var err error
	LogAccess, _, err = InitLog("discard", "info")
	if err != nil {
		LogSetupFatal(err)
	}
	LogError, _, err = InitLog("discard", "error")
	if err != nil {
		LogSetupFatal(err)
	}
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
