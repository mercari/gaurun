package gaurun

const (
	Version = "0.8.1"
)

const (
	PlatFormIos = iota + 1
	PlatFormAndroid

	// This is used for performance tuning or test.
	PlatformFake = 100
)

const (
	StatusAcceptedPush  = "accepted-push"
	StatusSucceededPush = "succeeded-push"
	StatusFailedPush    = "failed-push"
	StatusDisabledPush  = "disabled-push"
)
