package storageshortlink

import (
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	restorer "go-url-shortener/internal/storage/storageshortlink/restorer"
	fileRestorer "go-url-shortener/internal/storage/storageshortlink/restorer/filerestorer"
	"log"
)

func getPackageError(textError string) error {
	textModuleError := "storageshortlink: " + textError
	return errors.New(textModuleError)
}

// тип для одной запись с данными о короткой ссылке
type RowStorageShortLink struct {
	ShortLink string
	FullURL   string
	UUID      string
}

// вид хранения ссылок в памяти
type DataStorageShortLink map[string]RowStorageShortLink

// тип для хранилища данных ссылок
type StorageShortInterface interface {
	GetFullLinkByShort(shortLink string) (fullURL string, err error)
	GetShortLinkByURL(fullURL string) (shortLink string, err error)
	AddShortLinkForURL(fullURL, shortLink string) (err error)
	SetData(data DataStorageShortLink) (err error)
	GetCountLink() (count int, err error)
	GetShortLinks() (shortLinks DataStorageShortLink, err error)
	Init() (err error)
}

// Хранилище коротких ссылок в памяти
type StorageShortLink struct {
	Data     DataStorageShortLink
	Restorer restorer.Restorer
}

func NewStorageShorts() StorageShortInterface {
	// путь из конфигурации до файла хранилища с короткими ссылками
	pathFileStorage := config.GetAppConfig().GetFileStoragePath()
	return NewStorageShortsFromFileStorage(pathFileStorage)
}

func NewStorageShortsFromFileStorage(pathFileStorage string) StorageShortInterface {

	data := make(DataStorageShortLink)
	storageRestorer, err := fileRestorer.NewFileRestorer(pathFileStorage)
	if err != nil {
		logger.GetLogger().Error("При инициализации хранилища ссылок возникла ошибка: " + err.Error())
		log.Fatal("При инициализации хранилища ссылок возникла ошибка: " + err.Error())
	}

	storage := &StorageShortLink{
		Data:     data,
		Restorer: storageRestorer,
	}
	storage.Init()
	return storage
}

func (store *StorageShortLink) SetData(data DataStorageShortLink) (err error) {
	store.Data = data
	return
}

func (store *StorageShortLink) GetShortLinks() (shortLinks DataStorageShortLink, err error) {
	return store.Data, nil
}

func (store *StorageShortLink) GetCountLink() (count int, err error) {
	return len(store.Data), nil
}

func (store *StorageShortLink) AddShortLinkForURL(fullURL, shortLink string) (err error) {

	// uuid сделаем просто порядковым номером
	orderLink := len(store.Data) + 1
	uuid := fmt.Sprintf("%d", orderLink)

	store.Data[shortLink] = RowStorageShortLink{
		ShortLink: shortLink,
		FullURL:   fullURL,
		UUID:      uuid,
	}

	// делаем запись в файл с хранилищем данных
	rowDataRestorer := restorer.RowDataRestorer{
		ShortLink: shortLink,
		FullURL:   fullURL,
		UUID:      uuid,
	}
	store.Restorer.WriteRow(rowDataRestorer)

	return
}

func (store *StorageShortLink) GetShortLinkByURL(fullURL string) (shortLink string, err error) {

	for _, rowData := range store.Data {
		if fullURL == rowData.FullURL {
			shortLink = rowData.ShortLink
			break
		}
	}
	return
}

func (store *StorageShortLink) GetFullLinkByShort(shortLink string) (fullURL string, err error) {

	rowData, ok := store.Data[shortLink]
	if !ok {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	} else {
		fullURL = rowData.FullURL
	}
	return
}

func (store *StorageShortLink) SetRestorer(restorer restorer.Restorer) (err error) {
	store.Restorer = restorer
	return
}

func (store *StorageShortLink) clearData() (err error) {
	emptyData := make(DataStorageShortLink)
	store.SetData(emptyData)
	return
}

func (store *StorageShortLink) InitRestorer(restorer restorer.Restorer) (err error) {
	store.SetRestorer(restorer)
	store.Restore()
	return
}

func (store *StorageShortLink) Restore() (err error) {

	store.clearData()

	listRows, err := store.Restorer.ReadAll()
	logger.GetLogger().Debugf("Прочитано коротких ссылок из файла хранилища: %d", len(listRows))

	if err != nil {
		return
	} else {
		dataStorage := DataStorageShortLink{}

		for _, dataRow := range listRows {
			shortLink := dataRow.ShortLink
			dataStorage[shortLink] = RowStorageShortLink{
				ShortLink: shortLink,
				FullURL:   dataRow.FullURL,
				UUID:      dataRow.UUID,
			}
		}

		store.SetData(dataStorage)
	}
	return
}

// Метод вызывающийся при создании объекта
func (store *StorageShortLink) Init() (err error) {
	return store.Restore()
}
