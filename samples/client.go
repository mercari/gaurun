package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	CollapseKey    string `json:"collapse_key"`
	DelayWhileIdle bool   `json:"data_while_idle"`
	TimeToLive     int    `json:"time_to_live"`
	// iOS
	Title            string `json:"title"`
	Subtitle         string `json:"subtitle"`
	Badge            int    `json:"badge"`
	Category         string `json:"category"`
	Sound            string `json:"sound"`
	ContentAvailable bool   `json:"content_available"`
	Expiry           int    `json:"expiry"`
}

func main() {
	host := flag.String("s", "127.0.0.1:1056", "gaurun server")
	iOSToken := flag.String("i", "", "device token for APNs")
	androidToken := flag.String("a", "", "device token for Android")
	flag.Parse()

	i := 0
	c := 0
	if *iOSToken != "" {
		c++
	}

	if *androidToken != "" {
		c++
	}

	if c == 0 {
		flag.PrintDefaults()
		return
	}

	// build request body
	var req RequestGaurun
	req.Notifications = make([]RequestGaurunNotification, c)
	if *iOSToken != "" {
		req.Notifications[i].Tokens = append(req.Notifications[i].Tokens, *iOSToken)
		req.Notifications[i].Platform = 1
		req.Notifications[i].Message = "Hello, iOS!"
		req.Notifications[i].Title = "Greeting"
		req.Notifications[i].Subtitle = "greeting"
		req.Notifications[i].Badge = 1
		req.Notifications[i].Category = "category1"
		req.Notifications[i].Sound = "default"
		req.Notifications[i].ContentAvailable = true
		i++
	}

	if *androidToken != "" {
		req.Notifications[i].Tokens = append(req.Notifications[i].Tokens, *androidToken)
		req.Notifications[i].Platform = 2
		req.Notifications[i].Message = "Hello, Android!"
	}

	reqJson, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	reqBody := strings.NewReader(string(reqJson))

	// send push request to gaurun
	resp, err := http.Post("http://"+*host+"/push", "application/json", reqBody)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.Println("status: " + resp.Status)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("response body:" + string(respBody))
}
