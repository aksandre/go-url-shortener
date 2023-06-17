package compress

import (
	"compress/gzip"
	"fmt"
	"go-url-shortener/internal/logger"
	"io"
	"net/http"
	"strings"
)

// Это тип будет заменой стандартного http.ResponseWriter
// Он позволит сжимать тело ответа
type compressResponseWriter struct {
	// встраиваем оригинальный http.ResponseWriter
	originWriter http.ResponseWriter
	// Внедряем интерфейс объекта, который умеет работать со сжатием
	gzipWriter io.Writer
}

func (writer *compressResponseWriter) Write(b []byte) (countBytes int, error error) {
	return writer.gzipWriter.Write(b)
	//strLenContent := strconv.Itoa(countBytes)
	//writer.originWriter.Header().Set("Content-Length", strLenContent)
}

func (writer *compressResponseWriter) Header() http.Header {
	return writer.originWriter.Header()
}

func (writer *compressResponseWriter) WriteHeader(statusCode int) {
	writer.originWriter.Header().Del("Content-Length")
	writer.originWriter.Header().Set("Content-Encoding", "gzip")
	writer.originWriter.WriteHeader(statusCode)
}

// Тип при считывании, разжимающий тело запроса
// Должен удовлетворять интерфейсу io.ReadCloser
type decompressReaderCloser struct {
	originReader io.ReadCloser
}

func (reader decompressReaderCloser) Read(p []byte) (n int, err error) {
	// переменная readerZip будет читать входящие данные и распаковывать их
	readerZip, err := gzip.NewReader(reader.originReader)
	if err != nil {
		return 0, fmt.Errorf("ошибка распаковки входных данных: %w", err)
	}
	defer func() {
		readerZip.Close()
	}()
	return readerZip.Read(p)
}

func (reader decompressReaderCloser) Close() (err error) {
	if err = reader.originReader.Close(); err != nil {
		return err
	}
	return
}

// Нужно сжимать данные ответа ?
func isNeedCompressBodyResponse(request *http.Request) bool {

	isNeedCompress := false

	// проверяем, можно ли сжимать ответ
	acceptEncoding := request.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "gzip") {

		contentTypeRequest := request.Header.Get("Content-Type")

		listCompressContentType := []string{
			"application/javascript",
			"application/json",
			"text/css",
			"text/html",
			"text/plain",
			"text/xml",
		}
		for _, nameType := range listCompressContentType {
			if strings.Contains(contentTypeRequest, nameType) {
				isNeedCompress = true
				break
			}
		}

	}

	return isNeedCompress
}

func WrapCompression(handler http.Handler) http.Handler {
	compressFunc := func(respWriter http.ResponseWriter, request *http.Request) {

		// проверяем, пришли ли данные в сжатом виде
		contentEncoding := request.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			request.Body = decompressReaderCloser{
				originReader: request.Body,
			}
		}

		isNeedCompress := isNeedCompressBodyResponse(request)
		if isNeedCompress {

			// создаём объект, который будет реализовывать сжатие
			gz, err := gzip.NewWriterLevel(respWriter, gzip.BestSpeed)
			defer func() {
				gz.Close()
			}()
			if err != nil {
				logger.GetLogger().Error("Ошибка, не получилось сделать сжатие тела ответа: " + err.Error())
			} else {

				// Инициализируем свой аналог http.ResponseWriter,
				// внедряя объект для сжатия gzipWriter
				gzipWriter := &compressResponseWriter{
					originWriter: respWriter,
					gzipWriter:   gz,
				}
				respWriter = gzipWriter
			}
		}

		handler.ServeHTTP(respWriter, request)

	}
	return http.HandlerFunc(compressFunc)
}
