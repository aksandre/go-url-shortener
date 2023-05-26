package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/logger"
	modelsRequests "go-url-shortener/internal/models/requests"
	modelsResponses "go-url-shortener/internal/models/responses"
	"io"

	middlewareCompress "go-url-shortener/internal/middlewares/compress"
	middlewareLogging "go-url-shortener/internal/middlewares/logging"
	cookiesUserData "go-url-shortener/internal/userdata/usercookies"

	connDB "go-url-shortener/internal/database/connect"

	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Тип обрабочика маршрутов
type dataHandler struct {
	service service.ServiceShortInterface
}

// получаем список ссылок, которые генерировал пользователь
func (dh dataHandler) getUserListShortLinksByJSON(res http.ResponseWriter, req *http.Request) {

	listFullURL := []string{}
	userData, err := cookiesUserData.GetCookiesUserData(req)
	if err != nil {
		logger.GetLogger().Error("Ошибка получения данных пользователя из cookies: " + err.Error())
	} else {
		listFullURL = userData.ListFullURL
	}

	ctx := context.TODO()
	listShortLinks, err := dh.service.GetDataShortLinks(ctx, listFullURL)
	logger.GetLogger().Debugf("Список существующих коротких ссылок %+v", listShortLinks)

	if err != nil {
		err = fmt.Errorf("ошибка поулучения списка коротких ссылок: %w", err)
		strError := err.Error()
		logger.GetLogger().Debugf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return

	} else {

		// данные ответа
		bytesResult, _ := json.Marshal(&listShortLinks)
		res.Header().Set("Content-Type", "application/json")

		if len(listShortLinks) == 0 {
			res.WriteHeader(http.StatusNoContent)
		} else {
			res.WriteHeader(http.StatusOK)
		}

		res.Write(bytesResult)
	}
}

// Генерация короткой ссылки по Json запросу
func (dh dataHandler) getServiceLinkByJSON(res http.ResponseWriter, req *http.Request) {

	// получаем тело из запроса
	dataBody := req.Body

	// данные запроса
	dataRequest := modelsRequests.RequestServiceLink{}
	if err := json.NewDecoder(dataBody).Decode(&dataRequest); err != nil {
		err = fmt.Errorf("ошибка сериализации тела запроса: %w", err)
		strError := err.Error()
		logger.GetLogger().Debugf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return
	}

	// получаем адрес для которого формируем короткую ссылку
	urlFull := string(dataRequest.URL)
	urlFull = strings.TrimSpace(urlFull)

	if len(urlFull) == 0 {

		strError := "Ошибка создания короткой ссылки: "
		strError += "В запросе не указан URL, для которого надо сгенерировать короткую ссылку"
		logger.GetLogger().Debugf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return
	}

	logger.GetLogger().Debugf("Для генерации короткой ссылки пришел Url: %s", urlFull)

	ctx := context.TODO()
	serviceLink, err := dh.service.GetServiceLinkByURL(ctx, urlFull)
	logger.GetLogger().Debugf("Сделали короткую ссылку: %s", serviceLink)

	if err != nil {
		err = fmt.Errorf("ошибка создания короткой ссылки : %w", err)
		strError := err.Error()
		logger.GetLogger().Debugf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return

	} else {

		// добавляем ссылку в данные пользователя
		cookiesUserData.AddFullURLToUser(urlFull, res, req)

		// данные ответа
		dataResponse := modelsResponses.ResponseServiceLink{
			Result: serviceLink,
		}
		bytesResult, _ := json.Marshal(&dataResponse)

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		res.Write(bytesResult)
	}

}

// Генерация короткой ссылки по Url в текстовом виде
func (dh dataHandler) getServiceLinkByURL(res http.ResponseWriter, req *http.Request) {
	// получаем тело из запроса и проводим его к строке
	dataBody := req.Body
	result, err := io.ReadAll(dataBody)
	dataBody.Close()
	if err != nil {
		strError := err.Error()
		logger.GetLogger().Debugf("Ошибка чтения тела запроса: %s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
	}

	urlFull := string(result)
	urlFull = strings.TrimSpace(urlFull)
	logger.GetLogger().Debugf("Для генерации короткой ссылки пришел Url: %s", urlFull)

	ctx := context.TODO()
	serviceLink, err := dh.service.GetServiceLinkByURL(ctx, urlFull)
	logger.GetLogger().Debugf("Сделали короткую ссылку: %s", serviceLink)

	if err != nil {
		strError := err.Error()
		logger.GetLogger().Debugf("Ошибка создания короткой ссылки : %s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))

	} else {

		// добавляем ссылку в данные пользователя
		cookiesUserData.AddFullURLToUser(urlFull, res, req)

		lenResult := len(serviceLink)
		strLenResult := fmt.Sprintf("%d", lenResult)
		res.Header().Set("Content-Length", strLenResult)
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusCreated)

		bytesResult := []byte(serviceLink)
		res.Write(bytesResult)
	}

}

// Получение Url-адреса по короткой ссылке
func (dh dataHandler) getFullLinkByShort(res http.ResponseWriter, req *http.Request) {
	shortLink := chi.URLParam(req, "shortLink")
	shortLink = strings.TrimSpace(shortLink)
	logger.GetLogger().Debugf("Пришла короткая ссылка: %s", shortLink)

	ctx := context.TODO()
	fullLink, err := dh.service.GetFullLinkByShort(ctx, shortLink)
	logger.GetLogger().Debugf("Получили полную ссылку: %s", fullLink)

	if err != nil {
		strErr := err.Error()
		logger.GetLogger().Debugf("Ошибка получения полной ссылки: %s", strErr)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strErr))
	} else {
		res.Header().Set("Location", fullLink)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

// Получение Url-адреса по короткой ссылке
func (dh dataHandler) getStatusPingDB(res http.ResponseWriter, req *http.Request) {

	dbHandler := connDB.GetDBHandler()
	err := dbHandler.GetErrSetup()
	if err != nil {
		err = dbHandler.Ping()
	}

	if err != nil {
		err = fmt.Errorf("ошибка: пинг БД завершился ошибкой: %w", err)
		strError := err.Error()
		logger.GetLogger().Debugf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(strError))
	} else {
		res.WriteHeader(http.StatusOK)
	}

}

// создание обработчика запросов
func NewRouterHandler(serviceShortLink service.ServiceShortInterface) http.Handler {

	// создаем структуру с данными о сервисе
	var dataHandler = dataHandler{
		service: serviceShortLink,
	}

	router := chi.NewRouter()
	router.Post("/", dataHandler.getServiceLinkByURL)
	router.Get("/{shortLink}", dataHandler.getFullLinkByShort)
	router.Get("/api/user/urls", dataHandler.getUserListShortLinksByJSON)
	router.Post("/api/shorten", dataHandler.getServiceLinkByJSON)
	router.Get("/ping", dataHandler.getStatusPingDB)

	// когда метод не найден, то 400
	funcNotFoundMethod := http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Вызываемый адрес не существует"))
	})
	router.NotFound(funcNotFoundMethod)
	router.MethodNotAllowed(funcNotFoundMethod)

	// применяем к обработчику запросов логирование
	handlerRoute := http.Handler(router)
	handlerRoute = middlewareLogging.WrapLogging(middlewareCompress.WrapCompression(handlerRoute))

	return handlerRoute
}
