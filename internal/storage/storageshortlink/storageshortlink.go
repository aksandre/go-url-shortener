package storageshortlink

import (
	"go-url-shortener/internal/logger"
	storageredb "go-url-shortener/internal/storage/storageshortlink/storagedb"
	storagerestorer "go-url-shortener/internal/storage/storageshortlink/storagerestorer"
	"log"

	modelsStorage "go-url-shortener/internal/models/storageshortlink"
)

func NewStorageShorts() modelsStorage.StorageShortInterface {

	storage, err := NewStorageShortsDB()
	if err == nil {
		logger.GetLogger().Debugln("Успешно создали хранилиже в Базе данных")
	} else {
		storage, err = NewStorageShortsRestorer()
		if err == nil {
			logger.GetLogger().Debugln("Успешно создали хранилиже как Ресторер")
		} else {
			log.Fatal("Выход из программы: не удалось инициализировать хранилище ссылок")
		}

	}
	return storage
}

// создание хранилища в памяти с восстановлением из источника
func NewStorageShortsRestorer() (modelsStorage.StorageShortInterface, error) {

	storage, err := storagerestorer.NewStorageShorts()
	if err != nil {
		logger.GetLogger().Error("При инициализации хранилища StorageShortsRestorer возникла ошибка: " + err.Error())
		return nil, err
	}
	return storage, nil
}

// создание хранилища на базе таблицы базы данных
func NewStorageShortsDB() (modelsStorage.StorageShortInterface, error) {

	// путь из конфигурации до файла хранилища с короткими ссылками
	storage, err := storageredb.NewStorageShorts()
	if err != nil {
		logger.GetLogger().Error("При инициализации хранилища ссылок в БД возникла ошибка: " + err.Error())
		return nil, err
	}
	return storage, nil
}
