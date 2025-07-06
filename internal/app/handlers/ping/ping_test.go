package ping

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/service/mocks"

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
	m := mocks.NewMockRepository(ctrl)

	// инициализируем service объектом-заглушкой
	service := service.NewService(m)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(logger.Log))
	r.Use(mwgzip.Gzip)

	r.Get("/ping", GetHandler(service, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	tests := []struct {
		error          error
		name           string
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			error:          nil,
			wantStatusCode: 200,
		},
		{
			name:           "negative test #2",
			error:          errors.New("failed connect to database"),
			wantStatusCode: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m.EXPECT().Ping(gomock.Any()).Return(tt.error)

			resp, err := resty.New().R().Get(srv.URL + "/ping")

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())

		})
	}

}
