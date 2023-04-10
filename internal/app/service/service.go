package service

import (
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

func NewServiceShortLink(storage *storageShort.StorageShortLink, lengthShortLink int) *ServiceShortLink {

	if lengthShortLink == 0 {
		lengthShortLink = 8
	}

	return &ServiceShortLink{
		storage:         storage,
		lengthShortLink: lengthShortLink,
	}
}

type ServiceShortLink struct {
	storage         *storageShort.StorageShortLink
	lengthShortLink int
}

func (service *ServiceShortLink) SetLength(length int) {
	service.lengthShortLink = length
}

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

func (service *ServiceShortLink) GetServiceLinkByUrl(fullUrl, hostService string) (serviceLink string, err error) {

	shortLink, err := service.getShortLinkByUrl(fullUrl)
	if err == nil {
		if shortLink != "" {
			serviceLink = hostService + "/" + shortLink
			return
		}
	}

	return
}

func (service *ServiceShortLink) getShortLinkByUrl(fullUrl string) (shortLink string, err error) {

	shortLink, err = service.storage.GetShortLinkByUrl(fullUrl)
	if err == nil {
		if shortLink != "" {
			return
		} else {

			//такой url еще не приходил, генерируем новую ссылку
			lengthShort := service.lengthShortLink
			shortLink = service.getRandString(lengthShort)
			logger.AppLogger.Printf("Сформировали новый код  %s", shortLink)

			// добавим короткую ссылку в хранилище
			err = service.storage.AddShortLinkForUrl(fullUrl, shortLink)
			logger.AppLogger.Printf("Содержание storage %+v", service.storage)
		}
	}

	return
}

func (service *ServiceShortLink) GetFullLinkByShort(shortLink string) (fullUrl string, err error) {

	fullUrl, err = service.storage.GetFullLinkByShort(shortLink)
	if err != nil {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}
