package storagedb

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	errDriver "go-url-shortener/internal/database/errors/pgxerrors"
	"go-url-shortener/internal/logger"
	"strings"

	modelsStorage "go-url-shortener/internal/models/storageshortlink"
)

func getPackageError(textError string) error {
	textModuleError := "storageshortlink: " + textError
	return errors.New(textModuleError)
}

// Хранилище коротких ссылок в БД
type StorageShortLink struct {
	nameTableData string
	dbHandler     *dbconn.DBHandler
}

type StorageShortInterface interface {
	modelsStorage.StorageShortInterface
}

func NewStorageShorts() (storage StorageShortInterface, err error) {

	dbHandler := dbconn.GetDBHandler()
	err = dbHandler.GetErrSetup()
	if err == nil {
		err = dbHandler.Ping()
	}

	if err != nil {
		logger.GetLogger().Error("Не создать хранилище из БД, не возможно к БД подключиться: " + err.Error())
		return nil, err
	}

	// название таблицы в базе данных с короткими ссылками
	nameTableData := config.GetAppConfig().GetNameTableRestorer()
	// проверка существования таблицы
	err = createShortLinkTable(nameTableData)
	if err != nil {
		logger.GetLogger().Error("ошибка создания таблицы для хранения ссылок: " + err.Error())
		return nil, err
	}

	storage = &StorageShortLink{
		nameTableData: nameTableData,
		dbHandler:     dbHandler,
	}
	return storage, nil
}

func (store *StorageShortLink) SetData(ctx context.Context, data modelsStorage.DataStorageShortLink) (err error) {

	oldData, err := store.GetShortLinks(ctx, nil)
	if err != nil {
		return
	}

	err = store.ClearStorage(ctx)
	if err != nil {
		return
	}

	err = store.AddBatchShortLinks(ctx, data)
	if err != nil {
		store.AddBatchShortLinks(ctx, oldData)
		return
	}

	return
}

// добавление коротких ссылок группой
func (store *StorageShortLink) AddBatchShortLinks(ctx context.Context, data modelsStorage.DataStorageShortLink) (err error) {

	// открываем транзакцию
	poolConn := store.dbHandler.GetPool()
	tx, err := poolConn.BeginTx(ctx, nil)
	if err != nil {
		logger.GetLogger().Error("ошибка: не смогли открыть транзакцию: " + err.Error())
		return
	}

	nameTable := store.nameTableData
	for _, row := range data {
		// все изменения записываются в транзакцию
		// игнорируем ошибку дублирующего FULL_URL, чтобы транзакция выполнилась при ее наличии
		sqlAdd := "INSERT INTO " + nameTable + " (SHORT_LINK, FULL_URL) VALUES($1,$2) ON CONFLICT (FULL_URL) DO NOTHING"
		_, err := tx.ExecContext(
			ctx,
			sqlAdd,
			row.ShortLink,
			row.FullURL,
		)
		if err != nil {
			logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlAdd + ": " + err.Error())
			logger.GetLogger().Debugln("отменяем транзакцию")
			errRoll := tx.Rollback()
			if errRoll != nil {
				logger.GetLogger().Error("ошибка: не смогли сделать Rollback транзакции: " + errRoll.Error())
			}
			return err
		}
	}

	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		logger.GetLogger().Error("ошибка: не смогли сделать commit транзакции: " + err.Error())
	}

	return
}

// получаем список данных коротких ссылок по фильтру
func (store *StorageShortLink) GetShortLinks(ctx context.Context, options *modelsStorage.OptionsQuery) (shortLinks modelsStorage.DataStorageShortLink, err error) {

	var shortLinksRows []modelsStorage.RowStorageShortLink

	// если передан фильтр по полным ссылкам
	if options != nil {
		shortLinksRows, err = store.readFilter(ctx, options)
	} else {
		shortLinksRows, err = store.readAll(ctx)
	}

	// колличество результатов
	lenRows := len(shortLinksRows)
	shortLinks = make(modelsStorage.DataStorageShortLink, lenRows)
	for _, row := range shortLinksRows {

		shortLink := row.ShortLink
		shortLinks[shortLink] = modelsStorage.RowStorageShortLink{
			ShortLink: shortLink,
			FullURL:   row.FullURL,
			UUID:      row.UUID,
		}
	}

	return
}

