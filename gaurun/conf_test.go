package gaurun

import (
	_ "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"runtime"
	"testing"
)

const (
	ConfGaurunPath = "../conf/gaurun.toml"
)

type ConfigTestSuite struct {
	suite.Suite
	ConfGaurunDefault ConfToml
	ConfGaurun        ConfToml
}

func (suite *ConfigTestSuite) SetupTest() {
	suite.ConfGaurunDefault = BuildDefaultConfGaurun()
	var err error
	suite.ConfGaurun, err = LoadConfGaurun(suite.ConfGaurun, ConfGaurunPath)
	if err != nil {
		panic("failed to load " + ConfGaurunPath)
	}
}

func (suite *ConfigTestSuite) TestValidateConfDefault() {
	// Core
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.Port, "1056")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.WorkerNum, runtime.NumCPU())
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.QueueNum, 8192)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.NotificationMax, 100)
	// API
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Api.PushUri, "/push")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Api.StatGoUri, "/stat/go")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Api.StatAppUri, "/stat/app")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Api.ConfigAppUri, "/config/app")
	// Android
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.ApiKey, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.RetryMax, 1)
	// Ios
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.PemCertPath, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.PemKeyPath, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Sandbox, true)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Timeout, 0)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.RetryMax, 1)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.TimeoutError, 500)
	// Lo
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Log.AccessLog, "stdout")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Log.ErrorLog, "stderr")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Log.Level, "error")
}

func (suite *ConfigTestSuite) TestValidateConf() {
	// Core
	assert.Equal(suite.T(), suite.ConfGaurun.Core.Port, "1056")
	assert.Equal(suite.T(), suite.ConfGaurun.Core.WorkerNum, 8)
	assert.Equal(suite.T(), suite.ConfGaurun.Core.QueueNum, 512)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.NotificationMax, 100)
	// API
	assert.Equal(suite.T(), suite.ConfGaurun.Api.PushUri, "/push")
	assert.Equal(suite.T(), suite.ConfGaurun.Api.StatGoUri, "/stat/go")
	assert.Equal(suite.T(), suite.ConfGaurun.Api.StatAppUri, "/stat/app")
	assert.Equal(suite.T(), suite.ConfGaurun.Api.ConfigAppUri, "/config/app")
	// Android
	assert.Equal(suite.T(), suite.ConfGaurun.Android.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.ApiKey, "apikey for GCM")
	assert.Equal(suite.T(), suite.ConfGaurun.Android.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.RetryMax, 0)
	// Ios
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.PemCertPath, "cert.pem")
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.PemKeyPath, "key.pem")
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.Sandbox, true)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.Timeout, 0)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.RetryMax, 0)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.TimeoutError, 500)
	// Lo
	assert.Equal(suite.T(), suite.ConfGaurun.Log.AccessLog, "stdout")
	assert.Equal(suite.T(), suite.ConfGaurun.Log.ErrorLog, "stderr")
	assert.Equal(suite.T(), suite.ConfGaurun.Log.Level, "error")
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
