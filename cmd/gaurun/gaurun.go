package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/mercari/gaurun/gaurun"
)

func main() {
	versionPrinted := flag.Bool("v", false, "gaurun version")
	confPath := flag.String("c", "", "configuration file path for gaurun")
	listenPort := flag.String("p", "", "port number or unix socket path")
	workerNum := flag.Int("w", 0, "number of workers for push notification")
	queueNum := flag.Int("q", 0, "size of internal queue for push notification")
	flag.Parse()

	if *versionPrinted {
		gaurun.PrintVersion()
		return
	}

	// set default parameters
	gaurun.ConfGaurun = gaurun.BuildDefaultConf()

	// load configuration
	conf, err := gaurun.LoadConf(gaurun.ConfGaurun, *confPath)
	if err != nil {
		gaurun.LogSetupFatal(err)
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
	accessLogger, err := gaurun.InitLog(gaurun.ConfGaurun.Log.AccessLog)
	if err != nil {
		gaurun.LogSetupFatal(err)
	}
	errorLogger, err := gaurun.InitLog(gaurun.ConfGaurun.Log.ErrorLog)
	if err != nil {
		gaurun.LogSetupFatal(err)
	}

	if err := gaurun.SetLogLevel(accessLogger, "info"); err != nil {
		gaurun.LogSetupFatal(err)
	}
	if err := gaurun.SetLogLevel(errorLogger, gaurun.ConfGaurun.Log.Level); err != nil {
		gaurun.LogSetupFatal(err)
	}

	gaurun.LogAccess = accessLogger
	gaurun.LogError = errorLogger

	if !gaurun.ConfGaurun.Ios.Enabled && !gaurun.ConfGaurun.Android.Enabled {
		gaurun.LogSetupFatal(fmt.Errorf("What do you want to do?"))
	}

	if gaurun.ConfGaurun.Ios.Enabled {
		gaurun.CertificatePemIos.Cert, err = ioutil.ReadFile(gaurun.ConfGaurun.Ios.PemCertPath)
		if err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("A certification file for iOS is not found."))
		}

		gaurun.CertificatePemIos.Key, err = ioutil.ReadFile(gaurun.ConfGaurun.Ios.PemKeyPath)
		if err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("A key file for iOS is not found."))
		}

	}

	if gaurun.ConfGaurun.Android.Enabled {
		if gaurun.ConfGaurun.Android.ApiKey == "" {
			gaurun.LogSetupFatal(fmt.Errorf("APIKey for Android is empty."))
		}
	}

	if err := gaurun.InitHttpClient(); err != nil {
		gaurun.LogSetupFatal(fmt.Errorf("failed to init http client"))
	}
	gaurun.InitStat()
	gaurun.StartPushWorkers(gaurun.ConfGaurun.Core.WorkerNum, gaurun.ConfGaurun.Core.QueueNum)

	gaurun.RegisterHTTPHandlers()
	gaurun.RunHTTPServer()
}
