# Configuration for Gaurun

A configuration file format for Gaurun is [TOML](https://github.com/toml-lang/toml).

A configuration for Gaurun has some sections. A example is [here](conf/gaurun.toml).

 * [Core Section](#core-section)
 * [API Section](#api-section)
 * [iOS Section](#ios-section)
 * [Android Section](#android-section)
 * [Log Section](#log-section)

## Core Section

|name            |type  |description                                 |default         |note                                |
|----------------|------|--------------------------------------------|----------------|------------------------------------|
|port            |string|port number or unix socket path             |1056            |e.g.)1056, unix:/tmp/gaurun.sock    |
|workers         |int   |number of workers for push notification     |runtime.NumCPU()|                                    |
|queues          |int   |size of internal queue for push notification|8192            |                                    |
|notificatoin_max|int   |limit of push notifications once            |100             |                                    |

## API Section

|name          |type  |description                          |default    |note|
|--------------|------|-------------------------------------|-----------|----|
|push_uri      |string|URI for push notification            |/push      |    |
|stat_go_uri   |string|URI for statictics for golang-runtime|/stat/go   |    |
|stat_app_uri  |string|URI for statictics for Gaurun        |/stat/app  |    |
|config_app_uri|string|URI for view configuration for Gaurun|/config/app|    |

See [SPEC.md](SPEC.md) about details for APIs.

## iOS Section

|name                  |type  |description                                      |default   |note                           |
|----------------------|------|-------------------------------------------------|----------|-------------------------------|
|enabled               |bool  |On/Off for push notication to APNS               |true      |                               |
|pem_cert_path         |string|certification file path for APNS                 |          |                               |
|pem_key_path          |string|secret key file path for APNS                    |          |                               |
|sandbox               |bool  |On/Off for sandbox environment                   |true      |                               |
|retry_max             |int   |maximum retry count for push notication to APNS  |1         |                               |
|timeout_error         |int   |timeout for waiting error message from APNS      |500(msec) |                               |
|keepalive_max         |int   |try-counts for each keepalive connection         |0         |zero makes unlimited           |
|keepalive_idle_timeout|int   |timeout for idleling keepalive connectio for APNS|300       |                               |

The value of `timeout` should be zero in production.

## Android Section

|name         |type  |description                                   |default|note|
|-------------|------|----------------------------------------------|-------|----|
|enabled      |bool  |On/Off for push notication to GCM             |true   |    |
|apikey       |string|API key string for GCM                        |       |    |
|timeout      |int   |timeout for push notication to GCM            |5(sec) |    |
|retry_max    |int   |maximum retry count for push notication to GCM|1      |    |

## Log Section

|name      |type  |description    |default|note                             |
|----------|------|---------------|-------|---------------------------------|
|access_log|string|access log path|stdout |                                 |
|error_log |string|error log path |stderr |                                 |
|level     |string|log level      |error  |panic,fatal,error,warn,info,debug|

`access_log` and `error_log` are allowed to give not only file-path but `stdout` and `stderr`.
