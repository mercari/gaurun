package gaurun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
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
	Badge  int    `json:"badge,omitempty"`
	Sound  string `json:"sound,omitempty"`
	Expiry int    `json:"expiry,omitempty"`
}

type GaurunFormatter struct {
}

func (f *GaurunFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "[%s] ", entry.Level.String())
	fmt.Fprintf(b, "%s", entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func InitLog() *logrus.Logger {
	return logrus.New()
}

func SetLogOut(log *logrus.Logger, outString string) error {
	switch outString {
	case "stdout":
		log.Out = os.Stdout
	case "stderr":
		log.Out = os.Stderr
	default:
		f, err := os.OpenFile(outString, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		log.Out = f
	}
	return nil
}

func SetLogLevel(log *logrus.Logger, levelString string) error {
	level, err := logrus.ParseLevel(levelString)
	if err != nil {
		return err
	}
	log.Level = level
	return nil
}

func LogAcceptedRequest(uri, method, proto string, length int64) {
	log := &LogReq{
		Type:          "accepted-request",
		Time:          time.Now().Format("2006/01/02 15:04:05 MST"),
		URI:           uri,
		Method:        method,
		Proto:         proto,
		ContentLength: length,
	}
	logJSON, err := json.Marshal(log)
	if err != nil {
		LogError.Error("Marshaling JSON error")
		return
	}
	LogAccess.Info(string(logJSON))
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

	log := &LogPushEntry{
		Type:           status,
		Time:           time.Now().Format("2006/01/02 15:04:05 MST"),
		ID:             id,
		Platform:       plat,
		Token:          token,
		Message:        req.Message,
		Ptime:          ptime,
		Error:          errMsg,
		CollapseKey:    req.CollapseKey,
		DelayWhileIdle: req.DelayWhileIdle,
		TimeToLive:     req.TimeToLive,
		Badge:          req.Badge,
		Sound:          req.Sound,
		Expiry:         req.Expiry,
	}
	logJSON, err := json.Marshal(log)
	if err != nil {
		LogError.Error("Marshaling JSON error")
		return
	}

	switch status {
	case StatusAcceptedPush:
		fallthrough
	case StatusSucceededPush:
		LogAccess.Info(string(logJSON))
	case StatusFailedPush:
		LogError.Error(string(logJSON))
	}
}

func numberingPush() uint64 {
	var result uint64

	// nubmering push
	OnceNumbering.Do(func() {
		WgNumbering = new(sync.WaitGroup)
	})
	WgNumbering.Add(1)
	atomic.AddUint64(&SeqID, 1)
	result = SeqID
	WgNumbering.Done()
	WgNumbering.Wait()
	return result
}
