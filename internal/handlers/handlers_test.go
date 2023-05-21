package handlers

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"
	storageShort "go-url-shortener/internal/storage/storageshortlink"
	"io"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Это тесты без отправки данных на сервер
func TestNewRouterHandler(t *testing.T) {

	defer func() {
		dbHandler := dbconn.GetDBHandler()
		dbHandler.Close()
	}()

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
			name:             "get short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten",
			body:             "{\"url\":\"https://dsdsdsdds.com\"}",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				body:        "{\"result\":\"http://localhost:8080/UUUUUU\"}",
			},
		},
		{
			name:             "check WRONG short link from JSON request",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			url:              "/api/shorten",
			body:             "{\"url\":''}",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
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
				logger.GetLogger().Debugf("Получили тело ответа %+v", strBody)
				// получаем и проверяем тело запроса
				assert.Equal(t, true, strings.Contains(strBody, tt.want.body))
			}

			/*if got := MainPageHandler(tt.args.serviceShortLink); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MainPageHandler() = %v, want %v", got, tt.want)
			}*/

			logger.GetLogger().Debugf("### Конец теста: %s", tt.name)

		})
	}
}
