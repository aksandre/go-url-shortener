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

// установка всех данных хранилища
func (store *StorageShortLink) SetData(ctx context.Context, data modelsStorage.DataStorageShortLink) (err error) {

	// запомнили старые данные
	oldData := store.Data

	err = store.ClearStorage(ctx)
	if err != nil {
		return
	}

	err = store.AddBatchShortLinks(ctx, data)
	if err != nil {
		err = store.ClearStorage(ctx)
		if err != nil {
			err = store.AddBatchShortLinks(ctx, oldData)
		}

	}
	return
}

// добавление коротких ссылок группой
func (store *StorageShortLink) AddBatchShortLinks(ctx context.Context, data modelsStorage.DataStorageShortLink) (err error) {

	for _, row := range data {
		shortLink := row.ShortLink
		fullURL := row.FullURL
		err = store.AddShortLinkForURL(ctx, fullURL, shortLink)
		if err != nil {

			// если это ошибка, что мы не можем вставить дубль, то идем дальше
			if errors.Is(err, modelsStorage.ErrExistFullURL) {
				// обнуляем ошибку
				// эта ошибка не должна прерывать дальнейшее выполнение клиентского кода,
				// вызывающего у сервиса метод AddBatchShortLinks
				err = nil
				continue
			}

			err = errors.New("ошибка: групповая установка данных в хранилище была остановлена, произошла ошибка: " + err.Error())
			break
		}
	}
	return
}

// получаем список данных коротких ссылок по фильтру
func (store *StorageShortLink) GetShortLinks(ctx context.Context, options *modelsStorage.OptionsQuery) (shortLinks modelsStorage.DataStorageShortLink, err error) {

	// если передан фильтр по полным ссылкам
	if options != nil {

		shortLinks = modelsStorage.DataStorageShortLink{}
		if len(options.Filter.ListFullURL) > 0 {

			tempData := make(map[string]string, 0)
			for _, fullURL := range options.Filter.ListFullURL {
				tempData[fullURL] = fullURL
			}

			for _, dataRow := range store.Data {
				fullURL := dataRow.FullURL
				if _, ok := tempData[fullURL]; ok {
					shortLink := dataRow.ShortLink
					shortLinks[shortLink] = modelsStorage.RowStorageShortLink{
						ShortLink: shortLink,
						FullURL:   dataRow.FullURL,
						UUID:      dataRow.UUID,
					}
				}
			}
		}

		return shortLinks, nil

	} else {
		return store.Data, nil
	}

}

func (store *StorageShortLink) GetCountLink(ctx context.Context) (count int, err error) {
	return len(store.Data), nil
}

func (store *StorageShortLink) AddShortLinkForURL(ctx context.Context, fullURL, shortLink string) (err error) {

	// uuid сделаем просто порядковым номером
	orderLink := len(store.Data) + 1
	uuid := fmt.Sprintf("%d", orderLink)

	// надо проверить, что fullURL еще не существует в нашем хранилище
	for _, dataRow := range store.Data {
		fullURLRow := dataRow.FullURL
		if fullURLRow == fullURL {
			err = modelsStorage.NewErrExistFullURLExt(fullURL)
			break
		}
	}
	if err != nil {
		return
	}

	// важно
	// если не отследить, что в хранилище уже есть запись с указанной короткой ссылкой,
	// то данные в памяти просто обновятся с существующим ключом, а в рестороре добавится новая запись
	// данные перестанут соответсвовать в памяти и в рестороре
	if _, ok := store.Data[shortLink]; !ok {

		rowDataRestorer := restorer.RowDataRestorer{
			ShortLink: shortLink,
			FullURL:   fullURL,
			UUID:      uuid,
		}

		// делаем запись в ресторер
		err = store.Restorer.WriteRow(rowDataRestorer)
		if err == nil {
			// делаем запись в память
			store.Data[shortLink] = modelsStorage.RowStorageShortLink{
				ShortLink: shortLink,
				FullURL:   fullURL,
				UUID:      uuid,
			}
		}
	}

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

func (store *StorageShortLink) SetMemoryData(ctx context.Context, data modelsStorage.DataStorageShortLink) (err error) {
	store.Data = data
	return
}

// Удаление данных из памяти
// Данные Ресторера не трогаем
func (store *StorageShortLink) clearMemoryData(ctx context.Context) (err error) {
	emptyData := make(modelsStorage.DataStorageShortLink)
	store.SetMemoryData(ctx, emptyData)
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
	logger.GetLogger().Debugf("Прочитано коротких ссылок из Ресторера: %d", len(listRows))

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

		store.SetMemoryData(ctx, dataStorage)
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
	err = store.Restorer.ClearRows()
	if err == nil {
		err = store.clearMemoryData(ctx)
	}

	return err
}
