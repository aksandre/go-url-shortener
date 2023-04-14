package handlers

import (
	"fmt"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/logger"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
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

type dataHandler struct {
	service service.ServiceShortInterface
}

func (dh dataHandler) getServiceLinkByUrl(res http.ResponseWriter, req *http.Request) {
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
	serviceLink, err := dh.service.GetServiceLinkByUrl(urlFull, hostService)
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

}

func (dh dataHandler) getFullLinkByShort(res http.ResponseWriter, req *http.Request) {
	shortLink := chi.URLParam(req, "shortLink")
	shortLink = strings.TrimSpace(shortLink)
	logger.AppLogger.Printf("Пришла короткая ссылка: %s", shortLink)

	fullLink, err := dh.service.GetFullLinkByShort(shortLink)
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
}

func NewRouterHandler(serviceShortLink service.ServiceShortInterface) http.Handler {

	// создаем структуру с данными о сервисе
	var dataHandler = dataHandler{
		service: serviceShortLink,
	}

	router := chi.NewRouter()
	router.Post("/", dataHandler.getServiceLinkByUrl)
	router.Get("/{shortLink}", dataHandler.getFullLinkByShort)

	// когда метод не найден, то 400
	funcNotFoundMethod := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Вызываемый адрес не существует"))
	})
	router.NotFound(funcNotFoundMethod)
	router.MethodNotAllowed(funcNotFoundMethod)

	return router
}
