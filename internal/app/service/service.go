package service

import (
	"context"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"

	"errors"

	modelsStorage "go-url-shortener/internal/models/storageshortlink"
	"math/rand"
	"strings"

	modelsService "go-url-shortener/internal/models/service"
)

func getPackageError(textError string) error {
	textModuleError := "service: " + textError
	return errors.New(textModuleError)
}

// создание сервис коротких ссылок
func NewServiceShortLink(storage modelsStorage.StorageShortInterface, configApp config.ConfigTypeInterface) modelsService.ServiceShortInterface {
	return &ServiceShortLink{
		configApp:       configApp,
		storage:         storage,
		lengthShortLink: 8,
	}
}

type ServiceShortLink struct {
	storage         modelsStorage.StorageShortInterface
	lengthShortLink int
	configApp       config.ConfigTypeInterface
}

func (service *ServiceShortLink) SetLength(length int) {
	service.lengthShortLink = length
}

// Получаем рандомную строку
func (service *ServiceShortLink) getRandString(length int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	result := b.String()
	return result
}

// Получаем хост открытия коротких ссылок
func (service *ServiceShortLink) getHostShortLink() string {
	host := service.configApp.GetHostShortLink()
	return host
}

// Получаем слайс данных коротких ссылок из хранилища
// listFullURL - это слайс полных ссылок, для которых мы получаем списко коротких ссылок
func (service *ServiceShortLink) GetDataShortLinks(ctx context.Context, listFullURL any) (shortLinks modelsService.ListShortLinks, err error) {

	isFilterFullURL := false
	sliceListFullURL, ok := listFullURL.([]string)
	if ok {
		isFilterFullURL = true
	}

	listAllLinks, err := service.storage.GetShortLinks(ctx, nil)
	if err != nil {
		return
	}

	for _, rowData := range listAllLinks {

		isOkRow := true
		if isFilterFullURL {
			isOkRow = false
			if len(sliceListFullURL) > 0 {
				for _, fullURL := range sliceListFullURL {
					if fullURL == rowData.FullURL {
						isOkRow = true
					}
				}
			}
		}

		if isOkRow {
			shortLink := rowData.ShortLink
			shortLink, _ = service.getShortLinkWithHost(shortLink)

			shortLinks = append(shortLinks, modelsService.RowShortLink{
				ShortURL:    shortLink,
				OriginalURL: rowData.FullURL,
			})
		}
	}

	return
}

// Формируем короткую ссылку с хостом по Url-адресу
func (service *ServiceShortLink) getShortLinkWithHost(shortLink string) (shortLinkWithHost string, err error) {
	hostService := service.getHostShortLink()
	shortLinkWithHost = hostService + "/" + shortLink
	return
}

// Добавляем новый Url-адресу и получаем его короткую ссылку
// Если в хранилище уже есть такой URL, то ошибка
func (service *ServiceShortLink) AddNewFullURL(ctx context.Context, fullURL string) (serviceLink string, err error) {

	shortLink, err := service.addNewFullURL(ctx, fullURL)
	// если это ошибка дублирования записи, то поулчаем существующую короткую ссылку
	isErrExist := errors.Is(err, modelsStorage.ErrExistFullURL)
	if err == nil || isErrExist {
		if shortLink != "" {
			serviceLink, _ = service.getShortLinkWithHost(shortLink)
		}
	}

	return
}

func (service *ServiceShortLink) addNewFullURL(ctx context.Context, fullURL string) (shortLink string, err error) {

	//генерируем новую ссылку
	lengthShort := service.lengthShortLink
	shortLink = service.getRandString(lengthShort)
	logger.GetLogger().Debugf("Сформировали новый код  %s", shortLink)

	// добавим короткую ссылку в хранилище
	err = service.storage.AddShortLinkForURL(ctx, fullURL, shortLink)
	if err != nil {
		// если это ошибка дублирования записи, то поулчаем существующую короткую ссылку
		isErrExist := errors.Is(err, modelsStorage.ErrExistFullURL)
		if isErrExist {
			shortLinkExist, errGet := service.storage.GetShortLinkByURL(ctx, fullURL)
			if errGet != nil {
				err = errGet
			} else {
				// присваиваем существующую ссыку в переменную результата
				shortLink = shortLinkExist
			}
		}
	}

	return
}

