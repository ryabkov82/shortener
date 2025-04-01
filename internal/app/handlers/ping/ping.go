package ping

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type URLHandler interface {
	Ping(context.Context) error
}

func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		err := urlHandler.Ping(req.Context())
		if err != nil {
			http.Error(res, "Failed to connect to database", http.StatusInternalServerError)
			log.Error("Failed to connect to database", zap.Error(err))
			return
		}
		log.Debug("Connect to database is successful")

		// устанавливаем код 200
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Connect to database is successful"))
	}
}
