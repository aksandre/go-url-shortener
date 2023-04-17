package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
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
		err = errors.New("Передано пустое значение адреса сервера в аргументах командной строки, пример: localhost:8090")
		return
	}

	patrsString := strings.Split(strValue, ":")

	if len(patrsString) > 1 {
		as.host = string(patrsString[0])
		intPort, err := strconv.ParseInt(patrsString[1], 0, 0)
		if err != nil {
			err = fmt.Errorf(`Не удалось получить порт сервера (%s): %w`, strValue, err)
		}
		as.port = int(intPort)
	} else {
		err = errors.New("Некорректное значение адреса сервера в аргументах командной строки, пример: localhost:8090")
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
		err = errors.New("Передан пустой базовый адрес для формирования короткой ссылки, пример: http://localhost:8000")
		return
	}

	patrsString := strings.Split(strValue, "://")
	if len(patrsString) > 1 {
		hsl.protocol = string(patrsString[0])

		patrs2String := strings.Split(patrsString[1], ":")
		if len(patrs2String) > 1 {
			hsl.host = string(patrs2String[0])
			intPort, err := strconv.ParseInt(patrs2String[1], 0, 0)
			if err != nil {
				err = fmt.Errorf(`Не удалось получить порт сервера (%s): %w`, strValue, err)
			}
			hsl.port = int(intPort)
		} else {
			err = errors.New("Некорректное значение адреса сервера в аргументах командной строки")
		}
	} else {
		err = errors.New("Некорректный базовый адрес для формирования короткой ссылки, пример: http://localhost:8000")
	}
	return
}

func (hsl *hostShortLink) Type() string {
	return "hostShortLink"
}

// глобальные переменные флагов
var AddressServerFlag = &addressServer{
	host: "localhost",
	port: 8080,
}
var HostShortLinkFlag = &hostShortLink{
	protocol: "http",
	host:     "localhost",
	port:     8080,
}

func init() {
	flag.Var(AddressServerFlag, "a", "Адрес сервера")
	flag.Var(HostShortLinkFlag, "b", "Базовый адрес для формирования короткой ссылки")
}
