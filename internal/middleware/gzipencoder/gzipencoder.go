package gzipencoder

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// middleware для сжатия и разжатия данных.
func GZIPEncodingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Проверяем, что клиент отправил серверу сжатые данные в формате gzip
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				zr, err := gzip.NewReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// меняем тело запроса на новое
				r.Body = zr
				defer zr.Close()
			}

			// проверяем, что клиент поддерживает gzip-сжатие
			// это упрощённый пример. В реальном приложении следует проверять все
			// значения r.Header.Values("Accept-Encoding") и разбирать строку
			// на составные части, чтобы избежать неожиданных результатов
			actualW := w
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {

				// создаём gzip.Writer поверх текущего w
				gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					io.WriteString(w, err.Error())
					return
				}
				defer gz.Close()

				w.Header().Set("Content-Encoding", "gzip")
				actualW = gzipWriter{ResponseWriter: w, Writer: gz}
			}

			next.ServeHTTP(actualW, r)
		})
	}
}

// Обертка для ResponseWriter
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	return w.Writer.Write(b)
}
