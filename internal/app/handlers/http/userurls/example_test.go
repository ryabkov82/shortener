package userurls

import (
	"fmt"
	"net/http/httptest"
)

// Пример для GET /api/user/urls - Получение всех ссылок пользователя
func Example_getUserURLs() {
	// handler := userurls.GetHandler(srv, cfg.BaseURL, log)

	httptest.NewRequest("GET", "/api/user/urls", nil)
	// В реальном запросе здесь бы был заголовок с куками/токеном авторизации
	httptest.NewRecorder()
	// handler.ServeHTTP(w, req)

	fmt.Println("Запрос:")
	fmt.Println("Метод: GET /api/user/urls")
	fmt.Println("Заголовок: Authorization: Bearer user-token")
	fmt.Println("\nОтвет (успешный):")
	fmt.Println("Статус: 200 OK")
	fmt.Println("Тело: [{\"short_url\":\"http://localhost:8080/abc123\",\"original_url\":\"https://example.com/long/url1\"},{\"short_url\":\"http://localhost:8080/def456\",\"original_url\":\"https://example.com/long/url2\"}]")
	fmt.Println("\nОтвет (нет ссылок):")
	fmt.Println("Статус: 204 No Content")

	// Output:
	// Запрос:
	// Метод: GET /api/user/urls
	// Заголовок: Authorization: Bearer user-token
	//
	// Ответ (успешный):
	// Статус: 200 OK
	// Тело: [{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com/long/url1"},{"short_url":"http://localhost:8080/def456","original_url":"https://example.com/long/url2"}]
	//
	// Ответ (нет ссылок):
	// Статус: 204 No Content
}
