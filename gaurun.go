package main

import (
	"flag"
	"io/ioutil"
	"log"

	"./gaurun"
	"github.com/Sirupsen/logrus"
)

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
	gaurun.StartPushWorkers(gaurun.ConfGaurun.Core.WorkerNum, gaurun.ConfGaurun.Core.QueueNum)

	gaurun.RegisterHTTPHandlers()
	gaurun.RunHTTPServer()
}
