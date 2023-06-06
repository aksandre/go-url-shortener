package handlers

import (
	"context"
	"errors"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"
	modelsService "go-url-shortener/internal/models/service"
	modelsStorage "go-url-shortener/internal/models/storageshortlink"
	storageShort "go-url-shortener/internal/storage/storageshortlink"
	"io"
	"os"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Это тесты без отправки данных на сервер
func TestNewRouterHandlerNoServer(t *testing.T) {

	//--- Start устанавливаем данные конфигурации для теста
	// имя тестовой таблицы
	nameTestTable := "test_table_restore"
	// имя временного файла с хранилищем
	pathTempFile := os.TempDir() + "/storage/testStorage.json"
	configApp := config.GetAppConfig()
	configApp.SetFileStoragePath(pathTempFile)
	configApp.SetNameTableRestorer(nameTestTable)
	// дебаг режим
	configApp.SetLevelLogs(6)
	//--- End устанавливаем данные конфигурации для теста

	// контекст
	ctx := context.TODO()

	defer func() {
		var storageShortLink = storageShort.NewStorageShorts()
		err := storageShortLink.ClearStorage(ctx)
		if err != nil {
			logger.GetLogger().Debug("не смогли очистить данные хранилища: " + err.Error())
		}

		dbHandler := dbconn.GetDBHandler()
		dbHandler.Close()
	}()

	// заполняем данными хранилище
	var storageShortLink = storageShort.NewStorageShorts()
	testFullURL1 := "https://testSite.com"
	testShortLink1 := "RRRTTTTT"
	testFullURL2 := "https://dsdsdsdds.com"
	testShortLink2 := "UUUUUUUU"
	errSet := storageShortLink.SetData(
		ctx,
		modelsStorage.DataStorageShortLink{
			testShortLink1: modelsStorage.RowStorageShortLink{
				ShortLink: testShortLink1,
				FullURL:   testFullURL1,
				UUID:      "1",
			},
			testShortLink2: modelsStorage.RowStorageShortLink{
				ShortLink: testShortLink2,
				FullURL:   testFullURL2,
				UUID:      "2",
			},
		},
	)
	if errSet != nil {
		err := errors.New("Тесты НЕ выполнены, не удалось установить тестовые данные")
		logger.GetLogger().Debugln(err.Error())
		return
	}

	//dataStore, _ := storageShortLink.GetShortLinks(ctx, nil)
	//logger.GetLogger().Debugf("Установили данные хранилища ссылок: %+v", dataStore)

	// сервис на конфиге
	serviceShortLink := service.NewServiceShortLink(storageShortLink, configApp)

	// хост сервиса
	hostService := configApp.GetHostShortLink()

	type want struct {
		statusCode  int
		contentType string
		location    string
		body        string
	}

	tests := []struct {
		name             string
		serviceShortLink modelsService.ServiceShortInterface
		method           string
		url              string
		body             string
		want             want
	}{
		{
			name:             "add new short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             "https://11111google.com/",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "add new2 short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             "https://22222google.com/",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "add double short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             testFullURL1,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "get new short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/getAndAdd/",
			body:             "https://11111google.com/",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "get double short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/getAndAdd/",
			body:             testFullURL1,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "add and check short link into storage for full url",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             testFullURL2,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
				body:        hostService + "/" + testShortLink2,
			},
		},

		{
			name:             "get and check short link into storage for full url",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/getAndAdd/",
			body:             testFullURL2,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        hostService + "/" + testShortLink2,
			},
		},

		{
			name:             "add short link - No Valid method",
			serviceShortLink: serviceShortLink,
			method:           http.MethodGet,
			url:              "/",
			body:             testFullURL2,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "get full url by short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodGet,
			url:              "/" + testShortLink2,
			body:             "",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   testFullURL2,
			},
		},

		{
			name:             "get full url by Unknow short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodGet,
			url:              "/cxzcxcxcx1111",
			body:             "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "get full url - No Valid method",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/" + testShortLink2,
			body:             "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "add new short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten",
			body:             "{\"url\":\"https://1111dsdsdsdds.com\"}",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:             "add double short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten",
			body:             "{\"url\":\"" + testFullURL2 + "\"}",
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json",
			},
		},
		{
			name:             "get short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten/getAndAdd/",
			body:             "{\"url\":\"" + testFullURL2 + "\"}",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				body:        "{\"result\":\"" + hostService + "/" + testShortLink2 + "\"}",
			},
		},
		{
			name:             "add and check WRONG short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten",
			body:             "{\"url\":''}",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "get and check WRONG short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten/getAndAdd/",
			body:             "{\"url\":''}",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "get batch short links from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten/batch",
			body:             "[{\"correlation_id\":\"7777777\",\"original_url\":\"https://testSite.com\"},{\"correlation_id\":\"11111\",\"original_url\":\"22222\"}]",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				// тут один short_url известен, другой не известен
				body: "\"correlation_id\":\"7777777\",\"short_url\":\"" + hostService + "/RRRTTTTT\"}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger.GetLogger().Debugf("### Начало теста: %s", tt.name)

			bodyReader := strings.NewReader(tt.body)
			request := httptest.NewRequest(tt.method, tt.url, bodyReader)
			//request.Header.Add("Accept-Encoding", "gzip")

			respWriter := httptest.NewRecorder()

			handler := NewRouterHandler(tt.serviceShortLink)
			handler.ServeHTTP(respWriter, request)

			res := respWriter.Result()

			// проверяем код ответа
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			// проверяем заголовки ответа
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			}
			if tt.want.location != "" {
				// получаем и проверяем тело запроса
				assert.Equal(t, tt.want.location, res.Header.Get("Location"))
			}

			if tt.want.body != "" {
				// получаем и проверяем тело запроса
				resBody, _ := io.ReadAll(res.Body)
				res.Body.Close()
				strBody := string(resBody)

				// получаем и проверяем тело запроса
				statusAssert := assert.Equal(t, true, strings.Contains(strBody, tt.want.body))
				if !statusAssert {
					logger.GetLogger().Debugf("Получили тело ответа %+v", strBody)
				}
			}

			logger.GetLogger().Debugf("### Конец теста: %s", tt.name)
		})
	}
}
