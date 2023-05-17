package connect

import (
	"context"
	"errors"
	"fmt"
	"go-url-shortener/internal/config"
	"go-url-shortener/internal/logger"
	"time"

	pgx "github.com/jackc/pgx/v5"
)

type connectionDb struct {
	conn *pgx.Conn
}

func (connDb *connectionDb) Close() {
	if connDb.conn != nil {
		connDb.conn.Close(context.Background())
	}
}

// инициализация сущности
func (connDb *connectionDb) initConnect(databaseDsn string) {

	// закроем старое соединение
	connDb.Close()

	if databaseDsn == "" {
		// получаем источник подключения к БД из конфига
		databaseDsn = config.GetAppConfig().GetDatabaseDsn()
	}

	// databaseDsn = "postgres://postgres:123456789@localhost:5432/test_psg"
	// databaseDsn = "user=postgres password=123456789 host=localhost port=5432 dbname=test_psg"
	// config, err := pgx.ParseConfig(databaseDsn)
	connect, err := pgx.Connect(context.Background(), databaseDsn)
	if err != nil {
		err = fmt.Errorf("ошибка: невозможно подключиться к базе данных: %w", err)
		strError := err.Error()
		logger.GetLogger().Errorf("%s", strError)
	} else {
		connDb.conn = connect
	}
}

// инициализация сущности
func (connDb *connectionDb) Ping() (err error) {

	if connDb.conn != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		logger.GetLogger().Debugf("Соединение %+v", connDb.conn)
		err = connDb.conn.Ping(ctx)
	} else {
		err = errors.New("ошибка: не было установлено соединение с базой данных")
	}

	return
}

// переменная поключения к БД
var connection connectionDb

// Маркер синглтона, что сущность, уже инициировали
var setupConnection = false

// метод получения соединения с БД
func GetConnect() *connectionDb {
	if !setupConnection {
		connection.initConnect("")
		setupConnection = true
	}
	return &connection
}

// публичный метод установки соединения с БД
func SetConnect(conn connectionDb) {
	connection = conn
}
