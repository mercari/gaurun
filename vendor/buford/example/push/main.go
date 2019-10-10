package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"
)

func main() {
	var deviceToken, filename, password, environment, host string

	flag.StringVar(&deviceToken, "d", "", "Device token")
	flag.StringVar(&filename, "c", "", "Path to .p12 certificate file.")
	flag.StringVar(&password, "p", "", "Password for .p12 file.")
	flag.StringVar(&environment, "e", "development", "Environment")
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

	// construct a payload to send to the device:
	p := payload.APS{
		Alert: payload.Alert{Body: "Hello HTTP/2"},
		Badge: badge.New(42),
	}
	b, err := json.Marshal(p)
	exitOnError(err)

	// push the notification:
	id, err := service.Push(deviceToken, nil, b)
	exitOnError(err)

	fmt.Println("apns-id:", id)
}

func exitOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
