package gaurun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/mercari/gcm"
)

type RequestGaurun struct {
	Notifications []RequestGaurunNotification `json:"notifications"`
}

type RequestGaurunNotification struct {
	// Common
	Tokens   []string `json:"token"`
	Platform int      `json:"platform"`
	Message  string   `json:"message"`
	// Android
	CollapseKey    string `json:"collapse_key,omitempty"`
	DelayWhileIdle bool   `json:"delay_while_idle,omitempty"`
	TimeToLive     int    `json:"time_to_live,omitempty"`
	// iOS
	Badge            int          `json:"badge,omitempty"`
	Sound            string       `json:"sound,omitempty"`
	ContentAvailable bool         `json:"content_available,omitempty"`
	Expiry           int          `json:"expiry,omitempty"`
	Retry            int          `json:"retry,omitempty"`
	Extend           []ExtendJSON `json:"extend,omitempty"`
	// meta
	IDs []uint64 `json:"seq_id,omitempty"`
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

func InitHttpClient() error {
	TransportGaurun = &http.Transport{MaxIdleConnsPerHost: ConfGaurun.Core.WorkerNum}
	GCMClient = &gcm.Sender{ApiKey: ConfGaurun.Android.ApiKey}
	GCMClient.Http = &http.Client{Transport: TransportGaurun}
	GCMClient.Http.Timeout = time.Duration(ConfGaurun.Android.Timeout) * time.Second

	var err error
	APNSClient, err = NewApnsClientHttp2(
		ConfGaurun.Ios.PemCertPath,
		ConfGaurun.Ios.PemKeyPath,
	)
	if err != nil {
		return err
	}
	APNSClient.Timeout = time.Duration(ConfGaurun.Ios.Timeout) * time.Second
	return nil
}

func enqueueNotifications(notifications []RequestGaurunNotification) {
	for _, notification := range notifications {
		err := validateNotification(&notification)
		if err != nil {
			LogError.Error(err)
			continue
		}
		var enabledPush bool
		switch notification.Platform {
		case PlatFormIos:
			enabledPush = ConfGaurun.Ios.Enabled
		case PlatFormAndroid:
			enabledPush = ConfGaurun.Android.Enabled
		}
		if enabledPush {
			notification.IDs = make([]uint64, len(notification.Tokens))
			for i := 0; i < len(notification.IDs); i++ {
				notification.IDs[i] = numberingPush()
				LogPush(notification.IDs[i], StatusAcceptedPush, notification.Tokens[i], 0, notification, nil)
			}
			QueueNotification <- notification
		}
	}
}

func classifyByDevice(reqGaurun *RequestGaurun) ([]RequestGaurunNotification, []RequestGaurunNotification) {
	var (
		reqGaurunNotificationIos     []RequestGaurunNotification
		reqGaurunNotificationAndroid []RequestGaurunNotification
	)
	for _, notification := range reqGaurun.Notifications {
		switch notification.Platform {
		case PlatFormIos:
			reqGaurunNotificationIos = append(reqGaurunNotificationIos, notification)
		case PlatFormAndroid:
			reqGaurunNotificationAndroid = append(reqGaurunNotificationAndroid, notification)
		}
	}
	return reqGaurunNotificationIos, reqGaurunNotificationAndroid
}

func pushNotificationIos(req RequestGaurunNotification) bool {
	LogError.Debug("START push notification for iOS")

	service := NewApnsServiceHttp2(APNSClient)

	for i, token := range req.Tokens {
		id := req.IDs[i]
		headers := NewApnsHeadersHttp2(&req)
		payload := NewApnsPayloadHttp2(&req)

		stime := time.Now()
		err := ApnsPushHttp2(token, service, headers, payload)

		etime := time.Now()
		ptime := etime.Sub(stime).Seconds()

		if err != nil {
			atomic.AddInt64(&StatGaurun.Ios.PushError, 1)
			LogPush(req.IDs[i], StatusFailedPush, token, ptime, req, err)
			return false
		} else {
			atomic.AddInt64(&StatGaurun.Ios.PushSuccess, 1)
			LogPush(id, StatusSucceededPush, token, ptime, req, nil)
		}
	}

	LogError.Debug("END push notification for iOS")
	return true
}

func pushNotificationAndroid(req RequestGaurunNotification) bool {
	LogError.Debug("START push notification for Android")

	data := map[string]interface{}{"message": req.Message}
	if len(req.Extend) > 0 {
		for _, extend := range req.Extend {
			data[extend.Key] = extend.Value
		}
	}

	msg := gcm.NewMessage(data, req.Tokens...)
	msg.CollapseKey = req.CollapseKey
	msg.DelayWhileIdle = req.DelayWhileIdle
	msg.TimeToLive = req.TimeToLive

	stime := time.Now()
	resp, err := GCMClient.SendNoRetry(msg)
	etime := time.Now()
	ptime := etime.Sub(stime).Seconds()
	if err != nil {
		atomic.AddInt64(&StatGaurun.Android.PushError, 1)
		for i, token := range req.Tokens {
			LogPush(req.IDs[i], StatusFailedPush, token, ptime, req, err)
		}
		return false
	}

	if resp.Failure > 0 {
		atomic.AddInt64(&StatGaurun.Android.PushSuccess, int64(resp.Success))
		atomic.AddInt64(&StatGaurun.Android.PushError, int64(resp.Failure))
		if len(resp.Results) == len(req.Tokens) {
			for i, token := range req.Tokens {
				if resp.Results[i].Error != "" {
					LogPush(req.IDs[i], StatusFailedPush, token, ptime, req, errors.New(resp.Results[i].Error))
				}
			}
		}
		return true
	}

	for i, token := range req.Tokens {
		LogPush(req.IDs[i], StatusSucceededPush, token, ptime, req, nil)
	}
	atomic.AddInt64(&StatGaurun.Android.PushSuccess, int64(len(req.Tokens)))
	LogError.Debug("END push notification for Android")
	return true
}

func validateNotification(notification *RequestGaurunNotification) error {

	for _, token := range notification.Tokens {
		if len(token) == 0 {
			return errors.New("empty token")
		}
	}

	if notification.Platform < 1 || notification.Platform > 2 {
		return errors.New("invalid platform")
	}

	if len(notification.Message) == 0 {
		return errors.New("empty message")
	}

	return nil
}

func sendResponse(w http.ResponseWriter, msg string, code int) {
	var (
		respGaurun ResponseGaurun
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", serverHeader())

	w.WriteHeader(code)
	respGaurun.Message = msg
	err := json.NewEncoder(w).Encode(&respGaurun)
	if err != nil {
		// Internal Server Error(500) should be returned by right.
		// But 'code' is returned because of the limitation of json.NewEncoder and WriteHeader.
		msg := "Response-body could not be created"
		fmt.Fprintf(w, msg)
		LogError.Error(msg)
		return
	}
}

func PushNotificationHandler(w http.ResponseWriter, r *http.Request) {
	LogAcceptedRequest(ConfGaurun.Api.PushUri, r.Method, r.Proto, r.ContentLength)
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
		LogError.Debugf("parse request body: %s", reqBody)
		err = json.Unmarshal(reqBody, &reqGaurun)
	} else {
		LogError.Debug("parse request body")
		err = json.NewDecoder(r.Body).Decode(&reqGaurun)
	}

	if err != nil {
		LogError.Error(err)
		sendResponse(w, "Request-body is malformed", http.StatusBadRequest)
		return
	}

	if len(reqGaurun.Notifications) == 0 {
		LogError.Error("empty notification")
		sendResponse(w, "empty notification", http.StatusBadRequest)
		return
	} else if len(reqGaurun.Notifications) > ConfGaurun.Core.NotificationMax {
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