func (store *StorageShortLink) GetCountLink(ctx context.Context) (count int, err error) {

	nameTable := store.nameTableData
	sqlSelectQuery := "SELECT COUNT(*) as COUNT_ROWS FROM " + nameTable

	poolConn := store.dbHandler.GetPool()
	rows, err := poolConn.QueryContext(ctx, sqlSelectQuery)
	// обязательно закрываем чтение строк
	defer func() {
		if err == nil {
			err = rows.Close()
		}
	}()

	if err != nil {
		logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlSelectQuery + ": " + err.Error())
		return
	}

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			logger.GetLogger().Error("ошибка чтения списка всех ссылок из БД: " + err.Error())
			return count, err
		}
	}

	// Проверим ошибки, чтобы понять, что считывание полностью было завершено
	if err := rows.Err(); err != nil {
		logger.GetLogger().Error("чтение строк из таблицы не было завершено корректно, возникла ошибка: " + err.Error())
	}

	return
}

func (store *StorageShortLink) AddShortLinkForURL(ctx context.Context, fullURL, shortLink string) (err error) {

	nameTable := store.nameTableData
	sqlAddRow := "INSERT INTO " + nameTable + " (FULL_URL, SHORT_LINK) VALUES ($1, $2)"
	poolConn := store.dbHandler.GetPool()
	_, err = poolConn.ExecContext(ctx, sqlAddRow, fullURL, shortLink)
	if err != nil {
		isUniqErr, _ := errDriver.IsUniqueViolation(err)
		if isUniqErr {
			err = modelsStorage.NewErrExistFullURLExt(fullURL)
		}
		logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlAddRow + ": " + err.Error())
	}

	return
}

func (store *StorageShortLink) GetShortLinkByURL(ctx context.Context, fullURL string) (shortLink string, err error) {

	nameTable := store.nameTableData
	sqlSelectRow := "SELECT ID, FULL_URL, SHORT_LINK FROM " + nameTable + " WHERE FULL_URL=$1 LIMIT 1"
	allRows, err := store.readRows(ctx, sqlSelectRow, fullURL)
	if err != nil {
		logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlSelectRow + ": " + err.Error())
		return
	}

	if len(allRows) > 0 {
		row := allRows[0]
		return row.ShortLink, nil
	}

	return
}

func (store *StorageShortLink) GetFullLinkByShort(ctx context.Context, shortLink string) (fullURL string, err error) {

	nameTable := store.nameTableData
	sqlSelectRow := "SELECT ID, FULL_URL, SHORT_LINK FROM " + nameTable + " WHERE SHORT_LINK=$1 LIMIT 1"
	allRows, err := store.readRows(ctx, sqlSelectRow, shortLink)
	if err != nil {
		logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlSelectRow + ": " + err.Error())
		return
	}

	if len(allRows) > 0 {
		row := allRows[0]
		return row.FullURL, nil
	} else {
		// должны показать ошибку
		err = getPackageError("Короткая ссылка " + shortLink + " не зарегистрирована")
	}

	return
}

// Метод вызывающийся при создании объекта
func (store *StorageShortLink) Init(ctx context.Context) (err error) {
	return nil
}

// Прочитать строчки в базе по запросу
func (store *StorageShortLink) readRows(ctx context.Context, sqlSelectQuery string, args ...any) (allRows []modelsStorage.RowStorageShortLink, err error) {

	poolConn := store.dbHandler.GetPool()
	rows, err := poolConn.QueryContext(ctx, sqlSelectQuery, args...)
	// обязательно закрываем чтение строк
	defer func() {
		if err == nil {
			err = rows.Close()
		}
	}()

	if err != nil {
		logger.GetLogger().Errorf("ошибка: при выполении запроса " + sqlSelectQuery + ": " + err.Error())
		return
	}

	for rows.Next() {
		var uuid string
		var fullURL string
		var shortLink string
		if err := rows.Scan(&uuid, &fullURL, &shortLink); err != nil {
			logger.GetLogger().Error("ошибка чтения строки из БД хранилища: " + err.Error())
		} else {

			if fullURL != "" && shortLink != "" {
				allRows = append(allRows, modelsStorage.RowStorageShortLink{
					ShortLink: shortLink,
					FullURL:   fullURL,
					UUID:      uuid,
				})
			}
		}
	}

	// Проверим ошибки, чтобы понять, что считывание полностью было завершено
	if err := rows.Err(); err != nil {
		logger.GetLogger().Error("ошибка: чтение строк из таблицы было завершено некорректно, возникла ошибка: " + err.Error())
	}

	return
}

