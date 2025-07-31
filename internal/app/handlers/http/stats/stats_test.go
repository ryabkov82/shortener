package stats_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/ryabkov82/shortener/internal/app/config"
	"github.com/ryabkov82/shortener/internal/app/handlers/http/stats"
	"github.com/ryabkov82/shortener/internal/app/logger"
	"github.com/ryabkov82/shortener/internal/app/models"
	mwlogger "github.com/ryabkov82/shortener/internal/app/server/http/middleware/logger"
	"github.com/ryabkov82/shortener/internal/app/server/http/middleware/trustednet"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/internal/app/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetHandler_WithTrustedSubnet(t *testing.T) {
	// Инициализация логгера
	if err := logger.Initialize("debug"); err != nil {
		t.Fatalf("logger initialization failed: %v", err)
	}

	// Конфигурация с доверенной подсетью
	cfg := &config.Config{
		TrustedSubnet: "192.168.1.0/24",
	}

	// Создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаём мок репозитория
	mockRepo := mocks.NewMockRepository(ctrl)
	service := service.NewService(mockRepo)

	// Инициализируем роутер с middleware
	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(zap.L()))
	r.Use(trustednet.CheckTrustedSubnet(cfg.TrustedSubnet))
	r.Get("/api/internal/stats", stats.GetHandler(service, zap.L()))

	// Запускаем тестовый сервер
	srv := httptest.NewServer(r)
	defer srv.Close()

	tests := []struct {
		name           string
		headers        map[string]string
		urlCount       int
		userCount      int
		urlError       error
		userError      error
		wantStatusCode int
	}{
		{
			name: "positive test #1 - valid stats",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.100",
			},
			urlCount:       10,
			userCount:      5,
			urlError:       nil,
			userError:      nil,
			wantStatusCode: http.StatusOK,
		},
		{
			name: "negative test #2 - CountURLs error",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.100",
			},
			urlCount:       0,
			userCount:      0,
			urlError:       errors.New("database error"),
			userError:      nil,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "negative test #3 - CountUsers error",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.100",
			},
			urlCount:       10,
			userCount:      0,
			urlError:       nil,
			userError:      errors.New("database error"),
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name: "negative test #4 - IP not in trusted subnet",
			headers: map[string]string{
				"X-Real-IP": "10.0.0.1",
			},
			wantStatusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем ожидания только если IP валидный
			if tt.wantStatusCode != http.StatusForbidden {
				mockRepo.EXPECT().
					CountURLs(gomock.Any()).
					Return(tt.urlCount, tt.urlError).
					Times(1)

				// CountUsers вызывается только если CountURLs успешен
				if tt.urlError == nil {
					mockRepo.EXPECT().
						CountUsers(gomock.Any()).
						Return(tt.userCount, tt.userError).
						Times(1)
				}
			}

			// Создаем клиент и запрос
			req, err := http.NewRequest("GET", srv.URL+"/api/internal/stats", nil)
			assert.NoError(t, err)

			// Устанавливаем заголовки
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			// Выполняем запрос
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)

			// Для успешного случая проверяем тело ответа
			if tt.wantStatusCode == http.StatusOK {
				var response models.StatsResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, models.StatsResponse{
					URLs:  tt.urlCount,
					Users: tt.userCount,
				}, response)
			}
		})
	}
}

func TestGetHandler_WithoutTrustedSubnet(t *testing.T) {
	// Инициализация логгера
	if err := logger.Initialize("debug"); err != nil {
		t.Fatalf("logger initialization failed: %v", err)
	}

	// Конфигурация без доверенной подсети
	cfg := &config.Config{
		TrustedSubnet: "",
	}

	// Создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаём мок репозитория
	mockRepo := mocks.NewMockRepository(ctrl)
	service := service.NewService(mockRepo)

	// Инициализируем роутер с middleware
	r := chi.NewRouter()
	r.Use(mwlogger.RequestLogging(zap.L()))
	r.Use(trustednet.CheckTrustedSubnet(cfg.TrustedSubnet))
	r.Get("/api/internal/stats", stats.GetHandler(service, zap.L()))

	// Запускаем тестовый сервер
	srv := httptest.NewServer(r)
	defer srv.Close()

	t.Run("access denied when trusted subnet not configured", func(t *testing.T) {
		// Не ожидаем вызовов репозитория, так как middleware должен заблокировать запрос

		req, err := http.NewRequest("GET", srv.URL+"/api/internal/stats", nil)
		assert.NoError(t, err)
		req.Header.Set("X-Real-IP", "192.168.1.100") // Даже валидный IP не должен пройти

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
