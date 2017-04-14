package gaurun

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

type StatApp struct {
	QueueMax    int         `json:"queue_max"`
	QueueUsage  int         `json:"queue_usage"`
	PusherMax   int64       `json:"pusher_max"`
	PusherCount int64       `json:"pusher_count"`
	Ios         StatIos     `json:"ios"`
	Android     StatAndroid `json:"android"`
	Fake        StatFake    `json:"fake"`
}

type StatAndroid struct {
	PushSuccess int64 `json:"push_success"`
	PushError   int64 `json:"push_error"`
}

type StatIos struct {
	PushSuccess int64 `json:"push_success"`
	PushError   int64 `json:"push_error"`
}

// StatFake is stat for fake pusher.
type StatFake struct {
	PushSuccess int64 `json:"push_success"`
	PushError   int64 `json:"push_error"`
}

func InitStat() {
	StatGaurun.QueueUsage = 0
	StatGaurun.PusherCount = 0
	StatGaurun.Ios.PushSuccess = 0
	StatGaurun.Ios.PushError = 0
	StatGaurun.Android.PushSuccess = 0
	StatGaurun.Android.PushError = 0
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	var result StatApp
	result.QueueMax = cap(QueueNotification)
	result.QueueUsage = len(QueueNotification)
	result.PusherMax = ConfGaurun.Core.PusherMax * ConfGaurun.Core.WorkerNum
	result.PusherCount = atomic.LoadInt64(&PusherCountAll)
	result.Ios.PushSuccess = atomic.LoadInt64(&StatGaurun.Ios.PushSuccess)
	result.Ios.PushError = atomic.LoadInt64(&StatGaurun.Ios.PushError)
	result.Android.PushSuccess = atomic.LoadInt64(&StatGaurun.Android.PushSuccess)
	result.Android.PushError = atomic.LoadInt64(&StatGaurun.Android.PushError)

	respBody, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		msg := "Response-body could not be created"
		LogError.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Server", serverHeader())
	w.Write(respBody)
}
