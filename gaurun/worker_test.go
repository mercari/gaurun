package gaurun

import (
	"errors"
	"testing"

	"github.com/RobotsAndPencils/buford/push"
	"github.com/stretchr/testify/assert"
)

func TestIsExternalServerError(t *testing.T) {
	cases := []struct {
		Err      error
		Platform int
		Expected bool
	}{
		{push.ErrIdleTimeout, PlatFormIos, true},
		{push.ErrShutdown, PlatFormIos, true},
		{push.ErrInternalServerError, PlatFormIos, true},
		{push.ErrServiceUnavailable, PlatFormIos, true},
		{errors.New("no error"), PlatFormIos, false},

		{errors.New("Unavailable"), PlatFormAndroid, true},
		{errors.New("InternalServerError"), PlatFormAndroid, true},
		{errors.New("Timeout"), PlatFormAndroid, true},
		{errors.New("no error"), PlatFormAndroid, false},

		{errors.New("no error"), 100 /* neither iOS nor Android */, false},
	}

	for _, c := range cases {
		actual := isExternalServerError(c.Err, c.Platform)
		assert.Equal(t, actual, c.Expected)
	}
}
