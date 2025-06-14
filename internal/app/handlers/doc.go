// Пакет handlers содержит HTTP-обработчики для API сервиса сокращения URL.
//
// Основные обработчики:
//
//   - **shorturl** - создание короткой ссылки через форму (POST /)
//   - **redirect** - перенаправление по короткой ссылке (GET /{id})
//   - **shortenapi** - JSON API для создания короткой ссылки (POST /api/shorten)
//   - **batch** - пакетное создание ссылок (POST /api/shorten/batch)
//   - **userurls** - получение списка ссылок пользователя (GET /api/user/urls)
//   - **deluserurls** - удаление ссылок пользователя (DELETE /api/user/urls)
//   - **ping** - проверка доступности БД (GET /ping)
//
// Все обработчики:
//   - Принимают зависимости через замыкание (сервис, базовый URL, логгер)
//   - Возвращают стандартные http.HandlerFunc
//   - Поддерживают контекст запроса
//   - Обрабатывают ошибки и логируют их
//
// Пример создания обработчика:
//
//	func GetHandler(srv *service.Service, baseURL string, log *zap.Logger) http.HandlerFunc {
//	    return func(w http.ResponseWriter, r *http.Request) {
//	        // Логика обработки
//	    }
//	}
//
// Особенности:
//   - Единый стиль обработки ошибок
//   - Поддержка разных форматов ответов (текст, JSON)
//   - Интеграция с системой аутентификации
package handlers
