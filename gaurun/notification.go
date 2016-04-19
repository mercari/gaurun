package gaurun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/alexjlockwood/gcm"
	"github.com/cubicdaiya/apns"
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
	Badge  int          `json:"badge,omitempty"`
	Sound  string       `json:"sound,omitempty"`
	Expiry int          `json:"expiry,omitempty"`
	Retry  int          `json:"retry,omitempty"`
	Extend []ExtendJSON `json:"extend,omitempty"`
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

func InitGCMClient() {
	TransportGaurun = &http.Transport{MaxIdleConnsPerHost: ConfGaurun.Core.WorkerNum}
	GCMClient = &gcm.Sender{ApiKey: ConfGaurun.Android.ApiKey}
	GCMClient.Http = &http.Client{Transport: TransportGaurun}
	GCMClient.Http.Timeout = time.Duration(ConfGaurun.Android.Timeout) * time.Second
}

func StartPushWorkers(workerNum, queueNum int) {
	QueueNotification = make(chan RequestGaurunNotification, queueNum)
	for i := 0; i < workerNum; i++ {
		go pushNotificationWorker()
	}
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

func pushNotificationIos(req RequestGaurunNotification, client *apns.Client) bool {
	LogError.Debug("START push notification for iOS")

	for i, token := range req.Tokens {
		id := req.IDs[i]
		payload := apns.NewPayload()
		payload.Alert = req.Message
		payload.Badge = req.Badge
		payload.Sound = req.Sound

		pn := apns.NewPushNotification()
		pn.DeviceToken = token
		pn.Expiry = uint32(req.Expiry)
		pn.AddPayload(payload)

		if len(req.Extend) > 0 {
			for _, extend := range req.Extend {
				pn.Set(extend.Key, extend.Value)
			}
		}

		stime := time.Now()
		resp := client.Send(pn)
		etime := time.Now()
		ptime := etime.Sub(stime).Seconds()

		if resp.Error != nil {
			atomic.AddInt64(&StatGaurun.Ios.PushError, 1)
			LogPush(req.IDs[i], StatusFailedPush, token, ptime, req, resp.Error)
			client.Conn.Close()
			client.ConnTls.Close()
			return false
		} else {
			LogPush(id, StatusSucceededPush, token, ptime, req, nil)
			atomic.AddInt64(&StatGaurun.Ios.PushSuccess, 1)
		}
	}

	client = nil
	LogError.Debug("END push notification for iOS")
	return true
}

func pushNotificationIos2(req RequestGaurunNotification, client *http.Client) bool {
	LogError.Debug("START push notification for iOS")

	service := newApnsServiceHttp2(client)

	for i, token := range req.Tokens {
		id := req.IDs[i]
		headers := newApnsHeadersHttp2(&req)
		payload := newApnsPayloadHttp2(&req)

		stime := time.Now()
		err := apnsPushHttp2(token, service, headers, payload)

		etime := time.Now()
		ptime := etime.Sub(stime).Seconds()

		if err != nil {
			atomic.AddInt64(&StatGaurun.Ios.PushError, 1)
			LogPush(req.IDs[i], StatusFailedPush, token, ptime, req, err)
			return false
		} else {
			LogPush(id, StatusSucceededPush, token, ptime, req, nil)
			atomic.AddInt64(&StatGaurun.Ios.PushSuccess, 1)
		}
	}

	client = nil
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

func pushNotificationWorker() {
	var (
		success     bool
		retryMax    int
		ep          string
		apnsClient  *apns.Client
		apnsClient2 *http.Client
		loop        int
		err         error
	)
	if ConfGaurun.Ios.Sandbox {
		ep = EpApnsSandbox
	} else {
		ep = EpApnsProd
	}

	apnsClient = nil
	apnsClient2 = nil
	loop = 0
	for {
		stime := time.Now()

		notification := <-QueueNotification

		etime := time.Now()
		itime := etime.Sub(stime).Seconds()

		if notification.Platform == PlatFormIos {
			if !ConfGaurun.Ios.Http2 && apnsClient != nil && int(itime) > ConfGaurun.Ios.KeepAliveIdleTimeout {
				apnsClient.Conn.Close()
				apnsClient.ConnTls.Close()
				apnsClient = nil
			}

			if !ConfGaurun.Ios.Http2 && apnsClient != nil && ConfGaurun.Ios.KeepAliveMax > 0 && loop > ConfGaurun.Ios.KeepAliveMax {
				apnsClient.Conn.Close()
				apnsClient.ConnTls.Close()
				apnsClient = nil
				loop = 0
			}

			loop++

			if ConfGaurun.Ios.Http2 {
				if apnsClient2 == nil {
					apnsClient2, err = newApnsClientHttp2(
						ConfGaurun.Ios.PemCertPath,
						ConfGaurun.Ios.PemKeyPath,
					)
				}
			} else {
				if apnsClient == nil {
					apnsClient, err = apns.NewClient(
						ep,
						ConfGaurun.Ios.PemCertPath,
						ConfGaurun.Ios.PemKeyPath,
						0,
					)
					if err != nil {
						LogError.Errorf("failed to connect to APNS: %s", err.Error())
						apnsClient = nil
						loop = 0
						QueueNotification <- notification
						continue
					}
					apnsClient.TimeoutWaitError = time.Duration(ConfGaurun.Ios.TimeoutError) * time.Millisecond
				}
			}
		}

		switch notification.Platform {
		case PlatFormIos:
			if ConfGaurun.Ios.Http2 {
				success = pushNotificationIos2(notification, apnsClient2)
				if !success {
					apnsClient = nil
				}
			} else {
				success = pushNotificationIos(notification, apnsClient)
				if !success {
					apnsClient2 = nil
				}
			}
			retryMax = ConfGaurun.Ios.RetryMax
		case PlatFormAndroid:
			success = pushNotificationAndroid(notification)
			retryMax = ConfGaurun.Android.RetryMax
		}
		if !success && notification.Retry < retryMax {
			if len(QueueNotification) < cap(QueueNotification) {
				notification.Retry++
				QueueNotification <- notification
			}
		}
	}
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
