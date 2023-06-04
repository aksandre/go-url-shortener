package handlers

import (
	"context"
	"errors"
	"go-url-shortener/internal/app/service"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"
	modelsStorage "go-url-shortener/internal/models/storageshortlink"
	storagerestorer "go-url-shortener/internal/storage/storageshortlink/storagerestorer"
	"os"
	"strings"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

// Это тесты с реальной отправкой данных на сервер
func TestFileRestorerStorageHandlerServer(t *testing.T) {

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
		var storageShortLink, _ = storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)
		err := storageShortLink.ClearStorage(ctx)
		if err != nil {
			logger.GetLogger().Debug("не смогли очистить данные хранилища: " + err.Error())
		}

		dbHandler := dbconn.GetDBHandler()
		dbHandler.Close()
	}()

	// флаг, что критичная ошибка, после нее не будем продолжать тесты
	isCriticalErr := false
	nameMyTest := "Check create File storage"
	t.Run(nameMyTest, func(t *testing.T) {
		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest)
		_, err := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)
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
	storageShortLink, _ := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)

	testFullURL1 := "https://dsdsdsdds_W.com"
	testShortLink1 := "UUUUUUUU"
	storageShortLink.AddShortLinkForURL(ctx, testFullURL1, testShortLink1)

	testFullURL2 := "https://testSite.com"
	testShortLink2 := "RRRTTTTT"
	storageShortLink.AddShortLinkForURL(ctx, testFullURL2, testShortLink2)

	testFullURL3 := "https://testSite222.com"
	testShortLink3 := "RRRTTTTT222"
	storageShortLink.AddShortLinkForURL(ctx, testFullURL3, testShortLink3)

	logger.GetLogger().Debugf("Установили данные хранилища ссылок: %+v", storageShortLink)

	nameMyTest2 := "Check count link after restore"
	t.Run(nameMyTest2, func(t *testing.T) {

		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest2)

		// В винде не чистится хранилище, Винда блокирует файл
		/*
			// новое хранилище, при инициализации должно заполниться
			storageShortLink, _ := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)
			// вверху добавили две ссылки, проверяем, что в хранилище три ссылки
			countLinks, _ := storageShortLink.GetCountLink(ctx)
			data, _ := storageShortLink.GetShortLinks(ctx, nil)
			logger.GetLogger().Debugf("Данные хранилища: %+v", data)
			assert.Equal(t, countLinks, 3)
		*/
		logger.GetLogger().Debugf("### Конец теста: %s", nameMyTest2)

	})

	nameMyTest3 := "get short link with filter"
	t.Run(nameMyTest3, func(t *testing.T) {

		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest3)

		// новое хранилище, при инициализации должно заполниться
		storageShortLink, _ := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)

		// запрашиваем 3 ссылки
		// должны 2 получить, одна не существующая
		testFullURLUnknow := "https://unknow.com"
		options := &modelsStorage.OptionsQuery{
			Filter: modelsStorage.FilterOptionsQuery{
				ListFullURL: []string{testFullURL1, testFullURL3, testFullURLUnknow},
			},
		}

		rows, _ := storageShortLink.GetShortLinks(ctx, options)
		logger.GetLogger().Debugf("Полученные данные после фильтрации: %+v", rows)

		lenRows := len(rows)
		assert.Equal(t, 2, lenRows)

		_, ok1 := rows[testShortLink1]
		assert.Equal(t, true, ok1)
		assert.Equal(t, testFullURL1, rows[testShortLink1].FullURL)

		_, ok2 := rows[testShortLink3]
		assert.Equal(t, true, ok2)
		assert.Equal(t, testFullURL3, rows[testShortLink3].FullURL)

		logger.GetLogger().Debugf("### Конец теста: %s", nameMyTest3)
	})

	nameMyTest4 := "add batch shorts links"
	t.Run(nameMyTest4, func(t *testing.T) {

		// копия существующего в хранилище url
		testDoubleFullURL := testFullURL1
		testShortLinkForDouble := "7788879"

		// набор новых данных
		testFullURL1 := "https://test1.com"
		testShortLink1 := "test1"

		testFullURL2 := "https://test2.com"
		testShortLink2 := "test2"

		testFullURL3 := "https://test3.com"
		testShortLink3 := "test3"

		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest4)

		// новое хранилище, при инициализации должно заполниться
		storageShortLink, _ := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)

		listBatch := modelsStorage.DataStorageShortLink{
			testShortLink1: modelsStorage.RowStorageShortLink{
				ShortLink: testShortLink1,
				FullURL:   testFullURL1,
			},
			testShortLink2: modelsStorage.RowStorageShortLink{
				ShortLink: testShortLink2,
				FullURL:   testFullURL2,
			},
			testShortLink3: modelsStorage.RowStorageShortLink{
				ShortLink: testShortLink3,
				FullURL:   testFullURL3,
			},
			// данные копии
			testShortLinkForDouble: modelsStorage.RowStorageShortLink{
				ShortLink: testShortLinkForDouble,
				FullURL:   testDoubleFullURL,
			},
		}

		// добавим группу коротких ссылок
		err := storageShortLink.AddBatchShortLinks(ctx, listBatch)
		assert.NoError(t, err)

		if err == nil {

			// запрашиваем 3 ссылки
			options := &modelsStorage.OptionsQuery{
				Filter: modelsStorage.FilterOptionsQuery{
					ListFullURL: []string{testFullURL1, testFullURL2, testFullURL3},
				},
			}
			rows, _ := storageShortLink.GetShortLinks(ctx, options)
			logger.GetLogger().Debugf("Полученные данные после фильтрации: %+v", rows)

			lenRows := len(rows)
			assert.Equal(t, 3, lenRows)

			if lenRows > 0 {
				_, ok1 := rows[testShortLink1]
				assert.Equal(t, true, ok1)
				assert.Equal(t, testFullURL1, rows[testShortLink1].FullURL)

				_, ok2 := rows[testShortLink3]
				assert.Equal(t, true, ok2)
				assert.Equal(t, testFullURL3, rows[testShortLink3].FullURL)
			}
		}

		logger.GetLogger().Debugf("### Конец теста: %s", nameMyTest4)
	})

	nameMyTest5 := "add double full url to storage"
	t.Run(nameMyTest5, func(t *testing.T) {

		logger.GetLogger().Debugf("### Начало теста: %s", nameMyTest5)

		// новое хранилище, при инициализации должно заполниться
		storageShortLink, _ := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)

		// копия существующего url
		testDoubleFullURL := testFullURL1
		testShortLink := "какая-то тестовая короткая ссылка"
		errAdd := storageShortLink.AddShortLinkForURL(ctx, testDoubleFullURL, testShortLink)
		// проверяем, что запись дубля вызывает нужную ошибку
		statusAssert := assert.Equal(t, true, errAdd != nil)
		if statusAssert {
			// проверяем, что это именно наша ошибка
			isOk := errors.Is(errAdd, modelsStorage.ErrExistFullURL)
			assert.Equal(t, true, isOk)
		}

		logger.GetLogger().Debugf("### Конец теста: %s", nameMyTest5)
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
			name:   "add and check1 short link into storage for full url",
			method: http.MethodPost,
			url:    "/",
			body:   testFullURL1,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
				body:        "/" + testShortLink1,
			},
		},

		{
			name:   "add and check2 short link into storage for full url",
			method: http.MethodPost,
			url:    "/",
			body:   testFullURL2,
			want: want{
				statusCode:  http.StatusConflict,
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
			storageShortLink, _ := storagerestorer.NewStorageShortsFromFileStorage(pathTempFile)
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
