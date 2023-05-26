package storagerestorer

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	modelsStorage "go-url-shortener/internal/models/storageshortlink"
	restorer "go-url-shortener/internal/storage/storageshortlink/storagerestorer/restorer"
	dbRestorer "go-url-shortener/internal/storage/storageshortlink/storagerestorer/restorer/dbrestorer"
	fileRestorer "go-url-shortener/internal/storage/storageshortlink/storagerestorer/restorer/filerestorer"
)

func getPackageError(textError string) error {
	textModuleError := "storageshortlink: " + textError
	return errors.New(textModuleError)
}

// тип для хранилища данных ссылок
type StorageShortInterface interface {
	modelsStorage.StorageShortInterface

	GetRestorer() (restorer restorer.Restorer, err error)
}

// Хранилище коротких ссылок в памяти
type StorageShortLink struct {
	Data     modelsStorage.DataStorageShortLink
	Restorer restorer.Restorer
}

func NewStorageShorts() (StorageShortInterface, error) {

	// название таблицы в базе данных с короткими ссылками
	nameTableRestorer := config.GetAppConfig().GetNameTableRestorer()
	storage, err := NewStorageShortsFromDB(nameTableRestorer)
	if err != nil {
		// путь из конфигурации до файла хранилища с короткими ссылками
		pathFileStorage := config.GetAppConfig().GetFileStoragePath()
		storage, err = NewStorageShortsFromFileStorage(pathFileStorage)
		if err != nil {
			return nil, err
		}
	}
	return storage, nil
}

// создание хранилища на базе файла
func NewStorageShortsFromFileStorage(pathFileStorage string) (StorageShortInterface, error) {

	data := make(modelsStorage.DataStorageShortLink)
	storageRestorer, err := fileRestorer.NewFileRestorer(pathFileStorage)
	if err != nil {
		logger.GetLogger().Error("При инициализации хранилища ссылок в Файле возникла ошибка: " + err.Error())
		return nil, err
	}

	storage := &StorageShortLink{
		Data:     data,
		Restorer: storageRestorer,
	}
	storage.Init(context.TODO())
	return storage, nil
}

// создание хранилища на базе таблицы базы данных
func NewStorageShortsFromDB(nameTable string) (StorageShortInterface, error) {

	logger.GetLogger().Debug("Используется таблица для хранения коротких ссылок: " + nameTable)

	data := make(modelsStorage.DataStorageShortLink)
	// путь из конфигурации до файла хранилища с короткими ссылками
	storageRestorer, err := dbRestorer.NewDBRestorer(nameTable)
	if err != nil {
		logger.GetLogger().Error("При инициализации хранилища ссылок в БД возникла ошибка: " + err.Error())
		return nil, err
	}

	storage := &StorageShortLink{
		Data:     data,
		Restorer: storageRestorer,
	}
	storage.Init(context.TODO())
	return storage, nil
}

func (store *StorageShortLink) SetData(ctx context.Context, data modelsStorage.DataStorageShortLink) (err error) {
	store.Data = data
	return
}

func (store *StorageShortLink) GetShortLinks(ctx context.Context) (shortLinks modelsStorage.DataStorageShortLink, err error) {
	return store.Data, nil
}

func (store *StorageShortLink) GetCountLink(ctx context.Context) (count int, err error) {
	return len(store.Data), nil
}

func (store *StorageShortLink) AddShortLinkForURL(ctx context.Context, fullURL, shortLink string) (err error) {

	// uuid сделаем просто порядковым номером
	orderLink := len(store.Data) + 1
	uuid := fmt.Sprintf("%d", orderLink)

	store.Data[shortLink] = modelsStorage.RowStorageShortLink{
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

func (store *StorageShortLink) GetShortLinkByURL(ctx context.Context, fullURL string) (shortLink string, err error) {

	for _, rowData := range store.Data {
		if fullURL == rowData.FullURL {
			shortLink = rowData.ShortLink
			break
		}
	}
	return
}

func (store *StorageShortLink) GetFullLinkByShort(ctx context.Context, shortLink string) (fullURL string, err error) {

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

// Удаление данных из памяти
// Данные Ресторера не трогаем
func (store *StorageShortLink) clearMemoryData(ctx context.Context) (err error) {
	emptyData := make(modelsStorage.DataStorageShortLink)
	store.SetData(ctx, emptyData)
	return
}

func (store *StorageShortLink) InitRestorer(ctx context.Context, restorer restorer.Restorer) (err error) {
	store.SetRestorer(restorer)
	store.Restore(ctx)
	return
}

func (store *StorageShortLink) Restore(ctx context.Context) (err error) {

	store.clearMemoryData(ctx)

	listRows, err := store.Restorer.ReadAll()
	logger.GetLogger().Debugf("Прочитано коротких ссылок из файла хранилища: %d", len(listRows))

	if err != nil {
		return
	} else {
		dataStorage := modelsStorage.DataStorageShortLink{}

		for _, dataRow := range listRows {
			shortLink := dataRow.ShortLink
			dataStorage[shortLink] = modelsStorage.RowStorageShortLink{
				ShortLink: shortLink,
				FullURL:   dataRow.FullURL,
				UUID:      dataRow.UUID,
			}
		}

		store.SetData(ctx, dataStorage)
	}
	return
}

func (store *StorageShortLink) GetRestorer() (restorer restorer.Restorer, err error) {
	restorer = store.Restorer
	return
}

// Метод вызывающийся при создании объекта
func (store *StorageShortLink) Init(ctx context.Context) (err error) {
	return store.Restore(ctx)
}

// Удаляем данные хранилища
func (store *StorageShortLink) ClearStorage(ctx context.Context) (err error) {
	err = store.clearMemoryData(ctx)
	if err == nil {
		err = store.Restorer.ClearRows()
	}

	return err
}
