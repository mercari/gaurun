# Specification for Gaurun

Gaurun is a general push notification server. It accepts HTTP requests.

## API

Gaurun APIs:

 * [POST /push](#post-push)
 * [GET /stat/go](#get-statgo)
 * [GET /stat/app](#get-statapp)
 * [PUT /config/pushers](#put-configpushers)

URI and method of each API is fixed.

### POST /push

Accepts the HTTP request for push notifications and pushes notifications asynchronously.

The JSON below is the request-body example.

```json
{
    "notifications" : [
        {
            "token" : ["xxx"],
            "platform" : 1,
            "message" : "Hello, iOS!",
            "title": "Greeting",
            "subtitle": "greeting",
            "badge" : 1,
            "category": "category1",
            "sound" : "default",
            "content_available" : false,
            "mutable_content" : false,
            "expiry" : 10,
            "extend" : [{ "key": "url", "val": "..." }, { "key": "intent", "val": "..." }]
        },
        {
            "token" : ["yyy"],
            "platform" : 2,
            "message" : "Hello, Android!",
            "collapse_key" : "update",
            "delay_while_idle" : true,
            "time_to_live" : 10,
            "priority" : "normal"
        }
    ]
}
```

The request-body must have the `notifications` array. Table below shows the parameters of each notification:

|name             |type        |description                              |required|default|note                                      |
|-----------------|------------|-----------------------------------------|--------|-------|------------------------------------------|
|token            |string array|device tokens                            |o       |       |                                          |
|platform         |int         |platform(iOS, Android)                   |o       |       |1=iOS, 2=Android                          |
|message          |string      |message for notification                 |-       |       |                                          |
|title            |string      |title for notification                   |-       |       |only iOS                                  |
|subtitle         |string      |subtitle for notification                |-       |       |only iOS                                  |
|badge            |int         |badge count                              |-       |0      |only iOS                                  |
|category         |string      |unnotification category                  |-       |       |only iOS                                  |
|sound            |string      |sound type                               |-       |       |only iOS                                  |
|expiry           |int         |expiration for notification              |-       |0      |only iOS.                                 |
|content_available|bool        |indicate that new content is available   |-       |false  |only iOS.                                 |
|mutable_content  |bool        |enable Notification Service app extension|-       |false  |only iOS(10.0+)                           |
|collapse_key     |string      |the key for collapsing notifications     |-       |       |only Android                              |
|delay_while_idle |bool        |the flag for device idling               |-       |false  |only Android                              |
|time_to_live     |int         |expiration of message kept on FCM storage|-       |0      |only Android                              |
|priority         |string      |deliver immediately or save battery ( high or normal)      |-       |normal   |only Android        | 
|extend           |string array|extensible partition                     |-       |       |                                          |
|identifier        |string      |notification identifier                    |-       |       |an optional value to identify notification|
|push_type        |string      |apns-push-type                           |-       |alert  |only iOS(13.0+)                           |

The JSON below is the response-body example from Gaurun. In this case, the status is 200(OK).

```json
{
    "message" : "ok",
}
```

When Gaurun receives an invalid request(for example: malformed body), the status of response it returns is 400(Bad Request).


### GET /stat/go

Returns the statistics for Golang-runtime. See [golang-stats-api-handler](https://github.com/fukata/golang-stats-api-handler) about details.

### GET /stat/app

Returns the statistics for Gaurun. The JSON below is an example:

```json
{
    "queue_max": 8192,
    "queue_usage": 9,
    "pusher_max": 16,
    "pusher_count": 0,
    "ios": {
        "push_success": 2759,
        "push_error": 10
    },
    "android": {
        "push_success": 2985,
        "push_error": 35
    }
}
```

Table below shows the parameters:

|name        |description                                          |note       |
|------------|-----------------------------------------------------|-----------|
|queue_max   |size of internal queue for push notification         |           |
|queue_usage |usage of internal queue for push notification        |           |
|pusher_max  |maximum number of goroutines for asynchronous pushing|           |
|pusher_count|current number of goroutines for asynchronous pushing|           |
|push_success|number of succeeded push notifications               |           |
|push_error  |number of failed push notifications                  |           |

### PUT /config/pushers

Adjusts the `core.pusher_max`. Give the new value of `core.pusher_max` to `PUT /config/pushers` with the parameter `max` like below.

```
/config/pushers?max=24
```

**Note**: Do not give too large value.
