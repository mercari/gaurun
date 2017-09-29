package gaurun

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterHandlers(t *testing.T) {
	mux := http.NewServeMux()

	RegisterHandlers(mux)

	entrypoints := []string{
		"/push",
		"/stat/app",
		"/config/pushers",
		"/stat/go",
	}

	for _, e := range entrypoints {
		_, pattern := mux.Handler(&http.Request{
			Method: "GET", Host: "localhost", URL: &url.URL{Path: e},
		})
		assert.Equal(t, e, pattern)
	}
}

func TestGetListener(t *testing.T) {
	validConfigs := []ConfToml{
		{Core: SectionCore{Port: "8080"}},
		{Core: SectionCore{Port: "unix:/tmp/gaurun.sock"}},
	}
	invalidConfigs := []ConfToml{
		// port is empty
		{},
		// port is not listenable
		{Core: SectionCore{Port: "100000"}},
		// port specified neither TCP port nor UNIX socket
		{Core: SectionCore{Port: "invalid:/invalid"}},
	}

	for _, c := range validConfigs {
		_, err := getListener(&c)
		assert.Nil(t, err)
	}

	for _, c := range invalidConfigs {
		_, err := getListener(&c)
		assert.NotNil(t, err)
	}
}
