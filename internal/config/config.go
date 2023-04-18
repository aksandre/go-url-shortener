package config

import (
	"go-url-shortener/internal/logger"
)

type ConfigTypeInterface interface {
	install()
	GetAddrServer() string
	GetHostShortLink() string
	SetAddrServer(string)
	SetHostShortLink(string)
}

// Тип для хранения конфигурации приложения
// Экспортируемая - для мокирования
type ConfigType struct {
	addrServer    string
	hostShortLink string
}

func (ct *ConfigType) install() {

	ct.addrServer = AddressServerFlag.String()
	if EnviromentConfig.AddressServer != "" {
		ct.addrServer = EnviromentConfig.AddressServer
	}

	ct.hostShortLink = HostShortLinkFlag.String()
	if EnviromentConfig.HostShortLink != "" {
		ct.hostShortLink = EnviromentConfig.HostShortLink
	}

	logger.AppLogger.Printf("Данные конфигурации:  %+v", ct)
}

func (ct *ConfigType) GetAddrServer() string {
	return ct.addrServer
}

func (ct *ConfigType) SetHostShortLink(value string) {
	ct.hostShortLink = value
}

func (ct *ConfigType) SetAddrServer(value string) {
	ct.addrServer = value
}

func (ct *ConfigType) GetHostShortLink() string {
	return ct.hostShortLink
}

func NewConfigApp() ConfigTypeInterface {
	// переменная конфига
	configApp := &ConfigType{}
	configApp.install()
	return configApp
}
