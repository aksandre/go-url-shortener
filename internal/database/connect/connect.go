package connect

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBHandler struct {
	poolConn *sql.DB
	// Флаг, что подключение к БД было успешно
	isSuccessSetup bool
	// Ошибка установки объекта для работы с БД
	errSetup error
	// флаг, что соединение закрыли
	isClosed bool
}

// соединение готово к работе
func (dbHandler *DBHandler) isReady() bool {
	return (dbHandler.isSuccessSetup && !dbHandler.isClosed)
}

func (dbHandler *DBHandler) GetErrSetup() (err error) {
	if !dbHandler.isSuccessSetup {
		return dbHandler.errSetup
	}
	return nil
}

func (dbHandler *DBHandler) GetPool() (db *sql.DB) {
	return dbHandler.poolConn
}

func (dbHandler *DBHandler) Close() (err error) {

	if dbHandler.isReady() {
		// закрываем соединения с БД
		err = dbHandler.poolConn.Close()
		if err != nil {
			logger.GetLogger().Debug("Не удалось закрыть соединение с БД: " + err.Error())
		} else {
			// выставляем флаг, что закрыли соединение
			dbHandler.isClosed = true
			logger.GetLogger().Debug("Закрыли соединение с БД")
		}
	}
	return
}

// установка соединения с БД
func (dbHandler *DBHandler) initDB(databaseDsn string) (err error) {

	//databaseDsn = "postgres://postgres:123456789@localhost:5432/test_psg?sslmode=disable"
	//databaseDsn = "user=postgres password=123456789 host=localhost port=5432 dbname=test_psg"
	logger.GetLogger().Debug("Используемый databaseDsn :" + databaseDsn)

	// config, err := pgx.ParseConfig(databaseDsn)
	// connect, err := pgx.Connect(context.Background(), databaseDsn)
	dbValue, err := sql.Open("pgx", databaseDsn)
	if err != nil {
		err = fmt.Errorf("ошибка: невозможно подключиться к базе данных по переданных доступам: %w", err)
		strError := err.Error()
		logger.GetLogger().Errorf("%s", strError)
	} else {
		dbHandler.poolConn = dbValue
		logger.GetLogger().Debug("Открыли соединение с БД")
	}

	return err
}

// инициализация сущности
func (dbHandler *DBHandler) setup(databaseDsn string) (err error) {

	err = dbHandler.initDB(databaseDsn)
	if err != nil {
		dbHandler.errSetup = err
	} else {
		err = dbHandler.Ping()
		if err != nil {
			dbHandler.errSetup = err
		} else {
			dbHandler.isSuccessSetup = true
		}
	}

	return
}

// проверка подключения к БД в течении нкоторого времени
func (dbHandler *DBHandler) Ping() (err error) {

	if dbHandler.poolConn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		//logger.GetLogger().Debugf("Соединение с базой данных: %+v", dbHandler.poolConn)
		err = dbHandler.poolConn.PingContext(ctx)
	} else {
		err = errors.New("ошибка: не было установлено соединение с базой данных")
	}

	return
}

// переменная поключения к БД
var dbHandler = &DBHandler{}

// метод получения соединения с БД
func GetDBHandler() *DBHandler {
	if !dbHandler.isReady() {
		dbHandler = NewDBHandler("")
	}
	return dbHandler
}

// публичный метод установки соединения с БД
func SetDBHandler(dbValue *DBHandler) {
	dbHandler = dbValue
}

// Конструктор обработчика соединения с БД
// Надо создавать объект через него
func NewDBHandler(databaseDsn string) (dbHandler *DBHandler) {

	if databaseDsn == "" {
		configDatabaseDsn := config.GetAppConfig().GetDatabaseDsn()
		databaseDsn = configDatabaseDsn
	}

	dbHandler = &DBHandler{}
	dbHandler.setup(databaseDsn)
	return
}
