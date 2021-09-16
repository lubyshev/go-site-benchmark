package conf

import (
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"strconv"
	"time"
)

type AppConfig struct {
	ServerPort              int
	CacheTtl                time.Duration
	OverloadWorkers         int
	OverloadInitConnections int
	OverloadMaxLimit        int
	OverloadMaxConnections  int
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

	myEnv := make(map[string]string)
	fileName := fmt.Sprintf("%s/etc/.env", rootPath)
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		getEnv := func(key string) string {
			val, ok := os.LookupEnv(key)
			if !ok {
				return ""
			} else {
				return val
			}
		}
		myEnv["APP_SERVER_PORT"] = getEnv("APP_SERVER_PORT")
		myEnv["APP_CACHE_TTL"] = getEnv("APP_CACHE_TTL")
		myEnv["APP_OVERLOAD_QUEUE_WORKERS"] = getEnv("APP_OVERLOAD_QUEUE_WORKERS")
		myEnv["APP_OVERLOAD_QUEUE_WORKERS"] = getEnv("APP_OVERLOAD_QUEUE_WORKERS")
		myEnv["APP_OVERLOAD_MAX_LIMIT"] = getEnv("APP_OVERLOAD_MAX_LIMIT")
		myEnv["APP_OVERLOAD_MAX_CONNECTIONS"] = getEnv("APP_OVERLOAD_MAX_CONNECTIONS")
	} else {
		myEnv, err = godotenv.Read(fileName)
		if err != nil {
			log.Fatal(fmt.Sprintf("error while read config file: %s", fileName))
		}
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

	ttl, err := strconv.Atoi(env["APP_CACHE_TTL"])
	if err != nil {
		return err
	}
	config.CacheTtl = time.Duration(ttl * 1_000_000_000)

	workers, err := strconv.Atoi(env["APP_OVERLOAD_QUEUE_WORKERS"])
	if err != nil {
		return err
	}
	config.OverloadWorkers = workers

	cons, err := strconv.Atoi(env["APP_OVERLOAD_QUEUE_WORKERS"])
	if err != nil {
		return err
	}
	config.OverloadInitConnections = cons

	limit, err := strconv.Atoi(env["APP_OVERLOAD_MAX_LIMIT"])
	if err != nil {
		return err
	}
	config.OverloadMaxLimit = limit

	maxCons, err := strconv.Atoi(env["APP_OVERLOAD_MAX_CONNECTIONS"])
	if err != nil {
		return err
	}
	config.OverloadMaxConnections = maxCons

	return nil
}
