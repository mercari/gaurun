package gaurun

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/uber-go/zap"
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
	Expiry           int    `json:"expiry,omitempty"`
}

func InitLog(outString string) (zap.Logger, error) {
	var writer zap.WriteSyncer
	switch outString {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	default:
		f, err := os.OpenFile(outString, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		writer = f
	}

	encoder := zap.NewJSONEncoder(
		zap.RFC3339Formatter("time"),
	)
	return zap.New(encoder, zap.Output(writer), zap.ErrorOutput(writer)), nil
}

func SetLogLevel(log zap.Logger, levelString string) error {
	var level zap.Level
	if err := level.UnmarshalText([]byte(levelString)); err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}

// LogSetupError output error log with log package and exit immediately.
func LogSetupError(err error) {
	log.Fatal(err)
}

func LogAcceptedRequest(uri, method, proto string, length int64) {
	LogAccess.Info("accepted-request",
		zap.String("uri", uri),
		zap.String("method", method),
		zap.String("proto", proto),
		zap.Int64("content_length", length),
	)
}

func LogPush(id uint64, status, token string, ptime float64, req RequestGaurunNotification, errPush error) {
	var plat string
	switch req.Platform {
	case PlatFormIos:
		plat = "ios"
	case PlatFormAndroid:
		plat = "android"
	}

	ptime3 := fmt.Sprintf("%.3f", ptime)
	ptime, _ = strconv.ParseFloat(ptime3, 64)

	errMsg := ""
	if errPush != nil {
		errMsg = errPush.Error()
	}

	var logger func(string, ...zap.Field)
	switch status {
	case StatusAcceptedPush:
		fallthrough
	case StatusSucceededPush:
		logger = LogAccess.Info
	case StatusFailedPush:
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
		expiry,
	)
}

func numberingPush() uint64 {
	return atomic.AddUint64(&SeqID, 1)
}
