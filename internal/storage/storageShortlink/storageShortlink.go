package storageShortlink

import "errors"

func getPackageError(textError string) error {
	textModuleError := "storageShortlinks: " + textError
	return errors.New(textModuleError)
}

func NewStorageShorts() StorageShortLink {
	return StorageShortLink{
		Data: make(map[string]string),
	}
}

type StorageShortLink struct {
	Data map[string]string
}

func (store StorageShortLink) AddShortLinkForUrl(fullUrl, shortLink string) (err error) {
	store.Data[shortLink] = fullUrl
	return
}

func (store StorageShortLink) GetShortLinkByUrl(fullUrl string) (shortLink string, err error) {

	for short, urlForShort := range store.Data {
		if fullUrl == urlForShort {
			shortLink = short
			break
		}
	}
	return
}

func (store StorageShortLink) GetFullLinkByShort(shortLink string) (fullUrl string, err error) {

	fullUrl, ok := store.Data[shortLink]
	if !ok {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}
