package logger

import (
	"log"
	"os"
	"path/filepath"
)

// тип нашего логера приложения
type TypeAppLogger struct {
	*log.Logger
}

// переменная логера
var AppLogger TypeAppLogger

func init() {

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parent := filepath.Dir(wd)
	parent = filepath.Dir(parent)

	logDir := parent + "/logs"

	// создаем папку logs в корне проекта
	_, err = os.Stat(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(logDir, 0755)
			if err != nil {
				panic(err)
			}
		} else {
			// другая ошибка
		}
	}

	// путь до файла с логом
	pathLogFile := logDir + "/appLog.log"

	AppLogger = TypeAppLogger{
		createLogger(pathLogFile),
	}

}

func createLogger(pathToFileLog string) *log.Logger {

	f, err := os.OpenFile(pathToFileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	//f, err := os.OpenFile("./logs/info.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//defer f.Close()

	return log.New(f, "Logger:", log.Ldate|log.Ltime)
}
