package conf

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	OverloadMethodSimple = "simple"
	OverloadMethodStrong = "strong"
)

type AppConfig struct {
	ServerPort              int
	CacheBgFrequency        time.Duration
	CacheDebug              bool
	CacheTtl                time.Duration
	OverloadWorkers         int
	OverloadInitConnections int
	OverloadMaxLimit        int
	OverloadMaxConnections  int
	OverloadMethod          string
}

type TestConfig struct {
	CacheBgFrequency time.Duration
}

var config *AppConfig
var testConfig *TestConfig

func GetConfig() *AppConfig {
	if config == nil {
		loadConfig()
	}
	return config
}

func GetTestConfig() *TestConfig {
	if testConfig == nil {
		loadTestConfig()
	}
	return testConfig
}

func loadConfig() {
	config = new(AppConfig)

	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("can`t get working directory: %s", err.Error()))
	}

	fileExists := false
	myEnv := make(map[string]string)
	fileName := fmt.Sprintf("%s/etc/.env", rootPath)
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		// for tests && benchmarks
		fileName = fmt.Sprintf("%s/../etc/.env", rootPath)
		if _, err = os.Stat(fileName); err == nil {
			fileExists = true
		}
	} else {
		fileExists = true
	}
	if !fileExists {
		log.Printf("%s not found: load config from env", fileName)
		getEnv := func(key string) string {
			val, ok := os.LookupEnv(key)
			if !ok {
				return ""
			} else {
				return val
			}
		}
		myEnv["APP_SERVER_PORT"] = getEnv("APP_SERVER_PORT")
		myEnv["APP_CACHE_BACKGROUND_FREQUENCY"] = getEnv("APP_CACHE_BACKGROUND_FREQUENCY")
		myEnv["APP_CACHE_DEBUG"] = getEnv("APP_CACHE_DEBUG")
		myEnv["APP_CACHE_TTL"] = getEnv("APP_CACHE_TTL")
		myEnv["APP_OVERLOAD_QUEUE_WORKERS"] = getEnv("APP_OVERLOAD_QUEUE_WORKERS")
		myEnv["APP_OVERLOAD_QUEUE_WORKERS"] = getEnv("APP_OVERLOAD_QUEUE_WORKERS")
		myEnv["APP_OVERLOAD_MAX_LIMIT"] = getEnv("APP_OVERLOAD_MAX_LIMIT")
		myEnv["APP_OVERLOAD_MAX_CONNECTIONS"] = getEnv("APP_OVERLOAD_MAX_CONNECTIONS")
		myEnv["APP_OVERLOAD_METHOD"] = getEnv("APP_OVERLOAD_METHOD")
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

func loadTestConfig() {
	testConfig = new(TestConfig)

	rootPath, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("can`t get working directory: %s", err.Error()))
	}

	myEnv := make(map[string]string)
	fileName := fmt.Sprintf("%s/../etc/.test.env", rootPath)
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		log.Printf("%s not found: load config from env", fileName)
		getEnv := func(key string) string {
			val, ok := os.LookupEnv(key)
			if !ok {
				return ""
			} else {
				return val
			}
		}
		myEnv["TEST_CACHE_BACKGROUND_FREQUENCY"] = getEnv("TEST_CACHE_BACKGROUND_FREQUENCY")
	} else {
		myEnv, err = godotenv.Read(fileName)
		if err != nil {
			log.Fatal(fmt.Sprintf("error while read config file: %s", fileName))
		}
	}

	err = processTestEnv(myEnv)
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

	freq, err := strconv.Atoi(env["APP_CACHE_BACKGROUND_FREQUENCY"])
	if err != nil {
		return err
	}
	config.CacheBgFrequency = time.Duration(freq * 1_000_000_000)

	config.CacheDebug = "yes" == env["APP_CACHE_DEBUG"]

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

	switch env["APP_OVERLOAD_METHOD"] {
	case OverloadMethodSimple:
		config.OverloadMethod = OverloadMethodSimple
	case OverloadMethodStrong:
		config.OverloadMethod = OverloadMethodStrong
	default:
		return errors.New("invalid overload method")
	}

	return nil
}

func processTestEnv(env map[string]string) error {
	freq, err := strconv.Atoi(env["TEST_CACHE_BACKGROUND_FREQUENCY"])
	if err != nil {
		return err
	}
	testConfig.CacheBgFrequency = time.Duration(freq * 1_000_000_000)
	return nil
}
