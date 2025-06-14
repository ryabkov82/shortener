package httpgzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

// оптимизация, пулы
var writerPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

var readerPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

func init() {
	InitPools()
}

func InitPools() {

	// Предзаполнение пула (опционально)
	for i := 0; i < 32; i++ {
		writerPool.Put(writerPool.New())
		readerPool.Put(readerPool.New())
	}
}

func PutWriter(zw *gzip.Writer) {
	writerPool.Put(zw)
}

func PutReader(zr *gzip.Reader) {
	readerPool.Put(zr)
}

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := writerPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	c.w.Header().Del("Content-Length")
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	//if statusCode < 300 {
	c.w.Header().Set("Content-Encoding", "gzip")
	//}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	PutWriter(c.zw) // Возвращаем writer в пул
	return err
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	err := c.r.Close()
	PutReader(c.zr) // Возвращаем reader в пул
	return err
}
