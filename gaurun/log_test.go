package gaurun

import (
	"fmt"
	"io/ioutil"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	cfg := zap.NewProductionConfig().EncoderConfig
	cfg.TimeKey = "time"
	cfg.MessageKey = "message"
	cfg.EncodeTime = LocalTimeEncoder

	encoder := zapcore.NewJSONEncoder(cfg)

	LogAccess = zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(ioutil.Discard),
			zapcore.ErrorLevel,
		),
	)
	LogError = zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(ioutil.Discard),
			zapcore.ErrorLevel,
		),
	)
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
