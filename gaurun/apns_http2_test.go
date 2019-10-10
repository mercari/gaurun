package gaurun

import (
	"testing"

	"github.com/mercari/gaurun/buford/push"
	"github.com/stretchr/testify/assert"
)

func TestNewApnsClientHttp2(t *testing.T) {
	req := &RequestGaurunNotification{}
	headers := NewApnsHeadersHttp2(req)
	assert.Equal(t, push.PushTypeAlert, headers.PushType)

	req = &RequestGaurunNotification{PushType: ApnsPushTypeAlert}
	headers = NewApnsHeadersHttp2(req)
	assert.Equal(t, push.PushTypeAlert, headers.PushType)

	req = &RequestGaurunNotification{PushType: ApnsPushTypeBackground}
	headers = NewApnsHeadersHttp2(req)
	assert.Equal(t, push.PushTypeBackground, headers.PushType)
}
