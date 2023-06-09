package handlers

import (
	"bytes"
	"compress/gzip"
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
	"log"
	"os"

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

	// печатаем расположение лога после инициализации конфига
	logger.GetLogger().Debugf("Путь до файла хранилища ссылок: %+v", pathTempFile)

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

	// инициализируем сервис на базе конфига
	serviceShortLink := service.NewServiceShortLink(storageShortLink, configApp)

	// хост сервиса
	hostService := configApp.GetHostShortLink()

	type want struct {
		statusCode      int
		contentType     string
		location        string
		body            string
		decompressBody  string
		headersResponse http.Header
	}

	tests := []struct {
		name             string
		serviceShortLink modelsService.ServiceShortInterface
		headers          http.Header
		method           string
		url              string
		body             string
		want             want
	}{
		{
			name:             "compress: get short link - Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"text/plain"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/getAndAdd/",
			body: testFullURL2,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        "/" + testShortLink2,
			},
		},
		{
			name:             "compress: add double short link - Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"text/plain"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/",
			body: testFullURL2,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
				body:        "/" + testShortLink2,
			},
		},

		{
			name:             "compress: add double short link by JSON - Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"application/json"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/api/shorten",
			body: "{\"url\":\"" + testFullURL1 + "\"}",
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json",
				body:        "{\"result\":\"" + hostService + "/" + testShortLink1 + "\"}",
			},
		},

		{
			name:             "compress: add double short link - Accept-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":    []string{"text/plain"},
				"Accept-Encoding": []string{"gzip"},
			},
			url:  "/",
			body: testFullURL2,
			want: want{
				statusCode:     http.StatusConflict,
				contentType:    "text/plain; charset=utf-8",
				decompressBody: "/" + testShortLink2,
				headersResponse: http.Header{
					"Content-Encoding": []string{"gzip"},
				},
			},
		},

		{
			name:             "compress: add double short link by JSON - Accept-Encoding: gzip + Content-Encoding: gzip",
			serviceShortLink: serviceShortLink,
			method:           http.MethodPost,
			headers: http.Header{
				"Content-Type":     []string{"application/json"},
				"Accept-Encoding":  []string{"gzip"},
				"Content-Encoding": []string{"gzip"},
			},
			url:  "/api/shorten",
			body: "{\"url\":\"" + testFullURL2 + "\"}",
			want: want{
				statusCode:     http.StatusConflict,
				contentType:    "application/json",
				decompressBody: "{\"result\":\"" + hostService + "/" + testShortLink2 + "\"}",
				headersResponse: http.Header{
					"Content-Encoding": []string{"gzip"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger.GetLogger().Debugf("### Начало теста: %s", tt.name)

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

			} else if tt.want.decompressBody != "" {

				acceptEncodingRequest := request.Header.Get("Accept-Encoding")
				contentEncodingResponse := res.Header.Get("Content-Encoding")

				isTrueAssert1 := strings.Contains(acceptEncodingRequest, "gzip")
				if isTrueAssert1 {
					isTrueAssert2 := assert.Equal(t, true, strings.Contains(contentEncodingResponse, "gzip"))
					if isTrueAssert2 {
						// переменная readerZip будет читать входящие данные и распаковывать их
						readerZip, err := gzip.NewReader(res.Body)
						defer func() {
							readerZip.Close()
						}()

						if assert.NoError(t, err) {

							bytesBody, _ := io.ReadAll(readerZip)
							stringBody := string(bytesBody)

							// получаем и проверяем тело запроса
							assert.Equal(t, true, strings.Contains(stringBody, tt.want.decompressBody))

						}
					}

				}

			}

			logger.GetLogger().Debugf("### Конец теста: %s", tt.name)

		})
	}
}
