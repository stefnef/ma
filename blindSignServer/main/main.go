package main

import (
	"blindSignAccount/main/config"
	"blindSignServer/main/handlers"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
)

var router *gin.Engine

func main() {

	var configFile = "main/config/configTEST.json"
	if len(os.Args) == 2 {
		configFile = os.Args[1]
	}
	if err := config.ReadConfigFile(configFile); err != nil {
		var configFile = "config/configGCL.json"
		if errGCloud := config.ReadConfigFile(configFile); errGCloud != nil {
			panic(errGCloud)
		} else {
			// running in GCL
			port := os.Getenv("PORT")
			if port != "" {
				log.Printf("Defaulting to port %s", port)
				config.SetConfigPort(port)
			}

		}
	}

	gin.SetMode(config.GetConfigGinMode())
	log.Println("configuration name: '" + config.GetConfigName() + "'")
	if config.GetConfigGinMode() == gin.ReleaseMode {
		log.Println("running in " + config.GetConfigGinMode() + " mode")
		log.Println("Listening and serving HTTP on " + config.GetConfigAddress())
		gin.DefaultWriter = ioutil.Discard
		errorLogFile, _ := os.Create("error.log")
		gin.DefaultErrorWriter = errorLogFile
	}

	// Initialize routes
	router = gin.Default()
	handlers.InitRoutes(router)

	// Run
	if err := router.Run(config.GetConfigAddress()); err != nil {
		fmt.Print(err)
	}
}
