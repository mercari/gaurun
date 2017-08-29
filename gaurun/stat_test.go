package gaurun

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type GlobalsStore struct {
	ConfGaurun        ConfToml
	QueueNotification chan RequestGaurunNotification
	StatGaurun        StatApp
	PusherCountAll    int64
}

var (
	Store GlobalsStore
)

func (store GlobalsStore) Save() {
	store.StatGaurun = StatGaurun
	store.QueueNotification = QueueNotification
	store.ConfGaurun = ConfGaurun
	store.PusherCountAll = PusherCountAll
}

func (store GlobalsStore) Restore() {
	StatGaurun = store.StatGaurun
	QueueNotification = store.QueueNotification
	ConfGaurun = store.ConfGaurun
	PusherCountAll = store.PusherCountAll
}

func TestInitStat(t *testing.T) {
	Store.Save()
	defer Store.Restore()

	StatGaurun = StatApp{
		QueueMax:    1,
		PusherCount: 1,
		Ios:         StatIos{1, 1},
		Android:     StatAndroid{1, 1},
	}

	InitStat()

	cases := []struct {
		Label    string
		Expected int64
		Actual   int64
	}{
		{"StatGaurun.QueueUsage", 0, int64(StatGaurun.QueueUsage)},
		{"StatGaurun.PusherCount", 0, int64(StatGaurun.PusherCount)},
		{"StatGaurun.Ios.PushSuccess", 0, StatGaurun.Ios.PushSuccess},
		{"StatGaurun.Ios.PushError", 0, StatGaurun.Ios.PushError},
		{"StatGaurun.Android.PushSuccess", 0, StatGaurun.Android.PushSuccess},
		{"StatGaurun.Android.PushError", 0, StatGaurun.Android.PushError},
	}

	for _, c := range cases {
		assert.Equal(t, c.Actual, c.Expected)
	}
}

func TestStatsHandler(t *testing.T) {
	Store.Save()
	defer Store.Restore()

	// Set not initial values
	QueueNotification = make(chan RequestGaurunNotification, 1)
	QueueNotification <- RequestGaurunNotification{}
	ConfGaurun.Core.PusherMax = 1
	ConfGaurun.Core.WorkerNum = 1
	PusherCountAll = 1
	StatGaurun = StatApp{
		QueueMax:    1,
		PusherCount: 1,
		Ios:         StatIos{1, 1},
		Android:     StatAndroid{1, 1},
	}

	// Get dummy response
	s := httptest.NewServer(http.HandlerFunc(StatsHandler))
	defer s.Close()
	res, err := http.Get(s.URL)

	assert.Nil(t, err)
	assert.Equal(t, res.Header.Get("Content-Type"), "application/json")
	assert.Equal(t, res.Header.Get("Server"), serverHeader())

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, body, []byte(`{
 "queue_max": 1,
 "queue_usage": 1,
 "pusher_max": 1,
 "pusher_count": 1,
 "ios": {
  "push_success": 1,
  "push_error": 1
 },
 "android": {
  "push_success": 1,
  "push_error": 1
 }
}`))
}
