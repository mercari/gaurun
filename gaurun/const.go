package gaurun

const (
	Version = "0.13.1"
)

const (
	PlatFormIos = iota + 1
	PlatFormAndroid
)

const (
	StatusAcceptedPush  = "accepted-push"
	StatusSucceededPush = "succeeded-push"
	StatusFailedPush    = "failed-push"
	StatusDisabledPush  = "disabled-push"
)

const (
	ApnsPushTypeAlert      = "alert"
	ApnsPushTypeBackground = "background"
)