// Получаем короткую ссылку по Url-адресу
// Если ссылки не существует, то без ошибок добавляем ее
func (service *ServiceShortLink) GetServiceLinkByURL(ctx context.Context, fullURL string) (serviceLink string, err error) {

	shortLink, err := service.getShortLinkByURL(ctx, fullURL)
	if err == nil {
		if shortLink != "" {
			serviceLink, _ = service.getShortLinkWithHost(shortLink)
		}
	}

	return
}

// Получаем короткую ссылку по Url-адресу
func (service *ServiceShortLink) getShortLinkByURL(ctx context.Context, fullURL string) (shortLink string, err error) {

	shortLink, err = service.storage.GetShortLinkByURL(ctx, fullURL)
	if err == nil {
		if shortLink != "" {
			return
		} else {

			//такой url еще не приходил, генерируем новую ссылку
			lengthShort := service.lengthShortLink
			shortLink = service.getRandString(lengthShort)
			logger.GetLogger().Debugf("Сформировали новый код  %s", shortLink)

			// добавим короткую ссылку в хранилище
			err = service.storage.AddShortLinkForURL(ctx, fullURL, shortLink)
			logger.GetLogger().Debugf("Содержание storage %+v", service.storage)
		}
	}

	return
}

// Получаем Url-адрес по короткой ссылке
func (service *ServiceShortLink) GetFullLinkByShort(ctx context.Context, shortLink string) (fullURL string, err error) {

	fullURL, err = service.storage.GetFullLinkByShort(ctx, shortLink)
	if err != nil {
		logger.GetLogger().Errorf("Ошибка при получении полной ссылки: %s", err.Error())
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}

// получение коротких ссылок группой
func (service *ServiceShortLink) GetBatchShortLink(ctx context.Context, listFullURL []string) (resultBatch modelsService.BatchShortLinks, err error) {

	// Из списка запрашиваемых ссылок получим те, которые есть в хранилище
	// Остальные это новые ссылки, сгенерируем для них короткие ссылки

	options := &modelsStorage.OptionsQuery{
		Filter: modelsStorage.FilterOptionsQuery{
			ListFullURL: listFullURL,
		},
	}
	rowsExists, err := service.storage.GetShortLinks(ctx, options)
	if err != nil {
		return
	}

	// создадим map с ключом раынм оригинальному URL , чтобы поиск сделать O(1)
	mapFullURLs := make(modelsStorage.DataStorageShortLink, len(listFullURL))
	for _, dataRow := range rowsExists {
		fullURL := dataRow.FullURL
		mapFullURLs[fullURL] = dataRow
	}

	// инициализируем результирующие данные
	resultBatch = modelsService.BatchShortLinks{}

	listBatchUnknowFullURLs := modelsStorage.DataStorageShortLink{}
	for _, fullURL := range listFullURL {

		dataRow, ok := mapFullURLs[fullURL]
		if !ok {

			// создаем новую короткую ссылку
			lengthShort := service.lengthShortLink
			shortLink := service.getRandString(lengthShort)
			listBatchUnknowFullURLs[shortLink] = modelsStorage.RowStorageShortLink{
				ShortLink: shortLink,
				FullURL:   fullURL,
			}

			// добавляем в итоговый результат
			shortLink, _ = service.getShortLinkWithHost(shortLink)
			resultBatch[fullURL] = shortLink

		} else {
			// берем короткую ссылку из хранилища
			shortLink := dataRow.ShortLink

			// добавляем в итоговый результат
			shortLink, _ = service.getShortLinkWithHost(shortLink)
			resultBatch[fullURL] = shortLink
		}
	}

	// добавим группу коротких ссылок
	err = service.storage.AddBatchShortLinks(ctx, listBatchUnknowFullURLs)
	if err != nil {
		return nil, err
	}

	return
}
