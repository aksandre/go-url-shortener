package logger

import (
	"fmt"
	"go-url-shortener/internal/config"
	"log"
	"os"
)

// тип нашего логера приложения
type TypeAppLogger struct {
	*log.Logger
}

// переменная логера
var appLogger TypeAppLogger

// Маркер синглтона, что сущность, уже инициировали
var setupLogger = false

func GetLogger() TypeAppLogger {
	if setupLogger == false {
		initLogger()
		setupLogger = true
	}
	return appLogger
}

// инициализация сущности
func initLogger() {
	mainFolderLog := config.GetAppConfig().GetLogsPath()
	logAppDir, err := getFolderLogs(mainFolderLog)

	// путь до файла с логом
	pathLogFile := "/appLog.log"
	if err != nil {
		fmt.Println(err)
	} else {
		pathLogFile = logAppDir + "/appLog.log"
	}

	loggerBase := createBaseLogger(pathLogFile)
	appLogger = TypeAppLogger{
		loggerBase,
	}
}

// Получение пути до папки с логами
func getFolderLogs(mainFolderLog string) (logAppDir string, err error) {
	logGoDir := mainFolderLog + "/goLogs"
	err = createFolder(logGoDir)
	if err != nil {
		return
	}

	logAppDir = logGoDir + "/urlShortener"
	err = createFolder(logAppDir)
	return
}

// Созадние папки по указанному пути
func createFolder(folderPath string) error {
	// создаем папку logs в корне проекта
	_, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(folderPath, 0755)
			if err != nil {
				panic(err)
			}
		} else {
			// другая ошибка
		}
	}
	return err
}

// Создание объекта Логера из стандартной библиотеки
func createBaseLogger(pathToFileLog string) *log.Logger {

	f, err := os.OpenFile(pathToFileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//defer f.Close()
	return log.New(f, "Logger:", log.Ldate|log.Ltime)
}
