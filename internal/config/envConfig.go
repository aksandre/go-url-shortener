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

	LogsPath     string `env:"LOGS_PATH_GOLANG"`
	LevelLogs    int    `env:"LEVEL_LOGS_GOLANG"`
	UserHomePath string `env:"USER_HOME_PATH"`
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

	// если значение не установлено, то устанавливаем сами
	_, okLevelLogs := os.LookupEnv("LEVEL_LOGS_GOLANG")
	if !okLevelLogs {
		enviromentConfig.LevelLogs = -1
	}

	// путь до домашней директории пользователя по-умолчанию
	defaultHomePath := getDefaultUserHomePath()

	if enviromentConfig.UserHomePath == "" {
		enviromentConfig.UserHomePath = defaultHomePath
	}

	if enviromentConfig.LogsPath == "" {
		enviromentConfig.LogsPath = defaultHomePath
	}
}

// Получаем путь до домашней директории пользователя
func getDefaultUserHomePath() (homePath string) {
	valHome, okHome := os.LookupEnv("HOME")
	valHomepath, okHomepath := os.LookupEnv("HOMEPATH")
	if okHome {
		homePath = valHome
	} else if okHomepath {
		homePath = valHomepath
	}
	return
}
