# Specification for Gaurun

Gaurun is a general push notification server. It accepts a HTTP request.

## API

Gaurun has some APIs.

 * [POST /push](#post-push)
 * [GET /stat/go](#get-statgo)
 * [GET /stat/app](#get-statapp)
 * [GET /config/app](#get-configapp)

URI of each API is configurable. But the method is fixed.

### POST /push

Accepts a HTTP request for push notifications and pushes notifications asynchronously.

The JSON below is the request-body example.

```json
{
    "notifications" : [
        {
            "token" : ["xxx"],
            "platform" : 1,
            "message" : "Hello, iOS!",
            "badge" : 1,
            "sound" : "default",
            "content_available" : true,
            "expiry" : 10,
            "extend" : [{ "key": "url", "val": "..." }, { "key": "intent", "val": "..." }]
        },
        {
            "token" : ["yyy"],
            "platform" : 2,
            "message" : "Hello, Android!",
            "collapse_key" : "update",
            "delay_while_idle" : true,
            "time_to_live" : 10
        }
    ]
}
```

The request-body must has the `notifications` array. There is the parameter table for each notification below.

|name             |type        |description                              |required|note            |
|-----------------|------------|-----------------------------------------|--------|----------------|
|token            |string array|device tokens                            |o       |                |
|platform         |int         |platform(iOS,Android)                    |o       |1=iOS, 2=Android|
|message          |string      |message for notification                 |o       |                |
|badge            |int         |badge count                              |-       |only iOS        |
|sound            |string      |sound type                               |-       |only iOS        |
|expiry           |int         |expiration for notification              |-       |only iOS        |
|content_available|bool        |indicate that new content is available   |-       |only iOS        |
|collapse_key     |string      |the key for collapsing notifications     |-       |only Android    |
|delay_while_idle |bool        |the flag for device idling               |-       |only Android    |
|time_to_live     |int         |expiration of message kept on GCM storage|-       |only Android    |
|extend           |string array|extensible partition                     |-       |                |

The JSON below is a response-body example from Gaurun. In this case, a status is 200(OK).

```json
{
    "message" : "ok",
}
```

When Gaurun receives an invalid request(for example, malformed body is included), a status of response it returns is 400(Bad Request).


### GET /stat/go

Returns the statictics for golang-runtime. See [golang-stats-api-handler](https://github.com/fukata/golang-stats-api-handler) about details.

### GET /stat/app

Returns the statictics for Gaurun. The JSON below is the example.

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

|name        |description                                          |note       |
|------------|-----------------------------------------------------|-----------|
|queue_max   |size of internal queue for push notification         |           |
|queue_usage |usage of internal queue for push notification        |           |
|pusher_max  |maximum number of goroutines for asynchronous pushing|           |
|pusher_count|current number of goroutines for asynchronous pushing|           |
|push_success|number of succeeded push notification                |           |
|push_error  |number of failed push notification                   |           |

### GET /config/app

Returns the current configuration for Gaurun.
