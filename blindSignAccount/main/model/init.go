package model

import "blindSignAccount/main/config"

var configFile = "../config/configTEST.json"

func init() {
	if err := config.ReadConfigFile(configFile); err == nil {
		ServerAddress = "http://" + config.GetConfigAddress()
	}
}
