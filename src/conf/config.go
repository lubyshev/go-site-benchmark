package conf

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"strconv"
)

type AppConfig struct {
	ServerPort int
}

var config *AppConfig

func GetConfig() *AppConfig {
	if config == nil {
		loadConfig()
	}
	return config
}

func loadConfig() {
	config = new(AppConfig)

	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("can`t get working directory: %s", err.Error()))
	}
	fileName := fmt.Sprintf("%s/etc/.env", rootPath)

	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		log.Fatal(fmt.Sprintf("config file does not exist: %s", fileName))
	}

	myEnv, err := godotenv.Read(fileName)
	if err != nil {
		log.Fatal(fmt.Sprintf("error while read config file: %s", fileName))
	}

	err = processEnv(myEnv)
	if err != nil {
		log.Fatal(fmt.Sprintf("error while processing config file: %s", err.Error()))
	}
}

func processEnv(env map[string]string) error {
	port, err := strconv.Atoi(env["APP_SERVER_PORT"])
	if err != nil {
		return err
	}
	config.ServerPort = port

	return nil
}
