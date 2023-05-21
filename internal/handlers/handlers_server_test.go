package handlers

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"
	storageShort "go-url-shortener/internal/storage/storageshortlink"
	"strings"

	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// Это тесты с реальной отправкой данных на сервер
func TestNewRouterHandlerServer(t *testing.T) {

	defer func() {
		dbHandler := dbconn.GetDBHandler()
		dbHandler.Close()
	}()

	// заполняем хранилище
	var storageShortLink = storageShort.NewStorageShorts()
	storageShortLink.SetData(
		storageShort.DataStorageShortLink{
			"RRRTTTTT": storageShort.RowStorageShortLink{
				ShortLink: "RRRTTTTT",
				FullURL:   "https://testSite.com",
				UUID:      "1",
			},
			"UUUUUU": storageShort.RowStorageShortLink{
				ShortLink: "UUUUUU",
				FullURL:   "https://dsdsdsdds.com",
				UUID:      "2",
			},
		},
	)
	logger.GetLogger().Debugf("Установили данные хранилища ссылок: %+v", storageShortLink)

	// Создаем конфиг
	configApp := config.GetAppConfig()
	serviceShortLink := service.NewServiceShortLink(storageShortLink, configApp)

	// обработчик запросов
	handlerRouter := NewRouterHandler(serviceShortLink)

	// запускаем тестовый сервер
	serverTest := httptest.NewServer(handlerRouter)
	logger.GetLogger().Debugln("Сервер подняли на адресе: " + serverTest.URL)
	// останавливаем сервер после завершения теста
	defer serverTest.Close()

	type want struct {
		statusCode  int
		contentType string
		location    string
		body        string
	}

	tests := []struct {
		name             string
		serviceShortLink service.ServiceShortInterface
		method           string
		url              string
		body             string
		want             want
	}{
		{
			name:             "get new short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             "https://google.com/",
			want: want{
				statusCode:  201,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:             "get double short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             "https://google.com/",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "check short link into storage for full url",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/",
			body:             "https://dsdsdsdds.com",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
				body:        "/UUUUUU",
			},
		},

		{
			name:             "get short link - No Valid method",
			serviceShortLink: serviceShortLink,
			method:           http.MethodGet,
			url:              "/",
			body:             "https://google.com/",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "get full url by short link",
			serviceShortLink: serviceShortLink,
			method:           http.MethodGet,
			url:              "/UUUUUU",
			body:             "",
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://dsdsdsdds.com",
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
			url:              "/UUUUUU",
			body:             "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},

		{
			name:             "get USER list short links from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodGet,
			url:              "/api/user/urls",
			body:             "",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				body:        "\"original_url\":\"https://google.com/",
			},
		},
		/*
			{
				name:             "check DB",
				serviceShortLink: serviceShortLink,
				method:           http.MethodGet,
				url:              "/ping",
				body:             "",
				want: want{
					statusCode: http.StatusOK,
				},
			},*/
	}

	// создаем cookie jar для сохранения cookies между запросами
	jar, _ := cookiejar.New(nil)
	// не используем реальный редирект
	redirectPolicy := resty.NoRedirectPolicy()
	// Будем делать реальные запросы
	req := resty.New().
		SetRedirectPolicy(redirectPolicy).
		SetCookieJar(jar).
		R().
		SetHeader("Accept-Encoding", "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger.GetLogger().Debugf("### Начало теста: %s", tt.name)

			req.SetBody(tt.body)
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