// Прочитать все строки в таблице и вернуть результат в виде слайса
func (store *StorageShortLink) readAll(ctx context.Context) (allRows []modelsStorage.RowStorageShortLink, err error) {
	nameTable := store.nameTableData
	sqlSelectRows := "SELECT ID, FULL_URL, SHORT_LINK FROM " + nameTable + " ORDER BY ID ASC"
	allRows, err = store.readRows(ctx, sqlSelectRows)
	if err != nil {
		logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlSelectRows + ": " + err.Error())
	}
	return
}

// Прочитать все строки в таблице и вернуть результат в виде слайса
func (store *StorageShortLink) readFilter(ctx context.Context, options *modelsStorage.OptionsQuery) (allRows []modelsStorage.RowStorageShortLink, err error) {

	nameTable := store.nameTableData

	if options != nil {
		listFullURL := options.Filter.ListFullURL
		countURLs := len(listFullURL)
		if countURLs > 0 {

			sliceValueAny := make([]string, countURLs)
			for key, url := range listFullURL {
				sliceValueAny[key] = "\"" + url + "\""
			}
			strValueAny := "{" + strings.Join(sliceValueAny, ",") + "}"

			sqlSelectRows := "SELECT ID, FULL_URL, SHORT_LINK FROM " + nameTable + " "
			sqlSelectRows += "WHERE FULL_URL = ANY ($1) ORDER BY ID ASC"

			allRows, err = store.readRows(ctx, sqlSelectRows, strValueAny)
			//logger.GetLogger().Debugf("результат запроса: %+v", allRows)
			if err != nil {
				logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlSelectRows + ": " + err.Error())
			}
		}
	}
	return
}

// Очистить данные хранилища
func (store *StorageShortLink) ClearStorage(ctx context.Context) (err error) {
	tableName := store.nameTableData
	sqlTruncate := "TRUNCATE TABLE " + tableName
	poolConn := store.dbHandler.GetPool()
	_, err = poolConn.ExecContext(ctx, sqlTruncate)
	if err != nil {
		logger.GetLogger().Errorln("ошибка: при выполении запроса " + sqlTruncate + ": " + err.Error())
	}
	return
}

// создаем таблицу для хранения коротких ссылок
func createShortLinkTable(tableName string) (err error) {

	dbHandler := dbconn.GetDBHandler()
	poolConn := dbHandler.GetPool()

	ctx := context.TODO()
	tx, err := poolConn.BeginTx(ctx, nil)
	if err != nil {
		logger.GetLogger().Error("ошибка: не смогли открыть транзакцию: " + err.Error())
		return
	}

	sqlCreateTable := "" +
		"create table IF NOT EXISTS " + tableName + " (" +
		"	ID SERIAL PRIMARY KEY," +
		"	FULL_URL varchar(255)," +
		"	SHORT_LINK varchar(255)," +
		"   CREATED TIMESTAMP DEFAULT CURRENT_TIMESTAMP" +
		") "
	_, err = tx.ExecContext(ctx, sqlCreateTable)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			logger.GetLogger().Error("ошибка: не смогли сделать Rollback транзакции: " + errRoll.Error())
		}

		err = fmt.Errorf("ошибка: не смогли создать таблицу для хранения коротких ссылок: %w", err)
		return
	}

	// делаем индекс для быстрого поиска полной ссылки по короткой ссылке
	sqlCreateIndexShort := "CREATE INDEX IF NOT EXISTS SHORT_LINK_index_" + tableName + " ON " + tableName + " (SHORT_LINK)"
	_, err = tx.ExecContext(ctx, sqlCreateIndexShort)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			logger.GetLogger().Error("ошибка: не смогли сделать Rollback транзакции: " + errRoll.Error())
		}

		err = fmt.Errorf("ошибка: не смогли создать индекс у поля SHORT_LINK: %w", err)
		return
	}

	// делаем уникальный индекс поля FULL_URL как ограничение для целостности данных
	sqlCreateIndexFull := "CREATE UNIQUE INDEX IF NOT EXISTS FULL_URL_index_" + tableName + " ON " + tableName + " (FULL_URL)"
	_, err = tx.ExecContext(ctx, sqlCreateIndexFull)
	if err != nil {
		errRoll := tx.Rollback()
		if errRoll != nil {
			logger.GetLogger().Error("ошибка: не смогли сделать Rollback транзакции: " + errRoll.Error())
		}

		err = fmt.Errorf("ошибка: не смогли создать индекс у поля FULL_URL: %w", err)
		return
	}

	// завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		logger.GetLogger().Error("ошибка: не смогли сделать commit транзакции: " + err.Error())
	}

	return
}
