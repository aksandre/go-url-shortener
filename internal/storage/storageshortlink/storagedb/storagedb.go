package storagedb

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"

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

	// открываем транзакцию
	poolConn := store.dbHandler.GetPool()
	tx, err := poolConn.BeginTx(ctx, nil)
	defer func() {
		// закрываем транзакцию
		if err == nil {
			err = tx.Rollback()
		}
	}()

	nameTable := store.nameTableData
	for _, row := range data {
		// все изменения записываются в транзакцию
		sqlAdd := "INSERT INTO " + nameTable + " (SHORT_LINK, FULL_URL) VALUES($1,$2)"
		_, err := tx.ExecContext(
			ctx,
			sqlAdd,
			row.ShortLink,
			row.FullURL,
		)
		if err != nil {
			logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlAdd + ": " + err.Error())
			logger.GetLogger().Debugln("отменяем транзакцию")
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

func (store *StorageShortLink) GetShortLinks(ctx context.Context) (shortLinks modelsStorage.DataStorageShortLink, err error) {
	shortLinksRow, err := store.readAll(ctx)
	if err != nil {
		return
	}
	lenRows := len(shortLinksRow)
	shortLinks = make(modelsStorage.DataStorageShortLink, lenRows)
	for _, row := range shortLinksRow {

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
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlSelectQuery + ": " + err.Error())
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
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlAddRow + ": " + err.Error())
	}

	return
}

func (store *StorageShortLink) GetShortLinkByURL(ctx context.Context, fullURL string) (shortLink string, err error) {

	nameTable := store.nameTableData
	sqlSelectRow := "SELECT ID, FULL_URL, SHORT_LINK FROM " + nameTable + " WHERE FULL_URL=$1 LIMIT 1"
	allRows, err := store.readRows(ctx, sqlSelectRow, fullURL)
	if err != nil {
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlSelectRow + ": " + err.Error())
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
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlSelectRow + ": " + err.Error())
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
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlSelectQuery + ": " + err.Error())
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
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlSelectRows + ": " + err.Error())
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
		logger.GetLogger().Debugln("ошибка: при выполении запроса " + sqlTruncate + ": " + err.Error())
	}
	return
}

// создаем таблицу для хранения коротких ссылок
func createShortLinkTable(tableName string) (err error) {

	dbHandler := dbconn.GetDBHandler()
	poolConn := dbHandler.GetPool()

	sqlCreateTable := "" +
		"create table IF NOT EXISTS " + tableName + " (" +
		"	ID SERIAL PRIMARY KEY," +
		"	FULL_URL varchar(255)," +
		"	SHORT_LINK varchar(255)," +
		"   CREATED TIMESTAMP DEFAULT CURRENT_TIMESTAMP" +
		") "

	_, err = poolConn.Exec(sqlCreateTable)
	if err != nil {
		err = fmt.Errorf("ошибка: не смогли создать таблицу для хранения коротких ссылок: %w", err)
		return
	}

	return
}
