# Specification for Gaurun

Gaurun is a general push notificaiton server. It accepts a HTTP request.

## API

Gaurun has some APIs.

 * [POST /push](#post-push)
 * [GET /stat/go](#get-statgo)
 * [GET /stat/app](#get-statapp)
 * [GET /config/app](#get-configapp)

URI of each API is configurable. But a method is fixed.

### POST /push

Accepts a HTTP request for push notifications and pushes notifications asynchronously.

The following JSON is a request-body example.

```json
{
    "notifications" : [
        {
            "token" : ["xxx"],
            "platform" : 1,
            "message" : "Hello, iOS!",
            "badge" : 1,
            "sound" : "default",
            "expiry" : 10
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

A request-body must has a `notifications` array. The following is a paramter table for each notification.

|name            |type        |description                              |required|note            |
|----------------|------------|-----------------------------------------|--------|----------------|
|token           |string array|device tokens                            |o       |                |
|platform        |int         |platform(iOS,Android)                    |o       |1=iOS, 2=Android|
|message         |string      |message for notification                 |o       |                |
|badge           |int         |badge count                              |-       |only iOS        |
|sound           |string      |sound type                               |-       |only iOS        |
|expiry          |int         |expiration for notification              |-       |only iOS        |
|collapse_key    |string      |a key for collasing notifications        |-       |only Android    |
|delay_while_idle|bool        |a flag for device idling                 |-       |only Android    |
|time_to_live    |int         |expiration of message kept on GCM storage|-       |only Android    |

The following JSON is a response-body example from Gaurun. In this case, a status is 200(OK).

```json
{
    "message" : "ok",
}
```

When Gaurun receives an invalid request(for example, malformed body is included), a status of response it returns is 400(Bad Request).


### GET /stat/go

Returns a statictics for golang-runtime. See [golang-stats-api-handler](https://github.com/fukata/golang-stats-api-handler) about details.

### GET /stat/app

Returns a statictics for Gaurun.

### GET /config/app

Returns a current configuration for Gaurun.
