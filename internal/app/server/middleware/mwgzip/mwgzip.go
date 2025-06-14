// Пакет mwgzip предоставляет middleware для сжатия и распаковки HTTP-трафика в формате gzip.
package mwgzip

import (
	"net/http"
	"strings"

	"github.com/ryabkov82/shortener/internal/app/httpgzip"
)

// Gzip создает middleware для обработки gzip сжатия HTTP-запросов и ответов.
//
// Middleware выполняет:
// - Сжатие ответов в gzip, если клиент поддерживает прием сжатых данных
// - Распаковку входящих запросов, если они сжаты gzip
// - Прозрачную передачу данных, если gzip не используется
//
// Возвращает:
//
//	func(next http.Handler) http.Handler - middleware функцию
func Gzip(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Используем оригинальный ResponseWriter по умолчанию
		ow := w

		// Проверяем поддержку gzip на стороне клиента
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// Создаем обертку с поддержкой сжатия
			cw := httpgzip.NewCompressWriter(w)
			ow = cw
			// Гарантируем закрытие компрессора
			defer cw.Close()
		}

		// Проверяем сжатие входящего запроса
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// Создаем reader с поддержкой распаковки
			cr, err := httpgzip.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(fn)
}
