package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	//flag "github.com/spf13/pflag"
)

// Адрес запуска HTTP-сервера
// Значение может быть таким: localhost:8888
type addressServer struct {
	host string
	port int
}

func (as *addressServer) String() string {
	port := as.port
	strPort := strconv.Itoa(port)
	return as.host + ":" + strPort
}

func (as *addressServer) Set(strValue string) (err error) {

	if len(strValue) == 0 {
		err = errors.New("передано пустое значение адреса сервера в аргументах командной строки, пример: localhost:8090")
		return
	}

	patrsString := strings.Split(strValue, ":")

	if len(patrsString) > 1 {
		as.host = patrsString[0]
		intPort, errParse := strconv.ParseInt(patrsString[1], 0, 0)
		if errParse != nil {
			err = fmt.Errorf("не удалось получить порт сервера (%s): %w", strValue, errParse)
		} else {
			as.port = int(intPort)
		}
	} else {
		err = errors.New("некорректное значение адреса сервера в аргументах командной строки, пример: localhost:8090")
	}

	return
}

func (as *addressServer) Type() string {
	return "addressServer"
}

// Базовый адрес результирующего сокращённого URL
// Значение: адрес сервера перед коротким URL, например http://localhost:8000/qsd54gFg
type hostShortLink struct {
	protocol string
	host     string
	port     int
}

func (hsl *hostShortLink) String() string {
	protocol := hsl.protocol
	port := hsl.port
	strPort := strconv.Itoa(port)
	return protocol + "://" + hsl.host + ":" + strPort
}

func (hsl *hostShortLink) Set(strValue string) (err error) {

	if len(strValue) == 0 {
		err = errors.New("передан пустой базовый адрес для формирования короткой ссылки, пример: http://localhost:8000")
		return
	}

	patrsString := strings.Split(strValue, "://")
	if len(patrsString) > 1 {
		hsl.protocol = patrsString[0]

		patrs2String := strings.Split(patrsString[1], ":")
		if len(patrs2String) > 1 {
			hsl.host = string(patrs2String[0])
			intPort, errParse := strconv.ParseInt(patrs2String[1], 0, 0)
			if errParse != nil {
				err = fmt.Errorf(`не удалось получить порт сервера (%s): %w`, strValue, errParse)
			} else {
				hsl.port = int(intPort)
			}
		} else {
			err = errors.New("некорректное значение адреса сервера в аргументах командной строки")
		}
	} else {
		err = errors.New("некорректный базовый адрес для формирования короткой ссылки, пример: http://localhost:8000")
	}
	return
}

func (hsl *hostShortLink) Type() string {
	return "hostShortLink"
}

// Тип для хранения флагов запуска приложения
type FlagConfigType struct {
	AddressServer   *addressServer
	HostShortLink   *hostShortLink
	FileStoragePath string
	LevelLogs       int
}

// Глобальные переменные окружения
// Сделаем эспортируемыми, чтобы можно было управлять в тестах
var flagConfig = FlagConfigType{
	AddressServer: &addressServer{
		host: "localhost",
		port: 8080,
	},
	HostShortLink: &hostShortLink{
		protocol: "http",
		host:     "localhost",
		port:     8080,
	},
	FileStoragePath: "",
	LevelLogs:       int(log.InfoLevel),
}

// Маркер синглтона, что сущность, уже инициировали
var setupFlags = false

func GetFlagConfig() FlagConfigType {
	if !setupFlags {
		initFlags()
		setupFlags = true
	}
	return flagConfig
}

func SetFlagConfig(config FlagConfigType) {
	setupFlags = true
	flagConfig = config
}

// инициализация сущности
func initFlags() {
	flag.Var(flagConfig.AddressServer, "a", "Адрес сервера")
	flag.Var(flagConfig.HostShortLink, "b", "Базовый адрес для формирования короткой ссылки")
	flag.StringVar(&flagConfig.FileStoragePath, "f", "", "Путь до файла лога")
	flag.IntVar(&flagConfig.LevelLogs, "logLevel", int(log.InfoLevel), "Уровень логирования")
	flag.Parse()
}
