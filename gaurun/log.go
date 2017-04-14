package gaurun

import (
	"log"
	"math"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/client9/reopen"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogReq struct {
	Type          string `json:"type"`
	Time          string `json:"time"`
	URI           string `json:"uri"`
	Method        string `json:"method"`
	Proto         string `json:"proto"`
	ContentLength int64  `json:"content_length"`
}

type LogPushEntry struct {
	Type     string  `json:"type"`
	Time     string  `json:"time"`
	ID       uint64  `json:"id"`
	Platform string  `json:"platform"`
	Token    string  `json:"token"`
	Message  string  `json:"message"`
	Ptime    float64 `json:"ptime"`
	Error    string  `json:"error"`
	// Android
	CollapseKey    string `json:"collapse_key,omitempty"`
	DelayWhileIdle bool   `json:"delay_while_idle,omitempty"`
	TimeToLive     int    `json:"time_to_live,omitempty"`
	// iOS
	Badge            int    `json:"badge,omitempty"`
	Sound            string `json:"sound,omitempty"`
	ContentAvailable bool   `json:"content_available,omitempty"`
	MutableContent   bool   `json:"mutable_content,omitempty"`
	Expiry           int    `json:"expiry,omitempty"`
}

type Reopener interface {
	Reopen() error
}

func LocalTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006/01/02 15:04:05 MST"))
}

func InitLog(outString, levelString string) (*zap.Logger, Reopener, error) {
	var writer reopen.Writer
	switch outString {
	case "stdout":
		writer = reopen.Stdout
	case "stderr":
		writer = reopen.Stderr
	case "discard":
		writer = reopen.Discard
	default:
		f, err := reopen.NewFileWriterMode(outString, 0644)
		if err != nil {
			return nil, nil, err
		}
		writer = f
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(levelString)); err != nil {
		return nil, nil, err
	}

	cfg := zap.NewProductionConfig().EncoderConfig
	cfg.TimeKey = "time"
	cfg.MessageKey = "message"
	cfg.EncodeTime = LocalTimeEncoder

	encoder := zapcore.NewJSONEncoder(cfg)
	writeSyncer := zapcore.AddSync(writer)
	logger := zap.New(
		zapcore.NewCore(
			encoder,
			zapcore.Lock(writeSyncer),
			level,
		),
		zap.ErrorOutput(writeSyncer),
	)

	return logger, writer, nil
}

// LogSetupFatal output error log with log package and exit immediately.
func LogSetupFatal(err error) {
	log.Fatal(err)
}

func LogAcceptedRequest(r *http.Request) {
	LogAccess.Info("",
		zap.String("type", "accepted-request"),
		zap.String("uri", r.URL.String()),
		zap.String("method", r.Method),
		zap.String("proto", r.Proto),
		zap.Int64("content_length", r.ContentLength),
	)
}

func LogPush(id uint64, status, token string, ptime float64, req RequestGaurunNotification, errPush error) {
	var plat string
	switch req.Platform {
	case PlatFormIos:
		plat = "ios"
	case PlatFormAndroid:
		plat = "android"
	case PlatformFake:
		plat = "fake"
	}

	ptime = math.Floor(ptime*1000) / 1000 // %.3f conversion

	errMsg := ""
	if errPush != nil {
		errMsg = errPush.Error()
	}

	var logger func(string, ...zapcore.Field)
	switch status {
	case StatusAcceptedPush:
		fallthrough
	case StatusSucceededPush:
		logger = LogAccess.Info
	case StatusFailedPush:
		fallthrough
	case StatusDisabledPush:
		logger = LogError.Error
	}

	// omitempty handling for device dependent values
	collapseKey := zap.Skip()
	if req.CollapseKey != "" {
		collapseKey = zap.String("collapse_key", req.CollapseKey)
	}
	delayWhileIdle := zap.Skip()
	if req.DelayWhileIdle {
		delayWhileIdle = zap.Bool("delay_while_idle", req.DelayWhileIdle)
	}
	timeToLive := zap.Skip()
	if req.TimeToLive != 0 {
		timeToLive = zap.Int("time_to_live", req.TimeToLive)
	}
	badge := zap.Skip()
	if req.Badge != 0 {
		badge = zap.Int("badge", req.Badge)
	}
	sound := zap.Skip()
	if req.Sound != "" {
		sound = zap.String("sound", req.Sound)
	}
	contentAvailable := zap.Skip()
	if req.ContentAvailable {
		contentAvailable = zap.Bool("content_available", req.ContentAvailable)
	}
	mutableContent := zap.Skip()
	if req.MutableContent {
		mutableContent = zap.Bool("mutable_content", req.MutableContent)
	}
	expiry := zap.Skip()
	if req.Expiry != 0 {
		expiry = zap.Int("expiry", req.Expiry)
	}

	logger(req.Message,
		zap.Uint64("id", id),
		zap.String("platform", plat),
		zap.String("token", token),
		zap.String("type", status),
		zap.Float64("ptime", ptime),
		zap.String("error", errMsg),
		collapseKey,
		delayWhileIdle,
		timeToLive,
		badge,
		sound,
		contentAvailable,
		mutableContent,
		expiry,
	)
}

func numberingPush() uint64 {
	return atomic.AddUint64(&SeqID, 1)
}
