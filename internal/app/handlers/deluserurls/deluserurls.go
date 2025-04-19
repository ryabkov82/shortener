package deluserurls

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type URLHandler interface {
	DeleteUserUrls(context.Context, []string) error
}

func GetHandler(urlHandler URLHandler, baseURL string, log *zap.Logger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		var shortURLs []string
		if err := json.NewDecoder(req.Body).Decode(&shortURLs); err != nil {
			http.Error(res, "Invalid request body", http.StatusBadRequest)
			return
		}

		err := urlHandler.DeleteUserUrls(req.Context(), shortURLs)

		if err != nil {
			http.Error(res, "Failed to delete user urls", http.StatusBadRequest)
			log.Error("Failed to delete user urls", zap.Error(err))
			return
		}

		res.WriteHeader(http.StatusAccepted)

	}
}
