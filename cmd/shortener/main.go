package main

import (
	"fmt"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/storage/storageshortlink"
	"log"

	"net/http"
)

func main() {

	defer func() {
		dbHandler := dbconn.GetDBHandler()
		dbHandler.Close()
	}()

	// Получаем конфиг
	configApp := config.GetAppConfig()
	logger.GetLogger().Debugf("Настройки конфигурации:  %+v", configApp)

	// инициализируем хранилище ссылок
	var storageShortLink = storageshortlink.NewStorageShorts()
	var serviceShortLink = service.NewServiceShortLink(storageShortLink, configApp)

	// Адрес сервера из конфига
	addrServer := configApp.GetAddrServer()
	logger.GetLogger().Debugf("Поднимаем сервер по адресу:  %s", addrServer)

	// получаем обработчик запросов
	handler := handlers.NewRouterHandler(serviceShortLink)
	logger.GetLogger().Debugf("%s", "Запускаем сервер")
	err := http.ListenAndServe(addrServer, handler)
	if err != nil {
		err = fmt.Errorf("ошибка создания сервера: %w", err)
		strError := err.Error()
		logger.GetLogger().Fatalf("%s", strError)
		log.Fatal(strError)
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
var body = '[{"correlation_id":"123456","original_url":"https://123456.com"},{"correlation_id":"456789","original_url":"https://456789.com"}]'
xhr.open("POST", '/api/shorten/batch', true);
xhr.setRequestHeader('Content-Type', 'application/json');
//xhr.onreadystatechange = ...;
xhr.send(body);

var xhr = new XMLHttpRequest();
xhr.open("GET", '/MIy3I6N4', true);
*/

// SET SERVER_ADDRESS=localhost:8080
// SET BASE_URL=http://localhost:8080
// SET FILE_STORAGE_PATH=C:\Users\LENOVO\testLog.log

// SET DATABASE_DSN=postgres://postgres:123456789@localhost:5432/test_psg?sslmode=disable

// SET LEVEL_LOGS_GOLANG=6

// статический анализатор
// для винды выполянть в каждой отдельной папке
// go vet -vettool=C:\GoProjects\golang\project\go-url-shortener\statictest.exe

// генерация модели
// protoc --proto_path=proto --go_opt=paths=source_relative --go_out=. proto/requests_responses.proto

// go build -o shortener.exe

// go run cmd/shortener/main.go --d=postgres://postgres:123456789@localhost:5432/test_psg

// go run cmd/shortener/main.go --a="localhost:8010" --b="https://serviceshort.ru:8020"
// go run cmd/shortener/main.go --a="localhost:8080" --b="http://localhost:8080" --f="C:\Users\LENOVO\goLogs\testlogShortener.log"

// shortenertest -test.v -test.run=^TestIteration1$ -binary-path=C:\GoProjects\golang\project\go-url-shortener\cmd\shortener\shortener.exe
// shortenertest -test.v -test.run=^TestIteration2$ -source-path=.
// shortenertest -test.v -test.run=^TestIteration4$ -source-path=. -binary-path=C:\GoProjects\golang\project\go-url-shortener\cmd\shortener\shortener.exe
// shortenertest -test.v -test.run=^TestIteration5$ -binary-path=cmd/shortener/shortener -server-host=localhost -server-port=8050 -server-base-url="http://localhost:8050"

// shortenertest -test.v -test.run=^TestIteration6$ -binary-path=cmd/shortener/shortener -server-port=8050 -file-storage-path=C:\Users\LENOVO\goLogs\urlShortener\storageLink.json -source-path=.

// shortenertest -test.v -test.run=^TestIteration7$ -binary-path=cmd/shortener/shortener -server-port=8066 -file-storage-path=C:\Users\LENOVO\goLogs\urlShortener\storageLink.json -source-path=.
// тест будет проваливаться, если установлены переменные окружения BASE_URL и SERVER_ADDRESS, их надо обнулить

// shortenertest -test.v -test.run=^TestIteration8$ -binary-path=cmd/shortener/shortener -source-path=.
// shortenertest -test.v -test.run=^TestIteration9$ -source-path=. -binary-path=cmd/shortener/shortener

// shortenertest -test.v -test.run=^TestIteration10$ -source-path=. -binary-path=cmd/shortener/shortener -database-dsn=postgres://postgres:123456789@localhost:5432/test_psg?sslmode=disable
// shortenertest -test.v -test.run=^TestIteration11$ -binary-path=cmd/shortener/shortener -database-dsn=postgres://postgres:123456789@localhost:5432/test_psg?sslmode=disable
// shortenertest -test.v -test.run=^TestIteration12$ -binary-path=cmd/shortener/shortener -database-dsn=postgres://postgres:123456789@localhost:5432/test_psg?sslmode=disable
// shortenertest -test.v -test.run=^TestIteration13$ -binary-path=cmd/shortener/shortener -database-dsn=postgres://postgres:123456789@localhost:5432/test_psg?sslmode=disable
