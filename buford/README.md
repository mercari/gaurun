# Buford

Apple Push Notification (APN) Provider library for Go 1.6 and HTTP/2. Send remote notifications to iOS, macOS, tvOS and watchOS. Buford can also sign push packages for Safari notifications and Wallet passes.

Please see [releases](https://github.com/RobotsAndPencils/buford/releases) for updates.

[![GoDoc](https://godoc.org/github.com/RobotsAndPencils/buford?status.svg)](https://godoc.org/github.com/RobotsAndPencils/buford) [![Build Status](https://travis-ci.org/RobotsAndPencils/buford.svg?branch=ci)](https://travis-ci.org/RobotsAndPencils/buford) ![MIT](https://img.shields.io/badge/license-MIT-blue.svg) [![codecov](https://codecov.io/gh/RobotsAndPencils/buford/branch/master/graph/badge.svg)](https://codecov.io/gh/RobotsAndPencils/buford)

### Documentation

Buford uses Apple's new HTTP/2 Notification API that was announced at WWDC 2015 and [released on December 17, 2015](https://developer.apple.com/news/?id=12172015b).

[API documentation](https://godoc.org/github.com/RobotsAndPencils/buford/) is available from GoDoc.

Also see Apple's [Local and Remote Notification Programming Guide][notification], especially the sections on the JSON [payload][] and the [Notification API][notification-api].

[notification]: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/Introduction.html
[payload]: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/TheNotificationPayload.html#//apple_ref/doc/uid/TP40008194-CH107-SW1
[notification-api]: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/APNsProviderAPI.html#//apple_ref/doc/uid/TP40008194-CH101-SW1

#### Terminology

**APN** Apple Push Notification

**Provider** The Buford library is used to create a _provider_ of push notifications.

**Service** Apple's push notification service that Buford communicates with.

**Client** An `http.Client` provides an HTTP/2 client to communicate with the APN Service.

**Notification** A payload, device token, and headers.

**Device Token** An identifier for an application on a given device.

**Payload** The JSON sent to a device.

**Headers** HTTP/2 headers are used to set priority and expiration.

### Installation

This library requires [Go 1.6.3](https://golang.org/dl/) or better.

```
go get -u -d github.com/RobotsAndPencils/buford
```

Buford depends on several packages outside of the standard library, including the http2 package. Its certificate package depends on the pkcs12 and pushpackage depends on pkcs7. They can be retrieved or updated with:

```
go get -u golang.org/x/net/http2
go get -u golang.org/x/crypto/pkcs12
go get -u github.com/aai/gocrypto/pkcs7
```

I am still looking for feedback on the API so it may change. Please copy Buford and its dependencies into a `vendor/` folder at the root of your project.

### Examples

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/payload/badge"
	"github.com/RobotsAndPencils/buford/push"
)

// set these variables appropriately
const (
	filename = "/path/to/certificate.p12"
	password = ""
	host = push.Development
	deviceToken = "c2732227a1d8021cfaf781d71fb2f908c61f5861079a00954a5453f1d0281433"
)

func main() {
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
```

See `example/push` for the complete listing.

#### Concurrent use

HTTP/2 can send multiple requests over a single connection, but `service.Push` waits for a response before returning. Instead, you can wrap a `Service` in a queue to handle responses independently, allowing you to send multiple notifications at once.

```go
var wg sync.WaitGroup
queue := push.NewQueue(service, numWorkers)

// process responses (responses may be received in any order)
go func() {
	for resp := range queue.Responses {
		log.Println(resp)
		// done receiving and processing one response
		wg.Done()
	}
}()

// send the notifications
for i := 0; i < 100; i++ {
	// increment count of notifications sent and queue it
	wg.Add(1)
	queue.Push(deviceToken, nil, b)
}

// wait for all responses to be processed
wg.Wait()
// shutdown the channels and workers for the queue
queue.Close()
```

It's important to set up a goroutine to handle responses before sending any notifications, otherwise Push will block waiting for room to return a Response.

You can configure the number of workers used to send notifications concurrently, but be aware that a larger number isn't necessarily better, as Apple limits the number of concurrent streams. From the Apple Push Notification documentation:

> "The APNs server allows multiple concurrent streams for each connection. The exact number of streams is based on server load, so do not assume a specific number of streams."

See `example/concurrent/` for a complete listing.

#### Headers

You can specify an ID, expiration, priority, and other parameters via the Headers struct.

```go
headers := &push.Headers{
	ID:          "922D9F1F-B82E-B337-EDC9-DB4FC8527676",
	Expiration:  time.Now().Add(time.Hour),
	LowPriority: true,
}

id, err := service.Push(deviceToken, headers, b)
```

If no ID is specified APNS will generate and return a unique ID. When an expiration is specified, APNS will store and retry sending the notification until that time, otherwise APNS will not store or retry the notification. LowPriority should always be set when sending a ContentAvailable payload.

#### Custom values

To add custom values to an APS payload, use the Map method as follows:

```go
p := payload.APS{
	Alert: payload.Alert{Body: "Message received from Bob"},
}
pm := p.Map()
pm["acme2"] = []string{"bang", "whiz"}

b, err := json.Marshal(pm)
if err != nil {
	log.Fatal(b)
}

id, err := service.Push(deviceToken, nil, b)
```

#### Error responses

Errors from `service.Push` or `queue.Response` could be HTTP errors or an error response from Apple. To access the Reason and HTTP Status code, you must convert the `error` to a `push.Error` as follows:

```go
if e, ok := err.(*push.Error); ok {
	switch e.Reason {
	case push.ErrBadDeviceToken:
		// handle error
	}
}
```

### Website Push

Before you can send push notifications through Safari and the Notification Center, you must provide a push package, which is a signed zip file containing some JSON and icons.

Use `pushpackage` to write a zip to a `http.ResponseWriter` or to a file. It will create the `manifest.json` and `signature` files for you.

```go
pkg := pushpackage.New(w)
pkg.EncodeJSON("website.json", website)
pkg.File("icon.iconset/icon_128x128@2x.png", "static/icon_128x128@2x.png")
// other icons... (required)
if err := pkg.Sign(cert, nil); err != nil {
	log.Fatal(err)
}
```

NOTE: The filenames added to the zip may contain forward slashes but not back slashes or drive letters.

See `example/website/` and the [Safari Push Notifications][safari] documentation.

[safari]: https://developer.apple.com/library/mac/documentation/NetworkingInternet/Conceptual/NotificationProgrammingGuideForWebsites/PushNotifications/PushNotifications.html#//apple_ref/doc/uid/TP40013225-CH3-SW12

### Wallet (Passbook) Pass

A pass is a signed zip file with a .pkpass extension and a `application/vnd.apple.pkpass` MIME type. You can use `pushpackage` to write a .pkpass that contains a `pass.json` file.

See `example/wallet/` and the [Wallet Developer Guide][wallet].

[wallet]: https://developer.apple.com/library/prerelease/ios/documentation/UserExperience/Conceptual/PassKit_PG/index.html

### Related Projects

* [apns2](https://github.com/sideshow/apns2) Alternative HTTP/2 APN provider library (Go)
* [go-apns-server](https://github.com/CleverTap/go-apns-server) Mock APN server (Go)
* [gorush](https://github.com/appleboy/gorush) A push notification server (Go)
* [Push Encryption](https://github.com/GoogleChrome/push-encryption-go) Web Push for Chrome and Firefox (Go)
* [micromdm](https://micromdm.io/) Mobile Device Management server (Go)
* [Lowdown](https://github.com/alloy/lowdown) (Ruby)
* [Apnotic](https://github.com/ostinelli/apnotic) (Ruby)
* [Pigeon](https://github.com/codedge-llc/pigeon) (Elixir, iOS and Android)
* [APNSwift](https://github.com/kaunteya/APNSwift) (Swift)
