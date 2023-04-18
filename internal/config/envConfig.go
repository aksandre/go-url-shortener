package config

import (
	"os"

	env "github.com/caarlos0/env/v8"
)

// Тип для хранения переменных окружения
type EnviromentConfigType struct {
	AddressServer string `env:"SERVER_ADDRESS"`
	HostShortLink string `env:"BASE_URL"`
	LogsPath      string `env:"LOGS_PATH_GOLANG"`
}

// Глобальные переменные окружения
// Сделаем эспортируемыми, чтобы можно было управлять в тестах
var enviromentConfig = EnviromentConfigType{}

// Маркер синглтона, что сущность, уже инициировали
var setupEnviroment = false

func GetEnviromentConfig() EnviromentConfigType {
	if setupEnviroment == false {
		initEnviroment()
		setupEnviroment = true
	}
	return enviromentConfig
}

func SetEnviromentConfig(config EnviromentConfigType) {
	setupEnviroment = true
	enviromentConfig = config
}

// инициализация сущности
func initEnviroment() {

	env.Parse(&enviromentConfig)

	if enviromentConfig.LogsPath == "" {
		valHome, okHome := os.LookupEnv("HOME")
		valHomepath, okHomepath := os.LookupEnv("HOMEPATH")
		if okHome {
			enviromentConfig.LogsPath = valHome
		} else if okHomepath {
			enviromentConfig.LogsPath = valHomepath
		}
	}
}
