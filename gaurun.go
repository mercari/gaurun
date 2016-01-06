package main

import (
	"./gaurun"
	"errors"
	"flag"
	"github.com/Sirupsen/logrus"
	statsGo "github.com/fukata/golang-stats-api-handler"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func listenUnix(port string) (net.Listener, error) {
	sockPath := port[5:]
	fi, err := os.Lstat(sockPath)
	if err == nil && (fi.Mode()&os.ModeSocket) == os.ModeSocket {
		err := os.Remove(sockPath)
		if err != nil {
			return nil, errors.New("failed to remove " + sockPath)
		}
	}
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, errors.New("failed to listen: " + sockPath)
	}

	return l, nil
}

func main() {
	versionPrinted := flag.Bool("v", false, "gaurun version")
	confPath := flag.String("c", "", "configuration file path for gaurun")
	listenPort := flag.String("p", "", "port number or unix socket path")
	workerNum := flag.Int("w", 0, "number of workers for push notification")
	queueNum := flag.Int("q", 0, "size of internal queue for push notification")
	flag.Parse()

	if *versionPrinted {
		gaurun.PrintGaurunVersion()
		return
	}

	// set concurrency
	runtime.GOMAXPROCS(runtime.NumCPU())

	// set default parameters
	gaurun.ConfGaurun = gaurun.BuildDefaultConfGaurun()

	// init logger
	gaurun.LogAccess = logrus.New()
	gaurun.LogError = logrus.New()

	gaurun.LogAccess.Formatter = new(gaurun.GaurunFormatter)
	gaurun.LogError.Formatter = new(gaurun.GaurunFormatter)

	// load configuration
	conf, err := gaurun.LoadConfGaurun(gaurun.ConfGaurun, *confPath)
	if err != nil {
		gaurun.LogError.Fatal(err)
	}
	gaurun.ConfGaurun = conf

	// overwrite if port is specified by flags
	if *listenPort != "" {
		gaurun.ConfGaurun.Core.Port = *listenPort
	}

	// overwrite if workerNum is specified by flags
	if *workerNum > 0 {
		gaurun.ConfGaurun.Core.WorkerNum = *workerNum
	}

	// overwrite if queueNum is specified by flags
	if *queueNum > 0 {
		gaurun.ConfGaurun.Core.QueueNum = *queueNum
	}

	// set logger
	err = gaurun.SetLogLevel(gaurun.LogAccess, "info")
	if err != nil {
		log.Fatal(err)
	}
	err = gaurun.SetLogLevel(gaurun.LogError, gaurun.ConfGaurun.Log.Level)
	if err != nil {
		log.Fatal(err)
	}
	err = gaurun.SetLogOut(gaurun.LogAccess, gaurun.ConfGaurun.Log.AccessLog)
	if err != nil {
		log.Fatal(err)
	}
	err = gaurun.SetLogOut(gaurun.LogError, gaurun.ConfGaurun.Log.ErrorLog)
	if err != nil {
		log.Fatal(err)
	}

	if !gaurun.ConfGaurun.Ios.Enabled && !gaurun.ConfGaurun.Android.Enabled {
		gaurun.LogError.Fatal("What do you want to do?")
	}

	if gaurun.ConfGaurun.Ios.Enabled {
		gaurun.CertificatePemIos.Cert, err = ioutil.ReadFile(gaurun.ConfGaurun.Ios.PemCertPath)
		if err != nil {
			gaurun.LogError.Fatal("A certification file for iOS is not found.")
		}

		gaurun.CertificatePemIos.Key, err = ioutil.ReadFile(gaurun.ConfGaurun.Ios.PemKeyPath)
		if err != nil {
			gaurun.LogError.Fatal("A key file for iOS is not found.")
		}

	}

	if gaurun.ConfGaurun.Android.Enabled {
		if gaurun.ConfGaurun.Android.ApiKey == "" {
			gaurun.LogError.Fatal("APIKey for Android is empty.")
		}
	}

	gaurun.InitGCMClient()
	gaurun.InitStatGaurun()
	statsGo.PrettyPrintEnabled()
	gaurun.StartPushWorkers(gaurun.ConfGaurun.Core.WorkerNum, gaurun.ConfGaurun.Core.QueueNum)

	// push and stat server
	if gaurun.ConfGaurun.Core.Port == gaurun.ConfGaurun.Stat.Port {
		http.HandleFunc(gaurun.ConfGaurun.Api.StatGoUri, statsGo.Handler)
		http.HandleFunc(gaurun.ConfGaurun.Api.StatAppUri, gaurun.StatsGaurunHandler)
		http.HandleFunc(gaurun.ConfGaurun.Api.ConfigAppUri, gaurun.ConfigGaurunHandler)
		http.HandleFunc(gaurun.ConfGaurun.Api.PushUri, gaurun.PushNotificationHandler)
		if _, err := strconv.Atoi(gaurun.ConfGaurun.Core.Port); err == nil { // Listen TCP
			http.ListenAndServe(":"+gaurun.ConfGaurun.Core.Port, nil)
		} else if strings.HasPrefix(gaurun.ConfGaurun.Core.Port, "unix:/") { // Listen Unix
			l, err := listenUnix(gaurun.ConfGaurun.Core.Port)
			if err != nil {
				log.Fatal(err.Error())
			}
			http.Serve(l, nil)
		}
		log.Fatal("core.port parameter is invalid: " + gaurun.ConfGaurun.Core.Port)
	}

	// stat server
	muxStat := http.NewServeMux()
	muxStat.HandleFunc(gaurun.ConfGaurun.Api.StatGoUri, statsGo.Handler)
	muxStat.HandleFunc(gaurun.ConfGaurun.Api.StatAppUri, gaurun.StatsGaurunHandler)
	muxStat.HandleFunc(gaurun.ConfGaurun.Api.ConfigAppUri, gaurun.ConfigGaurunHandler)

	if _, err := strconv.Atoi(gaurun.ConfGaurun.Stat.Port); err == nil {
		go http.ListenAndServe(":"+gaurun.ConfGaurun.Stat.Port, muxStat)
	} else if strings.HasPrefix(gaurun.ConfGaurun.Stat.Port, "unix:/") {
		l, err := listenUnix(gaurun.ConfGaurun.Stat.Port)
		if err != nil {
			log.Fatal(err.Error())
		}
		go http.Serve(l, muxStat)
	} else {
		log.Fatal("stat.port parameter is invalid: " + gaurun.ConfGaurun.Stat.Port)
	}

	// push server
	muxPush := http.NewServeMux()
	muxPush.HandleFunc(gaurun.ConfGaurun.Api.PushUri, gaurun.PushNotificationHandler)

	if _, err := strconv.Atoi(gaurun.ConfGaurun.Core.Port); err == nil {
		http.ListenAndServe(":"+gaurun.ConfGaurun.Core.Port, muxPush)
	} else if strings.HasPrefix(gaurun.ConfGaurun.Core.Port, "unix:/") {
		l, err := listenUnix(gaurun.ConfGaurun.Core.Port)
		if err != nil {
			log.Fatal(err.Error())
		}
		http.Serve(l, muxPush)
	}

	log.Fatal("core.port parameter is invalid: " + gaurun.ConfGaurun.Core.Port)
}
