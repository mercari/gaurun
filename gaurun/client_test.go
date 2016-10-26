package gaurun

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKeepAliveInterval(t *testing.T) {
	assert.Equal(t, 30, keepAliveInterval(90))
	assert.Equal(t, 30, keepAliveInterval(30))
	assert.Equal(t, 25, keepAliveInterval(25))
	assert.Equal(t, 30, keepAliveInterval(50))
	assert.Equal(t, 90, keepAliveInterval(300))
	assert.Equal(t, 90, keepAliveInterval(600))
}
