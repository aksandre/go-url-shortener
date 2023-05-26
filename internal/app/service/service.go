package service

import (
	"context"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	modelsResponses "go-url-shortener/internal/models/responses"
	modelsStorage "go-url-shortener/internal/models/storageshortlink"

	"errors"
	"math/rand"
	"strings"
)

func getPackageError(textError string) error {
	textModuleError := "service: " + textError
	return errors.New(textModuleError)
}

// создание сервис коротких ссылок
func NewServiceShortLink(storage modelsStorage.StorageShortInterface, configApp config.ConfigTypeInterface) ServiceShortInterface {
	return &ServiceShortLink{
		configApp:       configApp,
		storage:         storage,
		lengthShortLink: 8,
	}
}

type RowShortLink modelsResponses.ResponseListShortLinks
type ListShortLinks []RowShortLink

type ServiceShortInterface interface {
	GetServiceLinkByURL(ctx context.Context, fullURL string) (serviceLink string, err error)
	GetFullLinkByShort(ctx context.Context, shortLink string) (fullURL string, err error)
	GetDataShortLinks(ctx context.Context, listFullURL any) (shortLinks ListShortLinks, err error)
	getHostShortLink() string
	SetLength(length int)
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
func (service *ServiceShortLink) GetDataShortLinks(ctx context.Context, listFullURL any) (shortLinks ListShortLinks, err error) {

	isFilterFullURL := false
	sliceListFullURL, ok := listFullURL.([]string)
	if ok {
		isFilterFullURL = true
	}

	listAllLinks, err := service.storage.GetShortLinks(ctx)
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

			shortLinks = append(shortLinks, RowShortLink{
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

// Получаем короткую ссылку с хостом по Url-адресу
func (service *ServiceShortLink) GetServiceLinkByURL(ctx context.Context, fullURL string) (serviceLink string, err error) {

	shortLink, err := service.getShortLinkByURL(ctx, fullURL)
	if err == nil {
		if shortLink != "" {
			serviceLink, _ = service.getShortLinkWithHost(shortLink)
		}
	}

	return
}

// Генерируем короткую ссылку по Url-адресу
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
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}
