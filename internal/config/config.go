package config

type ConfigTypeInterface interface {
	GetAddrServer() string
	GetHostShortLink() string

	SetAddrServer(string)
	SetHostShortLink(string)
	// для логирования
	GetLogsPath() string
	SetLogsPath(string)
}

// Тип для хранения конфигурации приложения
// Экспортируемая - для мокирования
type ConfigType struct {
	addrServer    string
	hostShortLink string
	logsPath      string
}

func (ct *ConfigType) GetAddrServer() string {
	return ct.addrServer
}

func (ct *ConfigType) SetAddrServer(value string) {
	ct.addrServer = value
}

func (ct *ConfigType) SetHostShortLink(value string) {
	ct.hostShortLink = value
}

func (ct *ConfigType) GetHostShortLink() string {
	return ct.hostShortLink
}

func (ct *ConfigType) SetLogsPath(value string) {
	ct.logsPath = value
}

func (ct *ConfigType) GetLogsPath() string {
	return ct.logsPath
}

func (ct *ConfigType) installConfig() {

	envVars := GetEnviromentConfig()
	flags := GetFlagConfig()

	ct.addrServer = flags.AddressServer.String()
	if envVars.AddressServer != "" {
		ct.addrServer = envVars.AddressServer
	}

	ct.hostShortLink = flags.HostShortLink.String()
	if envVars.HostShortLink != "" {
		ct.hostShortLink = envVars.HostShortLink
	}

	ct.logsPath = envVars.LogsPath
}

var appConfig = &ConfigType{}

// Маркер синглтона, что сущность, уже инициировали
var setupAppConfig = false

func GetAppConfig() ConfigTypeInterface {
	if !setupAppConfig {
		// переменная конфига
		appConfig.installConfig()
		setupAppConfig = true
	}
	return appConfig
}

func SetAppConfig(value *ConfigType) {
	setupAppConfig = true
	appConfig = value
}
