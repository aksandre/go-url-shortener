package config

import (
	env "github.com/caarlos0/env/v8"
)

type EnviromentConfigType struct {
	AddressServer string `env:"SERVER_ADDRESS"`
	HostShortLink string `env:"BASE_URL"`
}

// Глобальные переменные окружения
// Сделаем эспортируемыми, чтобы можно было управлять в тестах
var EnviromentConfig = EnviromentConfigType{}

func init() {
	env.Parse(&EnviromentConfig)
}
