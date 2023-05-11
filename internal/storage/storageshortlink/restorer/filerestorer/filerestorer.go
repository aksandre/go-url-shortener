package filerestorer

import (
	"bufio"
	"encoding/json"
	"go-url-shortener/internal/logger"
	restorer "go-url-shortener/internal/storage/storageshortlink/restorer"
	"os"
	"path/filepath"
)

type FileRestorer struct {
	file   *os.File
	writer *bufio.Writer
	reader *bufio.Scanner
}

func NewFileRestorer(pathFile string) (*FileRestorer, error) {

	pathFile, err := createRestoreFile(pathFile)
	if err != nil {
		logger.GetLogger().Error("ошибка создания файла хранилища ссылок: " + err.Error())
		return nil, err
	}

	// открываем файл для записи в конец
	file, err := os.OpenFile(pathFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}

	restorer := &FileRestorer{
		file:   file,
		writer: bufio.NewWriter(file),
		reader: bufio.NewScanner(file),
	}

	return restorer, nil
}

// Записать одну строчку в файл с данными востановления
func (fileRestorer *FileRestorer) WriteRow(dataRow restorer.RowDataRestorer) (err error) {

	dataBytes, err := json.Marshal(dataRow)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := fileRestorer.writer.Write(dataBytes); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := fileRestorer.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return fileRestorer.writer.Flush()

}

// Прочитать одну строчку в файле с данными востановления
func (fileRestorer *FileRestorer) ReadRow() (dataRow restorer.RowDataRestorer, isLastRow bool, err error) {

	// одиночное сканирование до следующей строки
	if !fileRestorer.reader.Scan() {
		errScan := fileRestorer.reader.Err()
		isLastRow = (errScan == nil)
		return dataRow, isLastRow, errScan
	}

	// читаем данные из scanner
	dataBytes := fileRestorer.reader.Bytes()
	err = json.Unmarshal(dataBytes, &dataRow)
	if err != nil {
		return dataRow, isLastRow, err
	}

	return dataRow, isLastRow, nil
}

// Прочитать весь файл с данными востановления и вернуть результат в виде слайса
func (fileRestorer *FileRestorer) ReadAll() (allRows []restorer.RowDataRestorer, err error) {

	for {
		dataRow, isLastRow, err := fileRestorer.ReadRow()
		if err != nil {
			logger.GetLogger().Error("ошибка чтения строки из файла хранилища: " + err.Error())
		} else {
			if dataRow.ShortLink != "" && dataRow.FullURL != "" {
				allRows = append(allRows, dataRow)
			}
		}

		if isLastRow {
			break
		}
	}

	return
}

func createRestoreFile(pathFile string) (pathRestoreFile string, err error) {

	nameFile := filepath.Base(pathFile)
	// путь с правильными разделителями операционной системы
	pathToFile := filepath.Dir(pathFile)

	// записываем путь с разделителями как в операционной системе
	separatorOS := string(filepath.Separator)
	pathRestoreFile = pathToFile + separatorOS + nameFile

	// создаем папку logs в корне проекта
	_, err = os.Stat(pathToFile)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(pathToFile, 0755)
			if err != nil && !os.IsExist(err) {
				return
			}
		}
	}
	return
}