package batch

import (
	"net/http"

	"go.uber.org/zap"
)

type URLHandler interface {
	Batch() error
}

func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

	}
}
