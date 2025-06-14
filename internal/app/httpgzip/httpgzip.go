// Пакет httpgzip предоставляет инструменты для сжатия и распаковки HTTP-трафика в формате gzip.
// Реализует пул объектов gzip.Writer и gzip.Reader для оптимизации производительности.
package httpgzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

// Пул gzip.Writer для повторного использования объектов
var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

// Пул gzip.Reader для повторного использования объектов
var readerPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

// InitPools инициализирует пулы объектов предварительным заполнением.
//
// Опционально вызывается при старте приложения для уменьшения
// накладных расходов на создание объектов при первой нагрузке.
func InitPools() {
	for i := 0; i < 32; i++ {
		writerPool.Put(writerPool.New())
		readerPool.Put(readerPool.New())
	}
}

// PutWriter возвращает gzip.Writer в пул для повторного использования.
//
// Параметры:
//   - zw: gzip.Writer для возврата в пул
func PutWriter(zw *gzip.Writer) {
	writerPool.Put(zw)
}

// PutReader возвращает gzip.Reader в пул для повторного использования.
//
// Параметры:
//   - zr: gzip.Reader для возврата в пул
func PutReader(zr *gzip.Reader) {
	readerPool.Put(zr)
}

// compressWriter реализует http.ResponseWriter с поддержкой gzip-сжатия.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter создает новый compressWriter.
//
// Параметры:
//   - w: оригинальный http.ResponseWriter
//
// Возвращает:
//   - *compressWriter: обертку с поддержкой сжатия
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := writerPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

// Header возвращает HTTP-заголовки ответа.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает сжатые данные в ответ.
func (c *compressWriter) Write(p []byte) (int, error) {
	c.w.Header().Del("Content-Length")
	return c.zw.Write(p)
}

// WriteHeader устанавливает код статуса и заголовки ответа.
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

// Close закрывает writer и возвращает его в пул.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	PutWriter(c.zw)
	return err
}

// compressReader реализует io.ReadCloser с поддержкой gzip-распаковки.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает новый compressReader.
//
// Параметры:
//   - r: оригинальный io.ReadCloser
//
// Возвращает:
//   - *compressReader: обертку с поддержкой распаковки
//   - error: ошибка инициализации
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr := readerPool.Get().(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		readerPool.Put(zr)
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает и распаковывает данные.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает reader и возвращает его в пул.
func (c *compressReader) Close() error {
	err := c.r.Close()
	PutReader(c.zr)
	return err
}
