package main

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/mercari/gaurun/buford/token"
	"github.com/mercari/gaurun/gaurun"
	"github.com/mercari/gaurun/gcm"
)

var (
	APNSClient gaurun.APNsClient
	GCMClient  *gcm.Client
)

func pushNotification(wg *sync.WaitGroup, req gaurun.RequestGaurunNotification, logPush gaurun.LogPushEntry) {
	var result bool
	switch logPush.Platform {
	case "ios":
		result = pushNotificationIos(req)
	case "android":
		result = pushNotificationAndroid(req)
	}
	if !result {
		msg := fmt.Sprintf("failed to push notification: %s %s %s", logPush.Token, logPush.Platform, logPush.Message)
		log.Println(msg)
	} else {
		msg := fmt.Sprintf("succeeded push notification: %s %s %s", logPush.Token, logPush.Platform, logPush.Message)
		log.Println(msg)
	}

	wg.Done()
}

func pushNotificationAndroid(req gaurun.RequestGaurunNotification) bool {
	data := map[string]interface{}{"message": req.Message}
	msg := gcm.NewMessage(data, req.Tokens...)
	msg.CollapseKey = req.CollapseKey
	msg.DelayWhileIdle = req.DelayWhileIdle
	msg.TimeToLive = req.TimeToLive
	msg.Priority = req.Priority

	_, err := GCMClient.Send(msg)
	if err != nil {
		return false
	}

	return true
}

func pushNotificationIos(req gaurun.RequestGaurunNotification) bool {

	service := gaurun.NewApnsServiceHttp2(APNSClient)

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
		gaurun.PrintVersion()
		os.Exit(0)
	}

	// set default parameters
	gaurun.ConfGaurun = gaurun.BuildDefaultConf()

	// load configuration
	conf, err := gaurun.LoadConf(gaurun.ConfGaurun, *confPath)
	if err != nil {
		gaurun.LogSetupFatal(err)
	}
	gaurun.ConfGaurun = conf

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
		err := json.Unmarshal([]byte(line), &logPush)
		if err != nil {
			log.Printf("JSON parse error(%s)", line)
			continue
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

	if gaurun.ConfGaurun.Ios.IsCertificateBasedProvider() {
		APNSClient, err = gaurun.NewApnsClientHttp2(
			gaurun.ConfGaurun.Ios.PemCertPath,
			gaurun.ConfGaurun.Ios.PemKeyPath,
			gaurun.ConfGaurun.Ios.PemKeyPassphrase,
		)
	} else if gaurun.ConfGaurun.Ios.IsTokenBasedProvider() {
		var authKey *ecdsa.PrivateKey
		authKey, err = token.AuthKeyFromFile(gaurun.ConfGaurun.Ios.TokenAuthKeyPath)
		if err != nil {
			gaurun.LogSetupFatal(err)
		}
		APNSClient, err = gaurun.NewApnsClientHttp2ForToken(
			authKey,
			gaurun.ConfGaurun.Ios.TokenAuthKeyID,
			gaurun.ConfGaurun.Ios.TokenAuthTeamID,
		)
	} else {
		gaurun.LogSetupFatal(fmt.Errorf("should be specify Token-based provider or Certificate-based provider"))
	}
	if err != nil {
		gaurun.LogSetupFatal(err)
	}
	APNSClient.HTTPClient.Timeout = time.Duration(gaurun.ConfGaurun.Ios.Timeout) * time.Second

	GCMClient, err := gcm.NewClient(gcm.FCMSendEndpoint, gaurun.ConfGaurun.Android.ApiKey)
	if err != nil {
		gaurun.LogSetupFatal(err)
	}

	TransportGCM := &http.Transport{
		MaxIdleConnsPerHost: gaurun.ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(gaurun.ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: time.Duration(gaurun.ConfGaurun.Android.KeepAliveTimeout) * time.Second,
		}).Dial,
	}

	GCMClient.Http = &http.Client{
		Transport: TransportGCM,
		Timeout:   time.Duration(gaurun.ConfGaurun.Android.Timeout) * time.Second,
	}

	wg := new(sync.WaitGroup)
	for _, logPush := range losts {
		tokens := make([]string, 1)
		var platform int
		tokens[0] = logPush.Token
		switch logPush.Platform {
		case "ios":
			platform = 1
		case "android":
			platform = 2

		}

		req := gaurun.RequestGaurunNotification{
			Tokens:           tokens,
			Platform:         platform,
			Message:          logPush.Message,
			CollapseKey:      logPush.CollapseKey,
			DelayWhileIdle:   logPush.DelayWhileIdle,
			TimeToLive:       logPush.TimeToLive,
			Title:            logPush.Title,
			Subtitle:         logPush.Subtitle,
			Badge:            logPush.Badge,
			Category:         logPush.Category,
			Sound:            logPush.Sound,
			ContentAvailable: logPush.ContentAvailable,
			Expiry:           logPush.Expiry,
		}
		wg.Add(1)
		go pushNotification(wg, req, logPush)
	}

	wg.Wait()
}
