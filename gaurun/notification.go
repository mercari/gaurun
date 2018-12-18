package gaurun

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/mercari/gaurun/gcm"

	"go.uber.org/zap"
)

type RequestGaurun struct {
	Notifications []RequestGaurunNotification `json:"notifications"`
}

type RequestGaurunNotification struct {
	// Common
	Tokens     []string `json:"token"`
	Platform   int      `json:"platform"`
	Message    string   `json:"message,omitempty"`
	Identifier string   `json:"identifier,omitempty"`
	// Android
	CollapseKey    string `json:"collapse_key,omitempty"`
	DelayWhileIdle bool   `json:"delay_while_idle,omitempty"`
	TimeToLive     int    `json:"time_to_live,omitempty"`
	// FCM v1
	Body string `json:"body,omitempty"`
	// iOS and FCM v1
	Title string `json:"title,omitempty"`
	// iOS
	Subtitle         string       `json:"subtitle,omitempty"`
	Badge            int          `json:"badge,omitempty"`
	Category         string       `json:"category,omitempty"`
	Sound            string       `json:"sound,omitempty"`
	ContentAvailable bool         `json:"content_available,omitempty"`
	MutableContent   bool         `json:"mutable_content,omitempty"`
	Expiry           int          `json:"expiry,omitempty"`
	Retry            int          `json:"retry,omitempty"`
	Extend           []ExtendJSON `json:"extend,omitempty"`
	// meta
	ID uint64 `json:"seq_id,omitempty"`
}

type ExtendJSON struct {
	Key   string `json:"key"`
	Value string `json:"val"`
}

type ResponseGaurun struct {
	Message string `json:"message"`
}

type CertificatePem struct {
	Cert []byte
	Key  []byte
}

func enqueueNotifications(notifications []RequestGaurunNotification) {
	for _, notification := range notifications {
		err := validateNotification(&notification)
		if err != nil {
			LogError.Error(err.Error())
			continue
		}
		var enabledPush bool
		switch notification.Platform {
		case PlatFormIos:
			enabledPush = ConfGaurun.Ios.Enabled
		case PlatFormAndroid:
			enabledPush = ConfGaurun.Android.Enabled
		case PlatFormFCMV1:
			enabledPush = ConfGaurun.FCMV1.Enabled
		}
		// Enqueue notification per token
		for _, token := range notification.Tokens {
			notification2 := notification
			notification2.Tokens = []string{token}
			notification2.ID = numberingPush()
			if enabledPush {
				LogPush(notification2.ID, StatusAcceptedPush, token, 0, notification2, nil)
				QueueNotification <- notification2
			} else {
				LogPush(notification2.ID, StatusDisabledPush, token, 0, notification2, nil)
			}
		}
	}
}

func pushNotificationIos(req RequestGaurunNotification) error {
	LogError.Debug("START push notification for iOS")

	service := NewApnsServiceHttp2(APNSClient)

	token := req.Tokens[0]

	headers := NewApnsHeadersHttp2(&req)
	payload := NewApnsPayloadHttp2(&req)

	stime := time.Now()
	err := ApnsPushHttp2(token, service, headers, payload)

	etime := time.Now()
	ptime := etime.Sub(stime).Seconds()

	if err != nil {
		atomic.AddInt64(&StatGaurun.Ios.PushError, 1)
		LogPush(req.ID, StatusFailedPush, token, ptime, req, err)
		return err
	}

	atomic.AddInt64(&StatGaurun.Ios.PushSuccess, 1)
	LogPush(req.ID, StatusSucceededPush, token, ptime, req, nil)

	LogError.Debug("END push notification for iOS")

	return nil
}

func pushNotificationAndroid(req RequestGaurunNotification) error {
	LogError.Debug("START push notification for Android")

	data := map[string]interface{}{"message": req.Message}
	if len(req.Extend) > 0 {
		for _, extend := range req.Extend {
			data[extend.Key] = extend.Value
		}
	}

	token := req.Tokens[0]

	msg := gcm.NewMessage(data, token)
	msg.CollapseKey = req.CollapseKey
	msg.DelayWhileIdle = req.DelayWhileIdle
	msg.TimeToLive = req.TimeToLive

	stime := time.Now()
	_, err := GCMClient.Send(msg)
	etime := time.Now()
	ptime := etime.Sub(stime).Seconds()
	if err != nil {
		atomic.AddInt64(&StatGaurun.Android.PushError, 1)
		LogPush(req.ID, StatusFailedPush, token, ptime, req, err)
		return err
	}

	LogPush(req.ID, StatusSucceededPush, token, ptime, req, nil)

	atomic.AddInt64(&StatGaurun.Android.PushSuccess, int64(len(req.Tokens)))
	LogError.Debug("END push notification for Android")

	return nil
}

