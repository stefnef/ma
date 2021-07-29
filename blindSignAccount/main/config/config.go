package config

import (
	"github.com/gin-gonic/gin"
	"github.com/tkanos/gonfig"
)

type configuration struct {
	Name      string
	Port      string
	Host      string
	GinMode   string
	ResultDir string
}

var config configuration

func ReadConfigFile(fileName string) error {
	config = configuration{}
	err := gonfig.GetConf(fileName, &config)
	if err != nil {
		return err
	}

	return nil
}

func GetConfigName() string {
	return config.Name
}

func SetConfigPort(port string) {
	config.Port = port
}

func GetConfigAddress() string {
	return config.Host + ":" + config.Port
}

func GetConfigGinMode() string {
	switch config.GinMode {
	case gin.ReleaseMode:
		return gin.ReleaseMode
	}
	return gin.DebugMode
}

func GetConfigResultDir() string {
	return config.ResultDir
}
