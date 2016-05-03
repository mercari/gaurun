package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/alexjlockwood/gcm"
	"github.com/mercari/gaurun/gaurun"
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

func pushNotificationIos(client *http.Client, req gaurun.RequestGaurunNotification) bool {

	service := gaurun.NewApnsServiceHttp2(client)

	for _, token := range req.Tokens {

		headers := gaurun.NewApnsHeadersHttp2(&req)
		payload := gaurun.NewApnsPayloadHttp2(&req)

		err := gaurun.ApnsPushHttp2(token, service, headers, payload)
		if err != nil {
			return false
		}
	}

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
		log.Fatal(err)
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

	apnsClient, err := gaurun.NewApnsClientHttp2(
		gaurun.ConfGaurun.Ios.PemCertPath,
		gaurun.ConfGaurun.Ios.PemKeyPath,
	)
	if err != nil {
		log.Fatal(err)
	}
	apnsClient.Timeout = time.Duration(gaurun.ConfGaurun.Ios.Timeout) * time.Second

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
				result = pushNotificationIos(apnsClient, *req)
			case "android":
				result = pushNotificationAndroid(*req)
			}
			if !result {
				msg := fmt.Sprintf("failed to push notification: %s %s %s", token, platform, message)
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
