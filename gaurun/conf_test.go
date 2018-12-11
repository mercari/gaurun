package gaurun

import (
	"runtime"
	"testing"

	_ "github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	suite.ConfGaurunDefault = BuildDefaultConf()
	suite.ConfGaurun = BuildDefaultConf()
	var err error
	suite.ConfGaurun, err = LoadConf(suite.ConfGaurun, ConfGaurunPath)
	if err != nil {
		panic("failed to load " + ConfGaurunPath)
	}
}

func (suite *ConfigTestSuite) TestValidateConfDefault() {
	// Core
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.Port, "1056")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.WorkerNum, int64(runtime.NumCPU()))
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.QueueNum, int64(8192))
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.NotificationMax, int64(100))
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.PusherMax, int64(0))
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Core.Pid, "")
	// Android
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.ApiKey, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.KeepAliveTimeout, 90)
	assert.Equal(suite.T(), int64(suite.ConfGaurunDefault.Android.KeepAliveConns), suite.ConfGaurunDefault.Core.WorkerNum)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.RetryMax, 1)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Android.UseFCM, true)
	// Ios
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.PemCertPath, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.PemKeyPath, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Sandbox, true)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.RetryMax, 1)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.KeepAliveTimeout, 90)
	assert.Equal(suite.T(), int64(suite.ConfGaurunDefault.Ios.KeepAliveConns), suite.ConfGaurunDefault.Core.WorkerNum)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Ios.Topic, "")
	// FCMv1
	assert.Equal(suite.T(), suite.ConfGaurunDefault.FCMV1.CredentialsFile, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.FCMV1.Project, "")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.FCMV1.Enabled, false)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.FCMV1.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.FCMV1.KeepAliveTimeout, 90)
	assert.Equal(suite.T(), int64(suite.ConfGaurunDefault.FCMV1.KeepAliveConns), suite.ConfGaurunDefault.Core.WorkerNum)
	assert.Equal(suite.T(), suite.ConfGaurunDefault.FCMV1.RetryMax, 1)
	// Log
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Log.AccessLog, "stdout")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Log.ErrorLog, "stderr")
	assert.Equal(suite.T(), suite.ConfGaurunDefault.Log.Level, "error")
}

func (suite *ConfigTestSuite) TestValidateConf() {
	// Core
	assert.Equal(suite.T(), suite.ConfGaurun.Core.Port, "1056")
	assert.Equal(suite.T(), suite.ConfGaurun.Core.WorkerNum, int64(8))
	assert.Equal(suite.T(), suite.ConfGaurun.Core.QueueNum, int64(8192))
	assert.Equal(suite.T(), suite.ConfGaurun.Core.NotificationMax, int64(100))
	assert.Equal(suite.T(), suite.ConfGaurun.Core.PusherMax, int64(0))
	assert.Equal(suite.T(), suite.ConfGaurun.Core.Pid, "")
	// Android
	assert.Equal(suite.T(), suite.ConfGaurun.Android.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.ApiKey, "apikey for GCM")
	assert.Equal(suite.T(), suite.ConfGaurun.Android.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.KeepAliveTimeout, 30)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.KeepAliveConns, 4)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.RetryMax, 1)
	assert.Equal(suite.T(), suite.ConfGaurun.Android.UseFCM, false)
	// Ios
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.PemCertPath, "cert.pem")
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.PemKeyPath, "key.pem")
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.Sandbox, true)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.RetryMax, 1)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.KeepAliveTimeout, 30)
	// FCMv1
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.Enabled, true)
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.CredentialsFile, "adminsdk.json")
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.Project, "project for fcm")
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.Timeout, 5)
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.KeepAliveTimeout, 30)
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.KeepAliveConns, 4)
	assert.Equal(suite.T(), suite.ConfGaurun.FCMV1.RetryMax, 1)
	// Log
	assert.Equal(suite.T(), suite.ConfGaurun.Ios.KeepAliveConns, 6)
	assert.Equal(suite.T(), suite.ConfGaurun.Log.AccessLog, "stdout")
	assert.Equal(suite.T(), suite.ConfGaurun.Log.ErrorLog, "stderr")
	assert.Equal(suite.T(), suite.ConfGaurun.Log.Level, "error")
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
