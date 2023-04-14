package main

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/handlers"
	"go-url-shortener/internal/storage/storageShortlink"
	"net/http"
)

func main() {

	// инициализируем хранилище ссылок
	var storageShortLink = storageShortlink.NewStorageShorts()
	var serviceShortLink = service.NewServiceShortLink(storageShortLink, 8)

	handler := handlers.NewRouterHandler(serviceShortLink)
	err := http.ListenAndServe(`:8080`, handler)
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
xhr.open("GET", '/MIy3I6N4', true);
*/

// go build -o shortener.exe
// shortenertest -test.v -test.run=^TestIteration1$ -binary-path=C:\GoProjects\golang\project\go-url-shortener\cmd\shortener\shortener.exe
