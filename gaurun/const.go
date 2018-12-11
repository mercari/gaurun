package gaurun

const (
	Version = "0.10.0"
)

const (
	PlatFormIos = iota + 1
	PlatFormAndroid
	PlatFormFCMV1
)

const (
	StatusAcceptedPush  = "accepted-push"
	StatusSucceededPush = "succeeded-push"
	StatusFailedPush    = "failed-push"
	StatusDisabledPush  = "disabled-push"
)
