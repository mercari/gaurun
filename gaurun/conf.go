package gaurun

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"net/http"
	"runtime"
)

type ConfToml struct {
	Core    SectionCore    `toml:"core"`
	Api     SectionApi     `toml:"api"`
	Android SectionAndroid `toml:"android"`
	Ios     SectionIos     `toml:"ios"`
	Log     SectionLog     `toml:"log"`
}

type SectionCore struct {
	Port            string `toml:"port"`
	WorkerNum       int    `toml:"workers"`
	QueueNum        int    `toml:"queues"`
	NotificationMax int    `toml:"notification_max"`
}

type SectionApi struct {
	PushUri      string `toml:"push_uri"`
	StatGoUri    string `toml:"stat_go_uri"`
	StatAppUri   string `toml:"stat_app_uri"`
	ConfigAppUri string `toml:"config_app_uri"`
}

type SectionAndroid struct {
	Enabled  bool   `toml:"enabled"`
	ApiKey   string `toml:"apikey"`
	Timeout  int    `toml:"timeout"`
	RetryMax int    `toml:"retry_max"`
}

type SectionIos struct {
	Enabled      bool   `toml:"enabled"`
	PemCertPath  string `toml:"pem_cert_path"`
	PemKeyPath   string `toml:"pem_key_path"`
	Sandbox      bool   `toml:"sandbox"`
	Timeout      int    `toml:"timeout"`
	RetryMax     int    `toml:"retry_max"`
	TimeoutError int    `toml:"timeout_error"`
	KeepAliveMax int    `toml:"keepalive_max"`
}

type SectionLog struct {
	AccessLog string `toml:"access_log"`
	ErrorLog  string `toml:"error_log"`
	Level     string `toml:"level"`
}

func BuildDefaultConfGaurun() ConfToml {
	var conf ConfToml
	// Core
	conf.Core.Port = "1056"
	conf.Core.WorkerNum = runtime.NumCPU()
	conf.Core.QueueNum = 512
	conf.Core.NotificationMax = 100
	// Api
	conf.Api.PushUri = "/push"
	conf.Api.StatGoUri = "/stat/go"
	conf.Api.StatAppUri = "/stat/app"
	conf.Api.ConfigAppUri = "/config/app"
	// Android
	conf.Android.ApiKey = ""
	conf.Android.Enabled = true
	conf.Android.Timeout = 5
	conf.Android.RetryMax = 0
	// iOS
	conf.Ios.Enabled = true
	conf.Ios.PemCertPath = ""
	conf.Ios.PemKeyPath = ""
	conf.Ios.Sandbox = true
	conf.Ios.Timeout = 0
	conf.Ios.RetryMax = 0
	conf.Ios.TimeoutError = 1000
	conf.Ios.KeepAliveMax = 0
	// log
	conf.Log.AccessLog = "stdout"
	conf.Log.ErrorLog = "stderr"
	conf.Log.Level = "error"
	return conf
}

func LoadConfGaurun(confGaurun ConfToml, confPath string) (ConfToml, error) {
	_, err := toml.DecodeFile(confPath, &confGaurun)
	if err != nil {
		return confGaurun, err
	}
	return confGaurun, nil
}

func ConfigGaurunHandler(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	e := toml.NewEncoder(&b)
	result := ConfGaurun
	// hide Apikey
	result.Android.ApiKey = "..."
	err := e.Encode(result)
	if err != nil {
		msg := "Response-body could not be created"
		LogError.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Server", fmt.Sprintf("Gaurun %s", Version))
	fmt.Fprintf(w, b.String())
}
