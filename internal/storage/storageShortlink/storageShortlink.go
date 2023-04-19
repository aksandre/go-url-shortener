package storageshortlink

import "errors"

func getPackageError(textError string) error {
	textModuleError := "storageshortlink: " + textError
	return errors.New(textModuleError)
}

type StorageShortInterface interface {
	GetFullLinkByShort(shortLink string) (fullURL string, err error)
	GetShortLinkByURL(fullURL string) (shortLink string, err error)
	AddShortLinkForURL(fullURL, shortLink string) (err error)
	SetData(data DataStorageShortLink) (err error)
}

// Хранилище коротких ссылок в памяти
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

func (store *StorageShortLink) AddShortLinkForURL(fullURL, shortLink string) (err error) {
	store.Data[shortLink] = fullURL
	return
}

func (store *StorageShortLink) GetShortLinkByURL(fullURL string) (shortLink string, err error) {

	for short, urlForShort := range store.Data {
		if fullURL == urlForShort {
			shortLink = short
			break
		}
	}
	return
}

func (store *StorageShortLink) GetFullLinkByShort(shortLink string) (fullURL string, err error) {

	fullURL, ok := store.Data[shortLink]
	if !ok {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}
	return
}
