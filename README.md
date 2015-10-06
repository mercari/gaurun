# Gaurun

Gaurun is a general push notification server in Go.

## Status

Gaurun is production ready.

## Supported Platforms

 * [APNS](https://developer.apple.com/library/ios/documentation/networkinginternet/conceptual/remotenotificationspg/Chapters/ApplePushService.html)
 * [GCM](https://developer.android.com/google/gcm/index.html)

## Installation

```
make gom
make bundle
make
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

## License

Copyright 2014-2015 Mercari, Inc.


Licensed under the MIT License.
