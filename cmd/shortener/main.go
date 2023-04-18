package main

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/storage/storageShortlink"
	"net/http"

	flag "github.com/spf13/pflag"
)

func installConfig() config.ConfigTypeInterface {
	configApp := config.NewConfigApp()
	flag.Parse()
	return configApp
}

func main() {

	// Создаем конфиг
	configApp := installConfig()

	// инициализируем хранилище ссылок
	var storageShortLink = storageShortlink.NewStorageShorts()
	var serviceShortLink = service.NewServiceShortLink(storageShortLink, configApp)

	// Адрес сервера из конфига
	addrServer := configApp.GetAddrServer()
	logger.AppLogger.Printf("Поднимаем сервер по адресу:  %s", addrServer)

	handler := handlers.NewRouterHandler(serviceShortLink)
	err := http.ListenAndServe(addrServer, handler)
	if err != nil {
		panic(err)
	}
}

/*
var xhr = new XMLHttpRequest();
var body = 'https://practicum.yandex.ru/';
xhr.open("POST", '/', true);
xhr.setRequestHeader('Content-Type', 'text/plain');
//xhr.onreadystatechange = ...;
xhr.send(body);

var xhr = new XMLHttpRequest();
var body = '{"url":"https://pract.yandex.ru/"}'
xhr.open("POST", '/api/shorten', true);
xhr.setRequestHeader('Content-Type', 'application/json');
//xhr.onreadystatechange = ...;
xhr.send(body);

var xhr = new XMLHttpRequest();
xhr.open("GET", '/MIy3I6N4', true);
*/

// SET SERVER_ADDRESS=localhost:8045
// SET BASE_URL=https://localhost:8041/hhhh/

// go build -o shortener.exe
// go run cmd/shortener/main.go --a="localhost:8010" --b="https://serviceshort.ru:8020"
// go run cmd/shortener/main.go --a="localhost:8080" --b="http://localhost:8080"
// shortenertest -test.v -test.run=^TestIteration1$ -binary-path=C:\GoProjects\golang\project\go-url-shortener\cmd\shortener\shortener.exe
