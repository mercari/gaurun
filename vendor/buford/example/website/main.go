package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/RobotsAndPencils/buford/certificate"
	"github.com/RobotsAndPencils/buford/payload"
	"github.com/RobotsAndPencils/buford/push"
	"github.com/RobotsAndPencils/buford/pushpackage"
	"github.com/gorilla/mux"
)

var (
	website = pushpackage.Website{
		Name:            "Buford",
		PushID:          "web.com.github.RobotsAndPencils.buford",
		AllowedDomains:  []string{"https://e31340d3.ngrok.io"},
		URLFormatString: `https://e31340d3.ngrok.io/click?q=%@`,
		// AuthenticationToken identifies the user (16+ characters)
		AuthenticationToken: "19f8d7a6e9fb8a7f6d9330dabe",
		WebServiceURL:       "https://e31340d3.ngrok.io",
	}

	// Cert for signing push packages.
	cert tls.Certificate

	// Service and device token to send push notifications.
	service     *push.Service
	deviceToken string

	templates = template.Must(template.ParseFiles("index.html", "request.html"))
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func requestPermissionHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "request.html", website)
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	p := payload.Browser{
		Alert: payload.BrowserAlert{
			Title: "Hello",
			Body:  "Hello HTTP/2",
		},
		// URLArgs must match placeholders in URLFormatString
		URLArgs: []string{"hello"},
	}
	b, err := json.Marshal(p)
	if err != nil {
		log.Fatal(err)
	}

	id, err := service.Push(deviceToken, nil, b)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("apns-id:", id)
}

func clickHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("clicked", r.URL.Query()["q"])
}

func pushPackagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println("building push package for", vars["websitePushID"])

	w.Header().Set("Content-Type", "application/zip")

	// create a push package and sign it with Cert/Key.
	pkg := pushpackage.New(w)
	pkg.EncodeJSON("website.json", website)
	pkg.File("icon.iconset/icon_128x128@2x.png", "../../testdata/gopher.png")
	pkg.File("icon.iconset/icon_128x128.png", "../../testdata/gopher.png")
	pkg.File("icon.iconset/icon_32x32@2x.png", "../../testdata/gopher.png")
	pkg.File("icon.iconset/icon_32x32.png", "../../testdata/gopher.png")
	pkg.File("icon.iconset/icon_16x16@2x.png", "../../testdata/gopher.png")
	pkg.File("icon.iconset/icon_16x16.png", "../../testdata/gopher.png")
	if err := pkg.Sign(cert, nil); err != nil {
		log.Fatal(err)
	}
}

func registerDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Printf("register device %s (user %s) for %s", vars["deviceToken"], getAuthenticationToken(r), vars["websitePushID"])

	deviceToken = vars["deviceToken"]
}

func forgetDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Printf("forget device %s (user %s) for %s", vars["deviceToken"], getAuthenticationToken(r), vars["websitePushID"])

	deviceToken = ""
}

func getAuthenticationToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	list := strings.SplitN(h, " ", 2)
	if len(list) != 2 || list[0] != "ApplePushNotifications" {
		return ""
	}
	return list[1]
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	var logs struct {
		Logs []string `json:"logs"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logs); err == io.EOF {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	for _, msg := range logs.Logs {
		log.Println(msg)
	}
}

func main() {
	var filename, password string

	flag.StringVar(&filename, "c", "", "Path to p12 certificate file")
	flag.StringVar(&password, "p", "", "Password for p12 file.")
	flag.Parse()

	var err error
	cert, err = certificate.Load(filename, password)
	if err != nil {
		log.Fatal(err)
	}

	client, err := push.NewClient(cert)
	if err != nil {
		log.Fatal(err)
	}

	service = push.NewService(client, push.Production)

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/request", requestPermissionHandler)
	r.HandleFunc("/push", pushHandler)
	r.HandleFunc("/click", clickHandler).Methods("GET")

	// WebServiceURL endpoints
	r.HandleFunc("/v1/pushPackages/{websitePushID}", pushPackagesHandler).Methods("POST")
	r.HandleFunc("/v1/devices/{deviceToken}/registrations/{websitePushID}", registerDeviceHandler).Methods("POST")
	r.HandleFunc("/v1/devices/{deviceToken}/registrations/{websitePushID}", forgetDeviceHandler).Methods("DELETE")
	r.HandleFunc("/v1/log", logHandler).Methods("POST")

	http.ListenAndServe(":5000", r)
}
