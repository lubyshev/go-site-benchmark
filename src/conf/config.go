package conf

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strconv"
)

type AppConfig struct {
	ServerHost string
	ServerPort int

	MaxConnections int
}

var config *AppConfig

func GetConfig() (*AppConfig, error) {
	if config == nil {
		err := loadConfig()
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func loadConfig() (err error) {
	config = new(AppConfig)

	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s/etc/.env", rootPath)

	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("config file does not exist: %s", fileName))
	}

	myEnv, err := godotenv.Read(fileName)
	if err != nil {
		return errors.New(fmt.Sprintf("error while read config file: %s", fileName))
	}

	err = processEnv(myEnv)
	if err != nil {
		return err
	}

	return nil
}

func processEnv(env map[string]string) error {
	port, err := strconv.Atoi(env["APP_SERVER_PORT"])
	if err != nil {
		return err
	}
	config.ServerPort = port

	conns, err := strconv.Atoi(env["APP_MAX_CONNECTIONS"])
	if err != nil {
		return err
	}
	config.MaxConnections = conns

	config.ServerHost = "localhost"
	if host, ok := env["APP_SERVER_HOST"]; ok && host != "" {
		config.ServerHost = host
	}

	return nil
}
