package logger

import (
	"go-url-shortener/internal/config"

	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// тип нашего логера приложения
type TypeAppLogger struct {
	*log.Logger
}

// своя реализация полей и методов внешнего пакета logrus,
// чтобы не разбрасывать по коду его вызовы
type CustomFields log.Fields

func (logger TypeAppLogger) WithFields(additinalFields CustomFields) *log.Entry {
	logrusFields := log.Fields(additinalFields)
	return logger.Logger.WithFields(logrusFields)
}

// переменная логера
var appLogger TypeAppLogger

// Маркер синглтона, что сущность, уже инициировали
var setupLogger = false

func GetLogger() TypeAppLogger {
	if !setupLogger {
		initLogger()
		setupLogger = true
	}
	return appLogger
}

// инициализация сущности
func initLogger() {

	// создаем файл логов
	pathLogFile, errorCreateLogFile := createLogFile()
	// fmt.Println("ЛОГИ ТУТ: " + pathLogFile)

	loggerBase := createBaseLogger(pathLogFile)
	appLogger = TypeAppLogger{
		loggerBase,
	}

	levelLogConfig := config.GetAppConfig().GetLevelLogs()
	SetLevelLog(levelLogConfig)
	appLogger.Debugf("Уровень логирования: %d", levelLogConfig)

	if errorCreateLogFile != nil {
		appLogger.Errorln("Произошла ошибка создания файла лога: " + errorCreateLogFile.Error())
	}

}

// Создание файла лога
func createLogFile() (pathLogFile string, errorCreateLogFile error) {

	// по тз нужно, чтобы  путь до логов получался через FILE_STORAGE_PATH
	pathLogFileConfig := config.GetAppConfig().GetFileStoragePath()
	if pathLogFileConfig != "" {

		nameFile := filepath.Base(pathLogFileConfig)
		// путь с правильными разделителями операционной системы
		pathToFile := filepath.Dir(pathLogFileConfig)

		err := os.MkdirAll(pathToFile, 0755)

		if err != nil && !os.IsExist(err) {
			errorCreateLogFile = err
		} else {
			// записываем путь с разделителями как в операционной системе
			separatorOS := string(filepath.Separator)
			pathLogFile = pathToFile + separatorOS + nameFile
		}

	} else {
		// раньше для пути сохранения логов была использована переменная LOGS_PATH_GOLANG
		// оставляем для обратной совместимости

		// получаем папку для логов
		mainFolderLog := config.GetAppConfig().GetLogsPath()
		logAppDir, err := getFolderLogs(mainFolderLog)
		if err != nil {
			errorCreateLogFile = err
		} else {
			separatorOS := string(filepath.Separator)
			pathLogFile = logAppDir + separatorOS + "appLog.log"
		}
	}

	// если не смогли создать файл с логом,
	// то пишем в домашнюю директорию в аварийный файл
	if errorCreateLogFile != nil {
		separatorOS := string(filepath.Separator)
		userHomePath := config.GetAppConfig().GetFileStoragePath()
		pathLogFile = userHomePath + separatorOS + "go_app_crash_log.log"
	}

	return
}

// Создание объекта Логера из стандартной библиотеки
func createBaseLogger(pathToFileLog string) *log.Logger {

	file, err := os.OpenFile(pathToFileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//defer file.Close()

	logger := log.New()

	// устанавливаем вывод логов в файл
	logger.SetOutput(file)

	// устанавливаем вывод логов в формате JSON
	logger.SetFormatter(&log.TextFormatter{})

	return logger
}

// Устанавливаем уровень логирования
func SetLevelLog(level int) {
	levelLog := log.Level(level)
	appLogger.SetLevel(levelLog)
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

// Создание папки по указанному пути
func createFolder(folderPath string) error {
	// создаем папку logs в корне проекта
	_, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(folderPath, 0755)
			if err != nil {
				panic(err)
			}
		}
	}
	return err
}
