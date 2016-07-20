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
|port            |string|port number or unix socket path             |1056            |e.g.)1056, unix:/tmp/gaurun.sock <br/> `-p` option can overwrite    |
|workers         |int   |number of workers for push notification     |runtime.NumCPU()|`-w` options can overwrite          |
|queues          |int   |size of internal queue for push notification|8192            |`-q` options can overwrite          |
|notification_max|int   |limit of push notifications once            |100             |                                    |
|pusher_max      |int64 |maximum goroutines for asynchronous pushing |0               |If the value is zero, each worker pushes synchronously|

## API Section

|name          |type  |description                          |default    |note|
|--------------|------|-------------------------------------|-----------|----|
|push_uri      |string|URI for push notification            |/push      |    |
|stat_go_uri   |string|URI for statictics for golang-runtime|/stat/go   |    |
|stat_app_uri  |string|URI for statictics for Gaurun        |/stat/app  |    |
|config_app_uri|string|URI for view configuration for Gaurun|/config/app|    |

See [SPEC.md](SPEC.md) about details for APIs.

## iOS Section

|name                  |type  |description                                           |default   |note                           |
|----------------------|------|------------------------------------------------------|----------|-------------------------------|
|enabled               |bool  |On/Off for push notication to APNs                    |true      |                               |
|pem_cert_path         |string|certification file path for APNs                      |          |                               |
|pem_key_path          |string|secret key file path for APNs                         |          |                               |
|sandbox               |bool  |On/Off for sandbox environment                        |true      |                               |
|retry_max             |int   |maximum retry count for push notication to APNs       |1         |                               |
|timeout               |int   |timeout for push notification to APNs                 |5         |                               |
|keepalive_timeout     |int   |time for continuing keep-alive connection to APNs     |30        |                               |
|keepalive_conns       |int   |number of keep-alive connection to APNs               |runtime.NumCPU()|                         |
|topic                 |string|the assigned value of `apns-topic` for Request headers|          |                               |

`topic` is mandatory when the client is connected using a certificate that supports multiple topics.

## Android Section

|name             |type  |description                                     |default|note|
|-----------------|------|------------------------------------------------|-------|----|
|enabled          |bool  |On/Off for push notication to GCM               |true   |    |
|apikey           |string|API key string for GCM                          |       |    |
|timeout          |int   |timeout for push notication to GCM              |5(sec) |    |
|keepalive_timeout|int   |time for continuing keep-alive connection to GCM|30     |    |
|keepalive_conns  |int   |number of keep-alive connection to GCM          |runtime.NumCPU()||
|retry_max        |int   |maximum retry count for push notication to GCM  |1      |    |

## Log Section

|name      |type  |description    |default|note                             |
|----------|------|---------------|-------|---------------------------------|
|access_log|string|access log path|stdout |                                 |
|error_log |string|error log path |stderr |                                 |
|level     |string|log level      |error  |panic,fatal,error,warn,info,debug|

`access_log` and `error_log` are allowed to give not only file-path but `stdout` and `stderr`.
