package config

import (
	"os"

	env "github.com/caarlos0/env/v8"
)

// Тип для хранения переменных окружения
type EnviromentConfigType struct {
	AddressServer   string `env:"SERVER_ADDRESS"`
	HostShortLink   string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`

	LogsPath  string `env:"LOGS_PATH_GOLANG"`
	LevelLogs int    `env:"LEVEL_LOGS_GOLANG"`
}

// Глобальные переменные окружения
// Сделаем эспортируемыми, чтобы можно было управлять в тестах
var enviromentConfig = EnviromentConfigType{}

// Маркер синглтона, что сущность, уже инициировали
var setupEnviroment = false

func GetEnviromentConfig() EnviromentConfigType {
	if !setupEnviroment {
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

	_, okLevelLogs := os.LookupEnv("LEVEL_LOGS_GOLANG")
	if !okLevelLogs {
		enviromentConfig.LevelLogs = -1
	}
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
