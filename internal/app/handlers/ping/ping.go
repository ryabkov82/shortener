package ping

import (
	"net/http"

	"go.uber.org/zap"
)

type URLHandler interface {
	Ping() error
}

func GetHandler(urlHandler URLHandler, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// Получаем адрес перенаправления
		err := urlHandler.Ping()
		if err != nil {
			http.Error(res, "Failed connect to database", http.StatusInternalServerError)
			log.Error("Failed connect to database", zap.Error(err))
			return
		}
		log.Debug("Connect to database is successful")

		// устанавливаем код 200
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("Connect to database is successful"))
	}
}
