package handlers

import (
	"fmt"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/logger"
	"net/http"
	"regexp"
	"strings"
)

func getUrlRequest(req *http.Request) string {
	url := req.Host + req.URL.Path
	return url
}

func getProtocolHttp() string {
	return "http"
}

func getUrlService(req *http.Request) string {
	return getProtocolHttp() + "://" + req.Host
}

func MainPageHandler(serviceShortLink service.ServiceShortInterface) http.HandlerFunc {

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		if req.URL.Path == "/" {

			// генерация короткой ссылки
			if req.Method == http.MethodPost {
				// получаем тело из запроса и проводим его к строке
				dataBody := req.Body
				lenBody := req.ContentLength

				result := make([]byte, lenBody)
				dataBody.Read(result)
				dataBody.Close()

				urlFull := string(result)
				urlFull = strings.TrimSpace(urlFull)
				logger.AppLogger.Printf("Для генерации короткой ссылки пришел Url: %s", urlFull)

				// url сервиса
				hostService := getUrlService(req)
				serviceLink, err := serviceShortLink.GetServiceLinkByUrl(urlFull, hostService)
				logger.AppLogger.Printf("Сделали короткую ссылку: %s", serviceLink)

				if err != nil {
					strError := err.Error()
					logger.AppLogger.Printf("Ошибка создания короткой ссылки : %s", strError)

					res.Header().Set("Content-Type", "text/plain; charset=8")
					res.WriteHeader(http.StatusBadRequest)
					res.Write([]byte(strError))

				} else {
					lenResult := len(serviceLink)
					strLenResult := fmt.Sprintf("%d", lenResult)
					res.Header().Set("Content-Length", strLenResult)
					res.Header().Set("Content-Type", "text/plain; charset=8")
					res.WriteHeader(http.StatusCreated)

					bytesResult := []byte(serviceLink)
					res.Write(bytesResult)
				}

			} else {
				res.Header().Set("Content-Type", "text/plain; charset=utf-8")
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
					logger.AppLogger.Printf("Пришла короткая ссылка: %s", shortLink)

					fullLink, err := serviceShortLink.GetFullLinkByShort(shortLink)
					logger.AppLogger.Printf("Получили полную ссылку: %s", fullLink)

					if err != nil {
						strErr := err.Error()
						logger.AppLogger.Printf("Ошибка получения полной ссылки: %s", strErr)

						res.Header().Set("Content-Type", "text/plain; charset=utf-8")
						res.WriteHeader(http.StatusBadRequest)
						res.Write([]byte(strErr))
					} else {
						res.Header().Set("Location", fullLink)
						res.WriteHeader(http.StatusTemporaryRedirect)
					}
				} else {
					res.Header().Set("Content-Type", "text/plain; charset=utf-8")
					res.WriteHeader(http.StatusBadRequest)
					res.Write([]byte("Вызываемый адрес не существует"))
				}

			} else {
				res.Header().Set("Content-Type", "text/plain; charset=utf-8")
				res.WriteHeader(http.StatusBadRequest)
				res.Write([]byte("Вызываемый адрес не существует"))
			}

		}
	})
}
