package shortenapi

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/service"
	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler(t *testing.T) {

	storage := storage.NewInMemoryStorage()

	service := service.NewService(storage)

	if err := logger.Initialize("debug"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	baseURL := "http://localhost:8080/"
	r.Post("/", GetHandler(service, baseURL, logger.Log))

	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(r)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	tests := []struct {
		name           string
		request        Request
		wantStatusCode int
	}{
		{
			name:           "positive test #1",
			request:        Request{URL: "https://practicum.yandex.ru/"},
			wantStatusCode: 201,
		},
		{
			name:           "negative test #2",
			request:        Request{URL: "not url"},
			wantStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := json.Marshal(tt.request)
			if err != nil {
				panic(err)
			}
			resp, err := resty.New().R().
				SetBody(req).
				Post(srv.URL)

			assert.NoError(t, err)

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			if tt.wantStatusCode == 201 {
				var response Response
				err = json.Unmarshal(resp.Body(), &response)
				assert.NoError(t, err)
				shortURL := response.Result
				// Проверяем, что получен URL
				_, err = url.Parse(shortURL)
				assert.NoError(t, err)
			}
		})
	}
}
