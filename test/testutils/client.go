package testutils

import (
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
)

// TestClient представляет тестовый HTTP клиент с ассоциированным сервером.
// Используется для тестирования HTTP обработчиков с полным жизненным циклом.
//
// Поля:
//
//	Client *resty.Client - предварительно настроенный HTTP клиент
//	Server *httptest.Server - тестовый HTTP сервер
type TestClient struct {
	Client *resty.Client
	Server *httptest.Server
}

// NewTestClient создает новое тестовое окружение для HTTP тестов.
//
// Параметры:
//
//	handlers ...func(r chi.Router) - обработчики для настройки роутера.
//	  Каждая функция получает возможность зарегистрировать свои эндпоинты.
//	  Можно передавать несколько обработчиков - они будут применены по порядку.
//
// Возвращает:
//
//	*TestClient - готовый к использованию тестовый клиент с сервером.
//
// Примеры:
//
//  1. Простое создание:
//     tc := NewTestClient(
//     func(r chi.Router) {
//     r.Get("/ping", PingHandler)
//     },
//     )
//     defer tc.Close()
//
//  2. С middleware:
//     tc := NewTestClient(
//     func(r chi.Router) {
//     r.Use(LoggerMiddleware)
//     r.Post("/data", DataHandler)
//     },
//     )
func NewTestClient(handlers ...func(r chi.Router)) *TestClient {
	r := chi.NewRouter()
	for _, h := range handlers {
		h(r)
	}
	srv := httptest.NewServer(r)

	return &TestClient{
		Client: resty.New().SetBaseURL(srv.URL),
		Server: srv,
	}
}

// Close освобождает ресурсы тестового клиента.
// Должен вызываться при завершении работы (обычно через defer).
//
// Выполняет:
//   - Остановку тестового сервера
//   - Закрытие всех idle-соединений клиента
func (tc *TestClient) Close() {
	tc.Server.Close()
	tc.Client.GetClient().CloseIdleConnections()
}
