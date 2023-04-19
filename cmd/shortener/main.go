package main

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/storage/storageshortlink"
	"net/http"
)

func main() {

	// Получаем конфиг
	configApp := config.GetAppConfig()
	logger.GetLogger().Printf("Настройки конфигурации:  %+v", configApp)

	// инициализируем хранилище ссылок
	var storageShortLink = storageshortlink.NewStorageShorts()
	var serviceShortLink = service.NewServiceShortLink(storageShortLink, configApp)

	// Адрес сервера из конфига
	addrServer := configApp.GetAddrServer()
	logger.GetLogger().Printf("Поднимаем сервер по адресу:  %s", addrServer)

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
// shortenertest -test.v -test.run=^TestIteration4$ -source-path=. -binary-path=C:\GoProjects\golang\project\go-url-shortener\cmd\shortener\shortener.exe
// shortenertest -test.v -test.run=^TestIteration5$ -binary-path=cmd/shortener/shortener -server-host=localhost -server-port=8050 -server-base-url="http://localhost:8050"
