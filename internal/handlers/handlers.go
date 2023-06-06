package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-url-shortener/internal/logger"
	modelsRequests "go-url-shortener/internal/models/requests"
	modelsResponses "go-url-shortener/internal/models/responses"
	modelsService "go-url-shortener/internal/models/service"
	modelsStorage "go-url-shortener/internal/models/storageshortlink"
	"io"
	"strconv"

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
	service modelsService.ServiceShortInterface
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

	logger.GetLogger().Debugf("Список URL запрошенных пользователем: %+v", listFullURL)

	ctx := context.TODO()
	listShortLinks, err := dh.service.GetDataShortLinks(ctx, listFullURL)
	//logger.GetLogger().Debugf("Данные коротких ссылок пользователя в хранилище: %+v", listShortLinks)

	if err != nil {
		err = fmt.Errorf("ошибка получения списка коротких ссылок: %w", err)
		strError := err.Error()
		logger.GetLogger().Errorf("%s", strError)

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

// Записываем в ответ ошибочное сообщение в JSON виде
func writeErrorJSONResponse(err error, res http.ResponseWriter) {
	err = fmt.Errorf("ошибка сериализации тела запроса: %w", err)
	strError := err.Error()
	logger.GetLogger().Errorf("%s", strError)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(strError))
}

// Записываем в ответ успешное сообщение в JSON виде
func writeSuccessJSONResponse(serviceLink string, statusResponse int, res http.ResponseWriter) {
	// данные ответа
	dataResponse := modelsResponses.ResponseServiceLink{
		Result: serviceLink,
	}
	bytesResult, _ := json.Marshal(&dataResponse)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusResponse)
	res.Write(bytesResult)
}

// Получаем из тестового тела запроса URL для генерации короткой ссылки
func getFullURLFromJSONBody(res http.ResponseWriter, req *http.Request) (urlFull string, err error) {
	// получаем тело из запроса
	dataBody := req.Body

	// данные запроса
	dataRequest := modelsRequests.RequestServiceLink{}
	if err = json.NewDecoder(dataBody).Decode(&dataRequest); err != nil {
		return
	}

	// получаем адрес для которого формируем короткую ссылку
	urlFull = string(dataRequest.URL)
	urlFull = strings.TrimSpace(urlFull)
	if len(urlFull) == 0 {
		err = errors.New("ошибка: в запросе не указан URL, для которого надо сгенерировать короткую ссылку")
	}

	logger.GetLogger().Debugf("Из запроса пришел Url: %s", urlFull)

	return
}

// Умное получение короткой ссылки по Json запросу
// Если ссылка не существует, то создаем
func (dh dataHandler) getServiceLinkByJSON(res http.ResponseWriter, req *http.Request) {

	urlFull, err := getFullURLFromJSONBody(res, req)
	if err != nil {
		writeErrorTextResponse(err, res)
		return
	}

	ctx := context.TODO()
	serviceLink, err := dh.service.GetServiceLinkByURL(ctx, urlFull)
	logger.GetLogger().Debugf("Сделали короткую ссылку: %s", serviceLink)

	isErrExist := errors.Is(err, modelsStorage.ErrExistFullURL)
	if err == nil || isErrExist {
		// добавляем ссылку в данные пользователя
		errAdd := cookiesUserData.AddListFullURLToUser([]string{urlFull}, res, req)
		if errAdd != nil {
			logger.GetLogger().Errorf("Ошибка сохранения у пользователя списка запрошенных коротких ссылок %s :", errAdd.Error())
		}

		statusResponse := http.StatusOK

		// записываем успешный ответ
		writeSuccessJSONResponse(serviceLink, statusResponse, res)

	} else {
		writeErrorJSONResponse(err, res)
	}
}

// Добавление нового URL в сервис по Json запросу
func (dh dataHandler) addNewFullURLByJSON(res http.ResponseWriter, req *http.Request) {

	urlFull, err := getFullURLFromJSONBody(res, req)
	if err != nil {
		writeErrorTextResponse(err, res)
		return
	}

	ctx := context.TODO()
	serviceLink, err := dh.service.AddNewFullURL(ctx, urlFull)
	logger.GetLogger().Debugf("Сделали короткую ссылку: %s", serviceLink)
	fmt.Printf("Сделали короткую ссылку: %s \n", serviceLink)

	isErrExist := errors.Is(err, modelsStorage.ErrExistFullURL)
	if err == nil || isErrExist {
		// добавляем ссылку в данные пользователя
		errAdd := cookiesUserData.AddListFullURLToUser([]string{urlFull}, res, req)
		if errAdd != nil {
			logger.GetLogger().Errorf("Ошибка сохранения у пользователя списка запрошенных коротких ссылок %s :", errAdd.Error())
		}

		statusResponse := http.StatusCreated
		if isErrExist {
			statusResponse = http.StatusConflict
		}

		// записываем успешный ответ
		writeSuccessJSONResponse(serviceLink, statusResponse, res)

	} else {
		writeErrorJSONResponse(err, res)
	}
}

