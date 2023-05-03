package handlers

import (
	"bytes"
	"compress/gzip"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	storageShort "go-url-shortener/internal/storage/storageshortlink"
	"io"
	"log"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func compressText(text string) []byte {
	rawBytes := []byte(text)

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(rawBytes)
	if err != nil {
		log.Fatal("ошибка: не смогли сжать тело запроса:" + err.Error())
	}
	gzipWriter.Close()
	return buf.Bytes()
}

// Это тесты без отправки данных на сервер
// тесты для проверки сжатия
func TestNewRouterCompressHandler(t *testing.T) {

	var storageShortLink = storageShort.NewStorageShorts()
	storageShortLink.SetData(
		storageShort.DataStorageShortLink{
			"RRRTTTTT": "https://testSite.com",
			"UUUUUU":   "https://dsdsdsdds.com",
		},
	)
	logger.GetLogger().Debugf("Установили данные хранилища ссылок: %+v", storageShortLink)

	// Создаем конфиг
	configApp := config.GetAppConfig()
	serviceShortLink := service.NewServiceShortLink(storageShortLink, configApp)

	type want struct {
		statusCode      int
		contentType     string
		location        string
		body            string
		decompress_body string
		headersResponse http.Header
	}

	tests := []struct {
		name             string
		serviceShortLink service.ServiceShortInterface
		headers          http.Header
		method           string
		url              string
		body             string
		want             want
	}{
		{
			name:             "compress: get new short link - Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"text/plain"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/",
			body: "https://dsdsdsdds.com",
			want: want{
				statusCode:  201,
				contentType: "text/plain; charset=utf-8",
				body:        "/UUUUUU",
			},
		},

		{
			name:             "compress: get new short link by JSON - Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"application/json"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/api/shorten",
			body: "{\"url\":\"https://dsdsdsdds.com\"}",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				body:        "{\"result\":\"http://localhost:8080/UUUUUU\"}",
			},
		},

		{
			name:             "compress: get new short link - Accept-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":    []string{"text/plain"},
				"Accept-Encoding": []string{"gzip"},
			},
			url:  "/",
			body: "https://dsdsdsdds.com",
			want: want{
				statusCode:      201,
				contentType:     "text/plain; charset=utf-8",
				decompress_body: "/UUUUUU",
				headersResponse: http.Header{
					"Content-Encoding": []string{"gzip"},
				},
			},
		},

		{
			name:             "compress: get new short link by JSON - Accept-Encoding: gzip + Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"application/json"},
				"Accept-Encoding":  []string{"gzip"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/api/shorten",
			body: "{\"url\":\"https://dsdsdsdds.com\"}",
			want: want{
				statusCode:      http.StatusCreated,
				contentType:     "application/json",
				decompress_body: "{\"result\":\"http://localhost:8080/UUUUUU\"}",
				headersResponse: http.Header{
					"Content-Encoding": []string{"gzip"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bodyReader := io.Reader(strings.NewReader(tt.body))

			// если мы хотим сжимать тело запроса
			acceptEncodingRequest := tt.headers.Get("Content-Encoding")
			if tt.body != "" && strings.Contains(acceptEncodingRequest, "gzip") {
				// Сжимаем тело запроса
				comressBoby := compressText(tt.body)
				bodyReader = bytes.NewReader(comressBoby)
			}

			request := httptest.NewRequest(tt.method, tt.url, bodyReader)
			for nameHeader, valuesHeader := range tt.headers {
				for _, valueHeader := range valuesHeader {
					request.Header.Add(nameHeader, valueHeader)
				}
			}

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

				//fmt.Println("данные сравниваем: " + string(resBody))
				res.Body.Close()
				// получаем и проверяем тело запроса
				assert.Equal(t, true, strings.Contains(string(resBody), tt.want.body))

			} else if tt.want.decompress_body != "" {

				acceptEncodingRequest := request.Header.Get("Accept-Encoding")
				contentEncodingResponse := res.Header.Get("Content-Encoding")

				isTrueAssert1 := strings.Contains(acceptEncodingRequest, "gzip")
				if isTrueAssert1 {
					isTrueAssert2 := assert.Equal(t, true, strings.Contains(contentEncodingResponse, "gzip"))
					if isTrueAssert2 {
						// переменная readerZip будет читать входящие данные и распаковывать их
						readerZip, err := gzip.NewReader(res.Body)
						defer readerZip.Close()

						if assert.NoError(t, err) {

							bytesBody, _ := io.ReadAll(readerZip)
							stringBody := string(bytesBody)

							// получаем и проверяем тело запроса
							assert.Equal(t, true, strings.Contains(stringBody, tt.want.decompress_body))

						}
					}

				}

			}

		})
	}
}
