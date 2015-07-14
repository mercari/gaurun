package main

import (
	"./gaurun"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/alexjlockwood/gcm"
	"github.com/cubicdaiya/apns"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func pushNotificationAndroid(req gaurun.RequestGaurunNotification) bool {
	data := map[string]interface{}{"message": req.Message}
	msg := gcm.NewMessage(data, req.Tokens...)
	msg.CollapseKey = req.CollapseKey
	msg.DelayWhileIdle = req.DelayWhileIdle
	msg.TimeToLive = req.TimeToLive

	sender := &gcm.Sender{ApiKey: gaurun.ConfGaurun.Android.ApiKey}
	sender.Http = new(http.Client)
	sender.Http.Timeout = time.Duration(gaurun.ConfGaurun.Android.Timeout) * time.Second

	resp, err := sender.SendNoRetry(msg)
	if err != nil {
		return false
	}

	if resp.Failure > 0 {
		return true
	}

	return true
}

func pushNotificationIos(req gaurun.RequestGaurunNotification) bool {
	var ep string
	if gaurun.ConfGaurun.Ios.Sandbox {
		ep = gaurun.EpApnsSandbox
	} else {
		ep = gaurun.EpApnsProd
	}

	client, err := apns.NewClient(
		ep,
		gaurun.ConfGaurun.Ios.PemCertPath,
		gaurun.ConfGaurun.Ios.PemKeyPath,
		time.Duration(gaurun.ConfGaurun.Ios.Timeout)*time.Second,
	)
	if err != nil {
		return false
	}

	client.TimeoutWaitError = time.Duration(gaurun.ConfGaurun.Ios.TimeoutError) * time.Millisecond

	for _, token := range req.Tokens {
		payload := apns.NewPayload()
		payload.Alert = req.Message
		payload.Badge = req.Badge
		payload.Sound = req.Sound

		pn := apns.NewPushNotification()
		pn.DeviceToken = token
		pn.Expiry = uint32(req.Expiry)
		pn.AddPayload(payload)

		resp := client.Send(pn)

		if resp.Error != nil {
			// reconnect
			client.Conn.Close()
			client.ConnTls.Close()
			client, err = apns.NewClient(
				ep,
				gaurun.ConfGaurun.Ios.PemCertPath,
				gaurun.ConfGaurun.Ios.PemKeyPath,
				time.Duration(gaurun.ConfGaurun.Ios.Timeout)*time.Second,
			)
			if err != nil {
				return false
			}
			client.TimeoutWaitError = time.Duration(gaurun.ConfGaurun.Ios.TimeoutError) * time.Millisecond
		}
	}

	client.Conn.Close()
	client.ConnTls.Close()

	return true
}

func main() {
	versionPrinted := flag.Bool("v", false, "gaurun version")
	confPath := flag.String("c", "", "configuration file path for gaurun")
	logPath := flag.String("l", "", "log file path for gaurun")
	flag.Parse()

	if *versionPrinted {
		gaurun.PrintGaurunVersion()
		os.Exit(0)
	}

	// set default parameters
	gaurun.ConfGaurun = gaurun.BuildDefaultConfGaurun()

	// load configuration
	conf, err := gaurun.LoadConfGaurun(gaurun.ConfGaurun, *confPath)
	if err != nil {
		gaurun.LogError.Fatal(err)
	}
	gaurun.ConfGaurun = conf

	// set concurrency
	runtime.GOMAXPROCS(runtime.NumCPU())

	f, err := os.Open(*logPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	accepts := make(map[uint64]gaurun.LogPushEntry)
	successes := make(map[uint64]gaurun.LogPushEntry)

	for scanner.Scan() {
		var logPush gaurun.LogPushEntry
		line := scanner.Text()
		idx := strings.Index(line, " ")
		JSONStr := line[idx+1:]
		err := json.Unmarshal([]byte(JSONStr), &logPush)
		if err != nil {
			log.Printf("JSON parse error(%s)", JSONStr)
		}
		if logPush.Type == "accepted-request" {
			continue
		}

		switch logPush.Type {
		case "accepted-push":
			accepts[logPush.ID] = logPush
		case "succeeded-push":
			successes[logPush.ID] = logPush
		}
	}

	losts := make(map[uint64]gaurun.LogPushEntry)
	for id, logPush := range accepts {
		if _, ok := successes[id]; !ok {
			losts[id] = logPush
		}
	}

	done := make(chan bool, len(losts))

	for _, logPush := range losts {
		tokens := make([]string, 1)
		var platform int
		tokens = append(tokens, logPush.Token)
		switch logPush.Platform {
		case "ios":
			platform = 1
		case "android":
			platform = 2

		}

		req := &gaurun.RequestGaurunNotification{
			Tokens:         tokens,
			Platform:       platform,
			Message:        logPush.Message,
			CollapseKey:    logPush.CollapseKey,
			DelayWhileIdle: logPush.DelayWhileIdle,
			TimeToLive:     logPush.TimeToLive,
			Badge:          logPush.Badge,
			Sound:          logPush.Sound,
			Expiry:         logPush.Expiry,
		}
		go func(req *gaurun.RequestGaurunNotification, token, platform, message string) {
			var result bool
			switch logPush.Platform {
			case "ios":
				result = pushNotificationIos(*req)
			case "android":
				result = pushNotificationAndroid(*req)
			}
			if !result {
				msg := fmt.Sprintf("failed to push notification: %s %s %s", logPush.Token, logPush.Platform, logPush.Message)
				log.Println(msg)
			} else {
				msg := fmt.Sprintf("succeeded push notification: %s %s %s", token, platform, message)
				log.Println(msg)
			}
			done <- true
		}(req, logPush.Token, logPush.Platform, logPush.Message)
	}

	for i := 0; i < len(losts); i++ {
		<-done
	}

}
