package push_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/RobotsAndPencils/buford/push"
)

func TestQueuePush(t *testing.T) {
	const (
		workers = 10
		number  = 100
	)
	payload := []byte(`{ "aps" : { "alert" : "Hello HTTP/2" } }`)

	handler := http.NewServeMux()
	server := httptest.NewServer(handler)

	handler.HandleFunc("/3/device/", func(w http.ResponseWriter, r *http.Request) {
		deviceToken := strings.TrimPrefix(r.URL.String(), "/3/device/")
		// echo back the deviceToken as the id (not the real behavior)
		w.Header().Set("apns-id", deviceToken)
	})

	service := push.NewService(http.DefaultClient, server.URL)
	queue := push.NewQueue(service, workers)
	var wg sync.WaitGroup

	go func() {
		for resp := range queue.Responses {
			if resp.Err != nil {
				t.Error(resp.Err)
			}
			if resp.ID != resp.DeviceToken {
				t.Errorf("Expected %q == %q.", resp.ID, resp.DeviceToken)
			}
			wg.Done()
		}
	}()

	for i := 0; i < number; i++ {
		wg.Add(1)
		queue.Push(fmt.Sprintf("%04d", i), nil, payload)
	}
	wg.Wait()
	queue.Close()
}
