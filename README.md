# Gaurun

![logo](https://raw.githubusercontent.com/mercari/gaurun/master/img/logo.png)

Gaurun is a general push notification server in Go.

## Status

Gaurun is production ready.

## Supported Platforms

 * [APNs](https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/ApplePushService.html)
 * [GCM](https://developers.google.com/cloud-messaging/)

## Installation

```
go get -u github.com/mercari/gaurun/...
```

## Configuration

See [CONFIGURATION.md](https://github.com/mercari/gaurun/blob/master/CONFIGURATION.md) about details.

The configuration for `gaurun` is conservative by default.
If you require higher performance of `gaurun`, you can tune the performance with some parameters such as `workers` and `queues` in the `core` section.

## Specification

See [SPEC.md](https://github.com/mercari/gaurun/blob/master/SPEC.md) about details.

## Run

```
bin/gaurun -c conf/gaurun.toml
```

## Crash Recovery

Gaurun supports re-push notifications lost by server-crash with access.log.

```
bin/gaurun_recover -c conf/gaurun.toml -l /tmp/gaurun.log
```

## Committers

 * Tatsuhiko Kubo([@cubicdaiya](https://github.com/cubicdaiya))
 * Masahiro Sano([@kazegusuri](https://github.com/kazegusuri))

## Contribution

Please read the CLA below carefully before submitting your contribution.

https://www.mercari.com/cla/

## License

Copyright 2014-2016 Mercari, Inc.


Licensed under the MIT License.
