package handlers

import (
	"encoding/json"
	"fmt"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/logger"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type dataHandler struct {
	service service.ServiceShortInterface
}

func (dh dataHandler) getServiceLinkByJson(res http.ResponseWriter, req *http.Request) {
	// получаем тело из запроса и проводим его к строке
	dataBody := req.Body
	lenBody := req.ContentLength

	bytesRawBody := make([]byte, lenBody)
	dataBody.Read(bytesRawBody)
	dataBody.Close()

	dataRequest := struct {
		Url string `json:"url"`
	}{}
	err := json.Unmarshal(bytesRawBody, &dataRequest)
	if err != nil {
		err = fmt.Errorf("Ошибка получения из запроса URl адреса: %w", err)
		strError := err.Error()
		logger.GetLogger().Printf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return
	}

	// получаем адрес для которого формируем короткую ссылку
	urlFull := string(dataRequest.Url)
	urlFull = strings.TrimSpace(urlFull)

	if len(urlFull) < 0 {

		strError := "Ошибка создания короткой ссылки: "
		strError += "В запросе не указан URL, для которого надо сгенерировать короткую ссылку"
		logger.GetLogger().Printf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return
	}

	logger.GetLogger().Printf("Для генерации короткой ссылки пришел Url: %s", urlFull)

	serviceLink, err := dh.service.GetServiceLinkByUrl(urlFull)
	logger.GetLogger().Printf("Сделали короткую ссылку: %s", serviceLink)

	if err != nil {
		err = fmt.Errorf("Ошибка создания короткой ссылки : %w", err)
		strError := err.Error()
		logger.GetLogger().Printf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return

	} else {

		// данные ответа
		dataResponse := struct {
			Result string `json:"result"`
		}{
			Result: serviceLink,
		}
		bytesResult, _ := json.Marshal(dataResponse)

		lenResult := len(string(bytesResult))
		strLenResult := fmt.Sprintf("%d", lenResult)
		res.Header().Set("Content-Length", strLenResult)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		res.Write(bytesResult)
	}

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
	logger.GetLogger().Printf("Для генерации короткой ссылки пришел Url: %s", urlFull)

	serviceLink, err := dh.service.GetServiceLinkByUrl(urlFull)
	logger.GetLogger().Printf("Сделали короткую ссылку: %s", serviceLink)

	if err != nil {
		strError := err.Error()
		logger.GetLogger().Printf("Ошибка создания короткой ссылки : %s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))

	} else {
		lenResult := len(serviceLink)
		strLenResult := fmt.Sprintf("%d", lenResult)
		res.Header().Set("Content-Length", strLenResult)
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusCreated)

		bytesResult := []byte(serviceLink)
		res.Write(bytesResult)
	}

}

func (dh dataHandler) getFullLinkByShort(res http.ResponseWriter, req *http.Request) {
	shortLink := chi.URLParam(req, "shortLink")
	shortLink = strings.TrimSpace(shortLink)
	logger.GetLogger().Printf("Пришла короткая ссылка: %s", shortLink)

	fullLink, err := dh.service.GetFullLinkByShort(shortLink)
	logger.GetLogger().Printf("Получили полную ссылку: %s", fullLink)

	if err != nil {
		strErr := err.Error()
		logger.GetLogger().Printf("Ошибка получения полной ссылки: %s", strErr)

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
	router.Post("/api/shorten", dataHandler.getServiceLinkByJson)
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

/*
func getProtocolHttp() string {
	return "http"
}

func getUrlService(req *http.Request) string {
	return getProtocolHttp() + "://" + req.Host
}
*/
