package storageShortlink

import "errors"

func getPackageError(textError string) error {
	textModuleError := "storageShortlinks: " + textError
	return errors.New(textModuleError)
}

type StorageShortInterface interface {
	GetFullLinkByShort(shortLink string) (fullUrl string, err error)
	GetShortLinkByUrl(fullUrl string) (shortLink string, err error)
	AddShortLinkForUrl(fullUrl, shortLink string) (err error)
	SetData(data DataStorageShortLink) (err error)
}

// Хранилище коротких ссылок
type DataStorageShortLink map[string]string

type StorageShortLink struct {
	Data DataStorageShortLink
}

func NewStorageShorts() StorageShortInterface {
	return &StorageShortLink{
		Data: make(DataStorageShortLink),
	}
}

func (store *StorageShortLink) SetData(data DataStorageShortLink) (err error) {
	store.Data = data
	return
}

func (store *StorageShortLink) AddShortLinkForUrl(fullUrl, shortLink string) (err error) {
	store.Data[shortLink] = fullUrl
	return
}

func (store *StorageShortLink) GetShortLinkByUrl(fullUrl string) (shortLink string, err error) {

	for short, urlForShort := range store.Data {
		if fullUrl == urlForShort {
			shortLink = short
			break
		}
	}
	return
}

func (store *StorageShortLink) GetFullLinkByShort(shortLink string) (fullUrl string, err error) {

	fullUrl, ok := store.Data[shortLink]
	if !ok {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}
