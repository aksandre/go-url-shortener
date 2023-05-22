package handlers

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"
	storageShort "go-url-shortener/internal/storage/storageshortlink"
	"strings"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// Это тесты с реальной отправкой данных на сервер
func TestDBRestorerStorageHandlerServer(t *testing.T) {

	// получаем значение подключения к БД
	// если установлено, но запускаем тесты
	configDatabaseDsn := config.GetAppConfig().GetDatabaseDsn()
	if configDatabaseDsn == "" {
		logger.GetLogger().Debugf("Тесты работы с базой данных не выполнялись: пустое значение DatabaseDsn")
		return
	}

	// имя тестовой таблицы
	nameTableRestorer := "temp_table_restore"

	// флаг, что критичная ошибка, после нее не будем продолжать тесты
	isCriticalErr := false
	nameMyTest := "Check create DB storage"
	t.Run(nameMyTest, func(t *testing.T) {
		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest)
		_, err := storageShort.NewStorageShortsFromDB(nameTableRestorer)
		assert.NoError(t, err)
		logger.GetLogger().Debugf("### Конец теста: %s", nameMyTest)

		if err != nil {
			isCriticalErr = true
		}
	})
	if isCriticalErr {
		return
	}

	// создаем пустое хранилище
	storageShortLink, _ := storageShort.NewStorageShortsFromDB(nameTableRestorer)

	testFullURL1 := "https://dsdsdsdds.com"
	testShortLink1 := "UUUUUUUU"
	storageShortLink.AddShortLinkForURL(testFullURL1, testShortLink1)

	testFullURL2 := "https://testSite.com"
	testShortLink2 := "RRRTTTTT"
	storageShortLink.AddShortLinkForURL(testFullURL2, testShortLink2)

	logger.GetLogger().Debugf("Установили данные хранилища ссылок: %+v", storageShortLink)

	defer func() {
		storageShortLink, _ := storageShort.NewStorageShortsFromDB(nameTableRestorer)
		dbRestorer, _ := storageShortLink.GetRestorer()
		err := dbRestorer.ClearRows()
		if err != nil {
			logger.GetLogger().Debug("не смогли очистить данные хранилища: " + err.Error())
		}

		dbHandler := dbconn.GetDBHandler()
		dbHandler.Close()
	}()

	nameMyTest2 := "Check count link after restore"
	t.Run(nameMyTest, func(t *testing.T) {

		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest2)

		// новое хранилище, при инициализации должно заполниться
		storageShortLink, _ := storageShort.NewStorageShortsFromDB(nameTableRestorer)
		// вверху добавили две ссылки, проверяем, что в хранилище две ссылки
		countLinks, _ := storageShortLink.GetCountLink()
		assert.Equal(t, countLinks, 2)

		logger.GetLogger().Debugf("### Конец теста: %s", nameMyTest)
	})

	type want struct {
		statusCode  int
		contentType string
		location    string
		body        string
	}

	tests := []struct {
		name   string
		method string
		url    string
		body   string
		want   want
	}{
		{
			name:   "check1 short link into storage for full url",
			method: http.MethodPost,
			url:    "/",
			body:   testFullURL1,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
				body:        "/" + testShortLink1,
			},
		},

		{
			name:   "check2 short link into storage for full url",
			method: http.MethodPost,
			url:    "/",
			body:   testFullURL2,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
				body:        "/" + testShortLink2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger.GetLogger().Debugf("### Начало теста: %s", tt.name)

			configApp := config.GetAppConfig()

			// новое хранилище, при инициализации должно заполниться
			storageShortLink, _ := storageShort.NewStorageShortsFromDB(nameTableRestorer)
			logger.GetLogger().Debugf("Восстановленные данные хранилища ссылок: %+v", storageShortLink)

			serviceShortLink := service.NewServiceShortLink(storageShortLink, configApp)
			handlerRouter := NewRouterHandler(serviceShortLink)
			serverTest := httptest.NewServer(handlerRouter)

			logger.GetLogger().Debugln("Сервер подняли на адресе: " + serverTest.URL)
			defer serverTest.Close()

			// не используем реальный редирект
			redirectPolicy := resty.NoRedirectPolicy()

			// Будем делать реальные запросы
			req := resty.New().
				SetRedirectPolicy(redirectPolicy).
				R().
				SetHeader("Accept-Encoding", "").
				SetBody(tt.body)

			req.Method = tt.method
			req.URL = serverTest.URL + tt.url
			res, _ := req.Send()

			// проверяем код ответа
			assert.Equal(t, tt.want.statusCode, res.StatusCode())

			// проверяем заголовки ответа
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			}

			if tt.want.location != "" {
				// получаем и проверяем тело запроса
				assert.Equal(t, tt.want.location, res.Header().Get("Location"))
			}

			if tt.want.body != "" {
				// получаем и проверяем тело запроса
				resBody := res.Body()
				res.RawBody().Close()
				// получаем и проверяем тело запроса

				if !assert.Equal(t, true, strings.Contains(string(resBody), tt.want.body)) {
					logger.GetLogger().Debug("Пришло тело ответа:" + string(resBody) + ", ожидали: " + tt.want.body)
				}

			}

			logger.GetLogger().Debugf("### Конец теста: %s", tt.name)
		})
	}
}
