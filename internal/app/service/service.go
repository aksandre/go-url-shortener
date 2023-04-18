package service

import (
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	storageShort "go-url-shortener/internal/storage/storageShortlink"

	"errors"
	"math/rand"
	"strings"
)

func getPackageError(textError string) error {
	textModuleError := "service: " + textError
	return errors.New(textModuleError)
}

// создание сервис коротких ссылок
func NewServiceShortLink(storage storageShort.StorageShortInterface, configApp config.ConfigTypeInterface) ServiceShortInterface {
	return &ServiceShortLink{
		configApp:       configApp,
		storage:         storage,
		lengthShortLink: 8,
	}
}

type ServiceShortInterface interface {
	GetServiceLinkByUrl(fullUrl string) (serviceLink string, err error)
	GetFullLinkByShort(shortLink string) (fullUrl string, err error)
	getHostShortLink() string
	SetLength(length int)
}

type ServiceShortLink struct {
	storage         storageShort.StorageShortInterface
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
func (service ServiceShortLink) getHostShortLink() string {
	host := service.configApp.GetHostShortLink()
	return host
}

// Получаем короткую ссылку с хостом по Url-адресу
func (service *ServiceShortLink) GetServiceLinkByUrl(fullUrl string) (serviceLink string, err error) {

	shortLink, err := service.getShortLinkByUrl(fullUrl)
	if err == nil {
		if shortLink != "" {
			hostService := service.getHostShortLink()
			serviceLink = hostService + "/" + shortLink
			return
		}
	}

	return
}

// Генерируем короткую ссылку по Url-адресу
func (service *ServiceShortLink) getShortLinkByUrl(fullUrl string) (shortLink string, err error) {

	shortLink, err = service.storage.GetShortLinkByUrl(fullUrl)
	if err == nil {
		if shortLink != "" {
			return
		} else {

			//такой url еще не приходил, генерируем новую ссылку
			lengthShort := service.lengthShortLink
			shortLink = service.getRandString(lengthShort)
			logger.GetLogger().Printf("Сформировали новый код  %s", shortLink)

			// добавим короткую ссылку в хранилище
			err = service.storage.AddShortLinkForUrl(fullUrl, shortLink)
			logger.GetLogger().Printf("Содержание storage %+v", service.storage)
		}
	}

	return
}

// Получаем Url-адрес по короткой ссылке
func (service *ServiceShortLink) GetFullLinkByShort(shortLink string) (fullUrl string, err error) {

	fullUrl, err = service.storage.GetFullLinkByShort(shortLink)
	if err != nil {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}
