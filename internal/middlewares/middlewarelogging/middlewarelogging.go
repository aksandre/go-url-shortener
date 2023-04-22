package middlewarelogging

import (
	"go-url-shortener/internal/logger"
	"net/http"
	"time"
)

// Сделаем структуру, где будем храненить сведения об ответе
type responseData struct {
	status int
	size   int
}

// Это будет наша замена стандартного http.ResponseWriter
type loggingResponseWriter struct {
	// встраиваем оригинальный http.ResponseWriter
	http.ResponseWriter
	// Внедряем хранилище с данными запроса
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	// сохраняем в хранилище размер
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	// сохраняем в хранилище код статуса
	r.responseData.status = statusCode
}

func WpapLogging(handler http.Handler) http.Handler {
	logFunc := func(respWriter http.ResponseWriter, request *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		// создаем свой собственный ResponseWriter
		// используя оригинальный ResponseWriter
		logWriter := loggingResponseWriter{
			ResponseWriter: respWriter,
			responseData:   responseData,
		}

		// внедряем оригинальную реализацию http.ResponseWriter
		handler.ServeHTTP(&logWriter, request)

		duration := time.Since(start)

		obLogger := logger.GetLogger()
		// определяем стандартные поля JSON
		additinalFields := logger.CustomFields{
			"uri":      request.RequestURI,
			"method":   request.Method,
			"duration": duration,
			// получаем перехваченный размер ответа
			"sizeResponse": responseData.size,
			// получаем перехваченный код статуса ответа
			"statusCodeResponse": responseData.status,
		}
		textLog := "!!! Зарегистрирован запрос"
		obLogger.WithFields(additinalFields).Info(textLog)
	}
	return http.HandlerFunc(logFunc)
}
