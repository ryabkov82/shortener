package ping

import (
	"net/http/httptest"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/handlers/ping/mocks"

	"github.com/ryabkov82/shortener/internal/app/logger"

	mwlogger "github.com/ryabkov82/shortener/internal/app/server/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/middleware/mwgzip"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockURLHandler(ctrl)

	m.EXPECT().Ping().Return(nil)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)

	r.Get("/ping", GetHandler(m, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	resp, err := resty.New().R().
		Get(srv.URL + "/ping")

	assert.NoError(t, err)

	// Проверяем статус ответа
	assert.Equal(t, 200, resp.StatusCode())

}
