package gaurun

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	statsGo "github.com/fukata/golang-stats-api-handler"
	"github.com/lestrrat/go-server-starter/listener"
)

func RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/push", PushNotificationHandler)
	mux.HandleFunc("/stat/app", StatsHandler)
	mux.HandleFunc("/config/app", ConfigHandler)
	mux.HandleFunc("/config/pushers", ConfigPushersHandler)

	statsGo.PrettyPrintEnabled()
	mux.HandleFunc("/stat/go", statsGo.Handler)
}

// getListener returns a listener.
func getListener(conf *ConfToml) (net.Listener, error) {
	// By default, it starts to listen a listener provided
	// by `go-server-starter`. If not, then check port defined
	// in configuration file.
	listeners, err := listener.ListenAll()
	if err != nil && err != listener.ErrNoListeningTarget {
		return nil, err
	}

	if len(listeners) > 0 {
		return listeners[0], nil
	}

	// If port is empty, nothing to listen so returns error.
	port := conf.Core.Port
	if len(port) == 0 {
		return nil, fmt.Errorf("no port to listen")
	}

	// Try to listen as TCP port, first
	if _, err := strconv.Atoi(port); err == nil {
		l, err := net.Listen("tcp", ":"+port)
		if err != nil {
			return nil, err
		}
		return l, nil
	}

	// Try to listen as UNIX socket.
	if strings.HasPrefix(port, "unix:/") {
		sockPath := port[5:]

		fi, err := os.Lstat(sockPath)
		if err == nil && (fi.Mode()&os.ModeSocket) == os.ModeSocket {
			if err := os.Remove(sockPath); err != nil {
				return nil, fmt.Errorf("failed to remove socket path: %s", err)
			}
		}

		l, err := net.Listen("unix", sockPath)
		if err != nil {
			return nil, fmt.Errorf("failed to listen unix socket: %s", err)
		}

		return l, nil
	}

	return nil, fmt.Errorf("invalid port %s (it must be number or path start with 'unix:/')", port)
}

func RunServer(server *http.Server, conf *ConfToml) error {
	l, err := getListener(conf)
	if err != nil {
		return err
	}

	return server.Serve(l)
}