// Генерация группы коротких ссылок по Json запросу
func (dh dataHandler) getBatchServiceLinkByJSON(res http.ResponseWriter, req *http.Request) {

	// получаем тело из запроса
	dataBody := req.Body

	// данные запроса
	dataBatchRequest := modelsRequests.RequestBatchServiceLinks{}
	jsonDecoder := json.NewDecoder(dataBody)
	err := jsonDecoder.Decode(&dataBatchRequest)
	if err != nil {
		err = fmt.Errorf("ошибка сериализации тела запроса: %w", err)
		strError := err.Error()
		logger.GetLogger().Errorf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return
	}

	debugInput := []string{}
	// слайс ссылок
	listFullURLs := []string{}
	// соответствие id коррелирования и ссылки на сторонний ресурс
	correlationMap := map[string]string{}
	for _, rowBatch := range dataBatchRequest {
		idCorrelation := rowBatch.CorrelationID
		if idCorrelation == "" {
			debugInput = append(debugInput, "пустой correlation_id")
			continue
		}

		urlFull := rowBatch.OriginalURL
		urlFull = strings.TrimSpace(urlFull)
		if urlFull != "" {
			correlationMap[idCorrelation] = urlFull
			listFullURLs = append(listFullURLs, urlFull)
		} else {
			debugInput = append(debugInput, "у correlation_id = "+idCorrelation+" пустой original_url")
		}
	}

	if len(dataBatchRequest) == 0 || len(correlationMap) == 0 {
		strError := "Ошибка создания группы коротких ссылок: "
		strError += "В запросе все данные пустые"
		logger.GetLogger().Errorf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return
	}

	if len(debugInput) > 0 {
		strDebugInput := "Замечания у входных данных при создании группы коротких ссылок: \n"
		for key, value := range debugInput {
			strDebugInput += strconv.Itoa(key) + ") " + value + "\n; "
		}
		logger.GetLogger().Debugf("%s", strDebugInput)
	}

	ctx := context.TODO()

	batchLinks, err := dh.service.GetBatchShortLink(ctx, listFullURLs)
	logger.GetLogger().Debugf("Сформировали для группы короткие ссылки: %+v", batchLinks)

	if err != nil {
		err = fmt.Errorf("ошибка создания группы коротких ссылок : %w", err)
		strError := err.Error()
		logger.GetLogger().Errorf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(strError))
		return

	} else {

		// добавляем ссылки в данные пользователя
		listFullURLs := []string{}
		for urlFull := range batchLinks {
			listFullURLs = append(listFullURLs, urlFull)
		}
		cookiesUserData.AddListFullURLToUser(listFullURLs, res, req)
		errAdd := cookiesUserData.AddListFullURLToUser(listFullURLs, res, req)
		if errAdd != nil {
			logger.GetLogger().Errorf("Ошибка сохранения у пользователя списка запрошенных коротких ссылок %s :", errAdd.Error())
		}

		// данные ответа
		dataResponse := make(modelsResponses.ResponseBatchServiceLinks, len(correlationMap))
		i := 0
		for idCorrelation, urlFull := range correlationMap {
			shortLink, ok := batchLinks[urlFull]
			if ok {
				dataResponse[i] = modelsResponses.RowBatchServiceLink{
					CorrelationID: idCorrelation,
					ShortURL:      shortLink,
				}
				i++
			}
		}

		buf := bytes.Buffer{}
		jsonEncoder := json.NewEncoder(&buf)
		jsonEncoder.Encode(dataResponse)

		bytesResult := buf.Bytes()

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)

		res.Write(bytesResult)
	}

}

// Записываем в ответ ошибочное сообщение в текстовом виде
func writeErrorTextResponse(err error, res http.ResponseWriter) {
	strError := err.Error()
	logger.GetLogger().Errorf("Ошибка чтения тела запроса: %s", strError)

	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(strError))
}

// Записываем в ответ успешное сообщение в текстовом виде
func writeSuccessTextResponse(result string, statusResponse int, res http.ResponseWriter) {
	lenResult := len(result)
	strLenResult := fmt.Sprintf("%d", lenResult)
	res.Header().Set("Content-Length", strLenResult)
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")

	res.WriteHeader(statusResponse)

	bytesResult := []byte(result)
	res.Write(bytesResult)
}

// Получаем из тестового тела запроса URL для генерации короткой ссылки
func getFullURLFromTextBody(res http.ResponseWriter, req *http.Request) (urlFull string, err error) {
	// получаем тело из запроса и проводим его к строке
	dataBody := req.Body
	resultRead, err := io.ReadAll(dataBody)
	dataBody.Close()

	if err != nil {
		return
	}

	urlFull = string(resultRead)
	urlFull = strings.TrimSpace(urlFull)
	if len(urlFull) == 0 {
		err = errors.New("ошибка: в запросе не указан URL, для которого надо сгенерировать короткую ссылку")
	}

	logger.GetLogger().Debugf("Из запроса пришел Url: %s", urlFull)

	return
}

