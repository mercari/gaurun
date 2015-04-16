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

## Specification

See [SPEC.md](https://github.com/mercari/gaurun/blob/master/SPEC.md) about details.

## Run

```
gaurun -c conf/gaurun.toml
```

## Crash Recovery

Gaurun supports re-push notifications lost by server-crash with access.log.

```
gaurun_recovery -c conf/gaurun.toml -l /tmp/gaurun.log
```

## Commiters

 * Tatsuhiko Kubo([@cubicdaiya](https://github.com/cubicdaiya))

## License

Copyright 2014-2015 Mercari, Inc.


Licensed under the MIT License.
