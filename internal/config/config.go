package config

type ConfigTypeInterface interface {
	GetAddrServer() string
	GetHostShortLink() string

	SetAddrServer(string)
	SetHostShortLink(string)

	GetFileStoragePath() string
	SetFileStoragePath(string)

	// для логирования
	GetLogsPath() string
	SetLogsPath(string)
	GetLevelLogs() int
	SetLevelLogs(int)
	GetUserHomePath() string
}

// Тип для хранения конфигурации приложения
// Экспортируемая - для мокирования
type ConfigType struct {
	addrServer      string
	hostShortLink   string
	fileStoragePath string

	logsPath     string
	levelLogs    int
	userHomePath string
}

func (ct *ConfigType) SetAddrServer(value string) {
	ct.addrServer = value
}

func (ct *ConfigType) GetAddrServer() string {
	return ct.addrServer
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

func (ct *ConfigType) SetLevelLogs(value int) {
	ct.levelLogs = value
}

func (ct *ConfigType) GetLevelLogs() int {
	return ct.levelLogs
}

func (ct *ConfigType) SetFileStoragePath(value string) {
	ct.fileStoragePath = value
}

func (ct *ConfigType) GetFileStoragePath() string {
	return ct.fileStoragePath
}

func (ct *ConfigType) GetUserHomePath() string {
	return ct.userHomePath
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

	ct.fileStoragePath = flags.FileStoragePath
	if envVars.FileStoragePath != "" {
		ct.fileStoragePath = envVars.FileStoragePath
	}

	ct.levelLogs = flags.LevelLogs
	if envVars.LevelLogs != -1 {
		ct.levelLogs = envVars.LevelLogs
	}

	ct.userHomePath = envVars.UserHomePath
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
