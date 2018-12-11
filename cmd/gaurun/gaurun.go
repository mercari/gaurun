package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/mercari/gaurun/gaurun"
)

const (
	DefaultPidPermission = 0644
)

func main() {
	versionPrinted := flag.Bool("v", false, "gaurun version")
	confPath := flag.String("c", "", "configuration file path for gaurun")
	listenPort := flag.String("p", "", "port number or unix socket path")
	workerNum := flag.Int64("w", 0, "number of workers for push notification")
	queueNum := flag.Int64("q", 0, "size of internal queue for push notification")
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
	accessLogger, accessLogReopener, err := gaurun.InitLog(gaurun.ConfGaurun.Log.AccessLog, "info")
	if err != nil {
		gaurun.LogSetupFatal(err)
	}
	errorLogger, errorLogReopener, err := gaurun.InitLog(gaurun.ConfGaurun.Log.ErrorLog, gaurun.ConfGaurun.Log.Level)
	if err != nil {
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

	sigHUPChan := make(chan os.Signal, 1)
	signal.Notify(sigHUPChan, syscall.SIGHUP)

	sighupHandler := func() {
		if err := accessLogReopener.Reopen(); err != nil {
			gaurun.LogError.Warn(fmt.Sprintf("failed to reopen access log: %v", err))
		}
		if err := errorLogReopener.Reopen(); err != nil {
			gaurun.LogError.Warn(fmt.Sprintf("failed to reopen error log: %v", err))
		}
	}

	go signalHandler(sigHUPChan, sighupHandler)

	if len(conf.Core.Pid) > 0 {
		if _, err := os.Stat(filepath.Dir(conf.Core.Pid)); os.IsNotExist(err) {
			gaurun.LogSetupFatal(fmt.Errorf("directory for pid file is not exist: %v", err))
		} else if err := ioutil.WriteFile(conf.Core.Pid, []byte(strconv.Itoa(os.Getpid())), DefaultPidPermission); err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("failed to create a pid file: %v", err))
		}
	}

	if gaurun.ConfGaurun.Android.Enabled {
		if err := gaurun.InitGCMClient(); err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("failed to init gcm/fcm client: %v", err))
		}
	}

	if gaurun.ConfGaurun.FCMV1.Enabled {
		if err := gaurun.InitFirebaseApp(); err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("failed to init firebase app client: %v", err))
		}
	}

	if gaurun.ConfGaurun.Ios.Enabled {
		if err := gaurun.InitAPNSClient(); err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("failed to init http client for APNs: %v", err))
		}
	}

	gaurun.InitStat()
	gaurun.StartPushWorkers(gaurun.ConfGaurun.Core.WorkerNum, gaurun.ConfGaurun.Core.QueueNum)

	mux := http.NewServeMux()
	gaurun.RegisterHandlers(mux)

	server := &http.Server{
		Handler: mux,
	}
	go func() {
		gaurun.LogError.Info("start server")
		if err := gaurun.RunServer(server, &gaurun.ConfGaurun); err != nil {
			gaurun.LogError.Info(fmt.Sprintf("failed to serve: %s", err))
		}
	}()

	// Graceful shutdown (kicked by SIGTERM).
	//
	// First, it shutdowns server and stops accepting new requests.
	// Then wait until all remaining queues in buffer are flushed.
	sigTERMChan := make(chan os.Signal, 1)
	signal.Notify(sigTERMChan, syscall.SIGTERM)

	<-sigTERMChan
	gaurun.LogError.Info("shutdown server")
	timeout := time.Duration(conf.Core.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		gaurun.LogError.Error(fmt.Sprintf("failed to shutdown server: %v", err))
	}

	// Start a goroutine to log number of job queue.
	go func() {
		for {
			queue := len(gaurun.QueueNotification)
			if queue == 0 {
				break
			}

			gaurun.LogError.Info(fmt.Sprintf("wait until queue is empty. Current queue len: %d", queue))
			time.Sleep(1 * time.Second)
		}
	}()

	// Block until all pusher worker job is done.
	gaurun.PusherWg.Wait()

	gaurun.LogError.Info("successfully shutdown")
}

func signalHandler(ch <-chan os.Signal, sighupFn func()) {
	for {
		select {
		case sig := <-ch:
			switch sig {
			case syscall.SIGHUP:
				sighupFn()
			}
		}
	}
}
