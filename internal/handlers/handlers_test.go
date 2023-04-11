package handlers

import (
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/logger"
	storageShort "go-url-shortener/internal/storage/storageShortlink"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainPageHandler(t *testing.T) {

	var storageShortLink = storageShort.NewStorageShorts()
	storageShortLink.SetData(
		storageShort.DataStorageShortLink{
			"RRRTTTTT": "https://testSite.com",
			"UUUUUU":   "https://dsdsdsdds.com",
		},
	)
	logger.AppLogger.Printf("Установили данные хранилища ссылок: %+v", storageShortLink)

	var serviceShortLink = service.NewServiceShortLink(&storageShortLink, 8)

	type want struct {
		statusCode  int
		response    string
		contentType string
		location    string
	}

	tests := []struct {
		name             string
		serviceShortLink *service.ServiceShortLink
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
				contentType: "text/plain; charset=8",
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
				contentType: "text/plain; charset=8",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bodyReader := strings.NewReader(tt.body)
			request := httptest.NewRequest(tt.method, tt.url, bodyReader)
			respWriter := httptest.NewRecorder()

			handlerFunc := MainPageHandler(tt.serviceShortLink)
			handlerFunc(respWriter, request)

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

			/*

				if want.response !== "any" {
					// получаем и проверяем тело запроса
					defer respWriter.Body.Close()
					resBody, err := io.ReadAll(res.Body)
				}
			*/

			/*if got := MainPageHandler(tt.args.serviceShortLink); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MainPageHandler() = %v, want %v", got, tt.want)
			}*/

		})
	}
}
