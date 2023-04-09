package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

func getUrlRequest(req *http.Request) string {
	url := req.Host + req.URL.Path
	return url
}

func getModuleError(textError string) error {
	textModuleError := "go-url-shortener: " + textError
	return errors.New(textModuleError)
}

func getProtocolHttp() string {
	return "http"
}

func getUrlService(req *http.Request) string {
	return getProtocolHttp() + "://" + req.Host
}

func NewDataShortLink() DataShortLink {
	return DataShortLink{
		Data:            make(map[string]string),
		lengthShortLink: 8,
	}
}

type DataShortLink struct {
	Data            map[string]string
	lengthShortLink int
}

func (store DataShortLink) SetLength(length int) {
	store.lengthShortLink = length
}

func (store DataShortLink) getRandString(length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	result := b.String()
	return result
}

func (store DataShortLink) getShortLinkByUrl(fullUrl string) (shortLink string, err error) {

	for short, urlForShort := range store.Data {
		if fullUrl == urlForShort {
			shortLink = short
			break
		}
	}

	if shortLink != "" {
		return
	} else {
		//такой url еще не приходил, генерируем новую ссылку
		lengthShort := store.lengthShortLink
		shortLink = store.getRandString(lengthShort)
		store.Data[shortLink] = fullUrl

		// err = getModuleError("Не сформировать короткую ссылку")
	}
	return
}
func (store DataShortLink) getFullLinkByShort(shortLink string) (fullUrl string, err error) {

	fullUrl, ok := store.Data[shortLink]
	if !ok {
		// должны показать ошибку
		err = getModuleError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}

func mainPage(res http.ResponseWriter, req *http.Request) {

	if req.URL.Path == "/" {

		// генерация короткой ссылки
		if req.Method == http.MethodPost {
			// получаем тело из запроса и проводим его к строке
			dataBody := req.Body
			lenBody := req.ContentLength
			result := make([]byte, lenBody)
			dataBody.Read(result)
			urlFull := string(result)
			urlFull = strings.TrimSpace(urlFull)
			shortLink, err := dataShortLink.getShortLinkByUrl(urlFull)
			fmt.Println(urlFull)
			fmt.Printf("%+v", dataShortLink)
			if err != nil {
				strError := err.Error()
				fmt.Println("Ошибка: " + strError)

				res.Header().Set("Content-Type", "text/plain; charset=8")
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte(strError))

			} else {
				strResult := getUrlService(req) + "/" + shortLink

				lenResult := len(strResult)
				strLenResult := fmt.Sprintf("%d", lenResult)
				res.Header().Set("Content-Length", strLenResult)
				res.Header().Set("Content-Type", "text/plain; charset=8")
				res.WriteHeader(http.StatusCreated)

				bytesResult := []byte(strResult)
				res.Write(bytesResult)
			}

		} else {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Вызываемый адрес не существует"))
		}

	} else {

		// мы хотим получить тут адреса типа /....
		re, _ := regexp.Compile(`^/([^/]+)$`)
		resFind := re.FindAllStringSubmatch(req.URL.Path, -1)
		if len(resFind) > 0 && len(resFind[0]) > 0 {

			// получение полной ссылки из короткой
			if req.Method == http.MethodGet {

				shortLink := resFind[0][1]
				shortLink = strings.TrimSpace(shortLink)
				fmt.Println("Короткая ссылка: " + shortLink)

				fullLink, err := dataShortLink.getFullLinkByShort(shortLink)
				fmt.Printf("%+v", dataShortLink)
				if err != nil {
					strErr := err.Error()

					res.Header().Set("Content-Type", "text/plain; charset=utf-8")
					res.WriteHeader(http.StatusBadRequest)
					res.Write([]byte(strErr))
				} else {
					res.Header().Set("Location", fullLink)
					res.WriteHeader(http.StatusTemporaryRedirect)
				}
			} else {
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte("Вызываемый адрес не существует"))
			}

		} else {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("Вызываемый адрес не существует"))
		}

	}
}

// инициализируем хранилище ссылок
var dataShortLink = NewDataShortLink()

func main() {

	muxApp := http.NewServeMux()
	muxApp.HandleFunc(`/`, mainPage)
	err := http.ListenAndServe(`:8080`, muxApp)
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

// shortenertest -test.v -test.run=^TestIteration1$ -binary-path=C:\Users\LENOVO\OneDrive\Desktop\golang\project\go-url-shortener\cmd\shortener\shortener\main.exe