// Умное получение короткой ссылки в текстовом виде
// Если ссылка не существует, то создаем
func (dh dataHandler) getServiceLinkByURL(res http.ResponseWriter, req *http.Request) {

	// получаем тело из запроса и проводим его к строке
	urlFull, err := getFullURLFromTextBody(res, req)
	if err != nil {
		writeErrorTextResponse(err, res)
		return
	}

	ctx := context.TODO()
	serviceLink, err := dh.service.GetServiceLinkByURL(ctx, urlFull)
	logger.GetLogger().Debugf("Сделали короткую ссылку: %s", serviceLink)

	isErrExist := errors.Is(err, modelsStorage.ErrExistFullURL)
	if err == nil || isErrExist {
		// добавляем ссылку в данные пользователя
		errAdd := cookiesUserData.AddListFullURLToUser([]string{urlFull}, res, req)
		if errAdd != nil {
			logger.GetLogger().Errorf("Ошибка сохранения у пользователя списка запрошенных коротких ссылок %s :", errAdd.Error())
		}

		statusResponse := http.StatusOK

		// записываем успешный ответ
		writeSuccessTextResponse(serviceLink, statusResponse, res)

	} else {
		writeErrorTextResponse(err, res)
	}
}

// Добавление нового URL в текстовом виде
func (dh dataHandler) addNewFullURLByURL(res http.ResponseWriter, req *http.Request) {

	// получаем тело из запроса и проводим его к строке
	urlFull, err := getFullURLFromTextBody(res, req)
	if err != nil {
		writeErrorTextResponse(err, res)
		return
	}

	ctx := context.TODO()
	serviceLink, err := dh.service.AddNewFullURL(ctx, urlFull)
	logger.GetLogger().Debugf("Сделали короткую ссылку: %s", serviceLink)

	isErrExist := errors.Is(err, modelsStorage.ErrExistFullURL)
	if err == nil || isErrExist {
		// добавляем ссылку в данные пользователя
		errAdd := cookiesUserData.AddListFullURLToUser([]string{urlFull}, res, req)
		if errAdd != nil {
			logger.GetLogger().Errorf("Ошибка сохранения у пользователя списка запрошенных коротких ссылок %s :", errAdd.Error())
		}

		statusResponse := http.StatusCreated
		if isErrExist {
			statusResponse = http.StatusConflict
		}

		// записываем успешный ответ
		writeSuccessTextResponse(serviceLink, statusResponse, res)

	} else {
		writeErrorTextResponse(err, res)
	}
}

// Получение Url-адреса по короткой ссылке
func (dh dataHandler) getFullLinkByShort(res http.ResponseWriter, req *http.Request) {
	shortLink := chi.URLParam(req, "shortLink")
	shortLink = strings.TrimSpace(shortLink)
	logger.GetLogger().Debugf("Пришла короткая ссылка: %s", shortLink)
	fmt.Printf("Пришла короткая ссылка: %s \n", shortLink)

	ctx := context.TODO()
	fullLink, err := dh.service.GetFullLinkByShort(ctx, shortLink)
	logger.GetLogger().Debugf("Получили полную ссылку: %s", fullLink)
	fmt.Printf("Получили полную ссылку: %s \n", fullLink)

	if err != nil {
		strErr := err.Error()
		logger.GetLogger().Errorf("Ошибка получения полной ссылки: %s", strErr)

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
		logger.GetLogger().Errorf("%s", strError)

		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(strError))
	} else {
		res.WriteHeader(http.StatusOK)
	}

}

// создание обработчика запросов
func NewRouterHandler(serviceShortLink modelsService.ServiceShortInterface) http.Handler {

	// создаем структуру с данными о сервисе
	var dataHandler = dataHandler{
		service: serviceShortLink,
	}

	router := chi.NewRouter()
	router.Post("/", dataHandler.addNewFullURLByURL)
	router.Get("/{shortLink}", dataHandler.getFullLinkByShort)
	router.Get("/api/user/urls", dataHandler.getUserListShortLinksByJSON)
	router.Post("/api/shorten", dataHandler.addNewFullURLByJSON)
	router.Post("/api/shorten/batch", dataHandler.getBatchServiceLinkByJSON)
	router.Get("/ping", dataHandler.getStatusPingDB)

	// получение коротких ссылок без ошибок
	router.Post("/getAndAdd/", dataHandler.getServiceLinkByURL)
	router.Post("/api/shorten/getAndAdd/", dataHandler.getServiceLinkByJSON)

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
