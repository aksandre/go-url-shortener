package dbrestorer

import (
	"context"
	"fmt"
	dbconn "go-url-shortener/internal/database/connect"
	"go-url-shortener/internal/logger"
	"go-url-shortener/internal/storage/storageshortlink/storagerestorer/restorer"
)

// Тип для восстановителя коротких ссылок из базы данных
type DBRestorer struct {
	nameTable string
}

func NewDBRestorer(nameTable string) (restorer *DBRestorer, err error) {

	dbHandler := dbconn.GetDBHandler()
	err = dbHandler.GetErrSetup()
	if err == nil {
		err = dbHandler.Ping()
	}

	if err != nil {
		logger.GetLogger().Error("Не создать восстановитель хранилища из БД, не возможно к БД подключиться: " + err.Error())
		return nil, err
	}

	// проверка существования таблицы
	err = createRestoreTable(nameTable)
	if err != nil {
		logger.GetLogger().Error("ошибка создания таблицы для хранения ссылок: " + err.Error())
		return
	}

	restorer = &DBRestorer{
		nameTable: nameTable,
	}

	return
}

// Записать одну строчку в файл с данными востановления
func (dbRestorer *DBRestorer) WriteRow(dataRow restorer.RowDataRestorer) (err error) {

	tableName := dbRestorer.nameTable
	fullURL := dataRow.FullURL
	shortLink := dataRow.ShortLink

	sqlAddRow := "INSERT INTO " + tableName + " (FULL_URL, SHORT_LINK) VALUES ($1, $2)"

	dbHandler := dbconn.GetDBHandler()
	poolConn := dbHandler.GetPool()
	_, err = poolConn.Exec(sqlAddRow, fullURL, shortLink)
	return
}

// Прочитать строчки в базе по запросу
func (dbRestorer *DBRestorer) readRows(sqlSelectQuery string) (allRows []restorer.RowDataRestorer, err error) {

	dbHandler := dbconn.GetDBHandler()
	poolConn := dbHandler.GetPool()
	rows, err := poolConn.Query(sqlSelectQuery)
	// обязательно закрываем чтение строк
	defer func() {
		if err == nil {
			err = rows.Close()
		}
	}()

	if err != nil {
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
				allRows = append(allRows, restorer.RowDataRestorer{
					ShortLink: shortLink,
					FullURL:   fullURL,
					UUID:      uuid,
				})
			}
		}
	}

	// Проверим ошибки, чтобы понять, что считывание полностью было завершено
	if err := rows.Err(); err != nil {
		logger.GetLogger().Error("чтение строк из таблицы не было завершено корректно, возникла ошщибка: " + err.Error())
	}

	return
}

// Прочитать одну строчку в таблице с данными востановления
func (dbRestorer *DBRestorer) ReadRow() (dataRow restorer.RowDataRestorer, err error) {
	tableName := dbRestorer.nameTable
	sqlSelectRow := "SELECT ID, FULL_URL, SHORT_LINK FROM " + tableName + " ORDER BY ID ASC LIMIT 1"
	allRows, err := dbRestorer.readRows(sqlSelectRow)
	if len(allRows) > 0 {
		dataRow = allRows[0]
	}
	return
}

// Прочитать все строки в таблице с данными востановления и вернуть результат в виде слайса
func (dbRestorer *DBRestorer) ReadAll() (allRows []restorer.RowDataRestorer, err error) {
	tableName := dbRestorer.nameTable
	sqlSelectRows := "SELECT ID, FULL_URL, SHORT_LINK FROM " + tableName + " ORDER BY ID ASC"
	allRows, err = dbRestorer.readRows(sqlSelectRows)
	return
}

// Очистить данные хранилища
func (dbRestorer *DBRestorer) ClearRows() (err error) {
	tableName := dbRestorer.nameTable
	sqlTruncate := "TRUNCATE TABLE " + tableName

	dbHandler := dbconn.GetDBHandler()
	poolConn := dbHandler.GetPool()
	_, err = poolConn.Exec(sqlTruncate)
	return
}

// создаем таблицу для хранения коротких ссылок
func createRestoreTable(tableName string) (err error) {

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

		err = fmt.Errorf("ошибка: не смогли создать таблицу для восстановления коротких ссылок: %w", err)
		return
	}

	sqlCreateIndex := "CREATE INDEX IF NOT EXISTS FULL_URL_index_" + tableName + " ON " + tableName + " (FULL_URL)"
	_, err = tx.ExecContext(ctx, sqlCreateIndex)
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