func pushNotificationFCMV1(req RequestGaurunNotification) error {
	LogError.Debug("START push notification for FCMv1")

	data := make(map[string]string)
	if len(req.Extend) > 0 {
		for _, extend := range req.Extend {
			data[extend.Key] = extend.Value
		}
	}

	client, err := FirebaseApp.Messaging(context.Background())
	if err != nil {
		return err
	}

	token := req.Tokens[0]

	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Title: req.Title,
			Body:  req.Body,
		},
		Token: token,
	}

	stime := time.Now()
	_, err = client.Send(context.Background(), msg)
	etime := time.Now()
	ptime := etime.Sub(stime).Seconds()
	if err != nil {
		atomic.AddInt64(&StatGaurun.Android.PushError, 1)
		LogPush(req.ID, StatusFailedPush, token, ptime, req, err)
		return err
	}

	LogPush(req.ID, StatusSucceededPush, token, ptime, req, nil)

	atomic.AddInt64(&StatGaurun.Android.PushSuccess, int64(len(req.Tokens)))
	LogError.Debug("END push notification for FCMv1")

	return nil
}

func validateNotification(notification *RequestGaurunNotification) error {

	for _, token := range notification.Tokens {
		if len(token) == 0 {
			return errors.New("empty token")
		}
	}

	if notification.Platform < PlatFormIos || notification.Platform > PlatFormFCMV1 {
		return errors.New("invalid platform")
	}

	if (notification.Platform == PlatFormIos || notification.Platform == PlatFormAndroid) && len(notification.Message) == 0 {
		return errors.New("empty message")
	} else if notification.Platform == PlatFormFCMV1 && (len(notification.Title) == 0 || len(notification.Body) == 0) {
		return errors.New("empty title or body")
	}

	return nil
}

func sendResponse(w http.ResponseWriter, msg string, code int) {
	respGaurun := ResponseGaurun{
		Message: msg,
	}
	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(respGaurun); err != nil {
		buf = bytes.NewBufferString("{\"message\":\"Response-body could not be created\"}")
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Server", serverHeader())

	w.WriteHeader(code)

	w.Write(buf.Bytes())
}

func PushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	LogAcceptedRequest(r)
	LogError.Debug("push-request is Accepted")

	LogError.Debug("method check")
	if r.Method != "POST" {
		sendResponse(w, "method must be POST", http.StatusBadRequest)
		return
	}

	LogError.Debug("content-length check")
	if r.ContentLength == 0 {
		sendResponse(w, "request body is empty", http.StatusBadRequest)
		return
	}

	var (
		reqGaurun RequestGaurun
		err       error
	)

	if ConfGaurun.Log.Level == "debug" {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendResponse(w, "failed to read request-body", http.StatusInternalServerError)
			return
		}
		if m := LogError.Check(zap.DebugLevel, "parse request body"); m != nil {
			m.Write(zap.String("body", string(reqBody)))
		}
		err = json.Unmarshal(reqBody, &reqGaurun)
	} else {
		LogError.Debug("parse request body")
		err = json.NewDecoder(r.Body).Decode(&reqGaurun)
	}

	if err != nil {
		LogError.Error(err.Error())
		sendResponse(w, "Request-body is malformed", http.StatusBadRequest)
		return
	}

	if len(reqGaurun.Notifications) == 0 {
		LogError.Error("empty notification")
		sendResponse(w, "empty notification", http.StatusBadRequest)
		return
	} else if int64(len(reqGaurun.Notifications)) > ConfGaurun.Core.NotificationMax {
		msg := fmt.Sprintf("number of notifications(%d) over limit(%d)", len(reqGaurun.Notifications), ConfGaurun.Core.NotificationMax)
		LogError.Error(msg)
		sendResponse(w, msg, http.StatusBadRequest)
		return
	}

	LogError.Debug("enqueue notification")
	go enqueueNotifications(reqGaurun.Notifications)

	LogError.Debug("response to client")
	sendResponse(w, "ok", http.StatusOK)
}
