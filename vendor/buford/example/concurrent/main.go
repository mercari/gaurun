package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/push"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	var deviceToken, filename, password, environment, host string
	var workers uint
	var number int

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file")
	flag.StringVar(&environment, "e", "development", "Environment")
	flag.UintVar(&workers, "w", 20, "Workers to send notifications")
	flag.IntVar(&number, "n", 100, "Number of notifications to send")
	flag.Parse()

	// ensure required flags are set:
	halt := false
	if deviceToken == "" {
		fmt.Println("Device token is required.")
		halt = true
	}
	if filename == "" {
		fmt.Println("Path to .p12 certificate file is required.")
		halt = true
	}
	switch environment {
	case "development":
		host = push.Development
	case "production":
		host = push.Production
	default:
		fmt.Println("Environment can be development or production.")
		halt = true
	}
	if halt {
		flag.Usage()
		os.Exit(2)
	}

	// load a certificate and use it to connect to the APN service:
	cert, err := certificate.Load(filename, password)
	exitOnError(err)

	client, err := push.NewClient(cert)
	exitOnError(err)
	service := push.NewService(client, host)
	queue := push.NewQueue(service, workers)
	var wg sync.WaitGroup

	// process responses
	// NOTE: Responses may be received in any order.
	go func() {
		count := 1
		for resp := range queue.Responses {
			if resp.Err != nil {
				log.Printf("(%d) device: %s, error: %v", count, resp.DeviceToken, resp.Err)
			} else {
				log.Printf("(%d) device: %s, apns-id: %s", count, resp.DeviceToken, resp.ID)
			}
			count++
			wg.Done()
		}
	}()

	// prepare notification(s) to send
	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
	}
	b, err := json.Marshal(p)
	exitOnError(err)

	// send notifications:
	start := time.Now()
	for i := 0; i < number; i++ {
		wg.Add(1)
		queue.Push(deviceToken, nil, b)
	}
	// done sending notifications, wait for all responses and shutdown:
	wg.Wait()
	queue.Close()
	elapsed := time.Since(start)

	log.Printf("Time for %d responses: %s (%s ea.)", number, elapsed, elapsed/time.Duration(number))
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
