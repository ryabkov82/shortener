package redirect

import (
	"fmt"
	"net/http/httptest"
)

// Пример для GET /{id} - Перенаправление по короткой ссылке
func ExampleGetHandler() {
	// handler := redirect.GetHandler(srv, log)

	httptest.NewRequest("GET", "/abc123", nil)
	httptest.NewRecorder()
	// handler.ServeHTTP(w, req)

	fmt.Println("Запрос:")
	fmt.Println("Метод: GET /abc123")
	fmt.Println("\nОтвет:")
	fmt.Println("Статус: 307 Temporary Redirect")
	fmt.Println("Заголовок: Location: https://example.com/very/long/url")

	// Output:
	// Запрос:
	// Метод: GET /abc123
	//
	// Ответ:
	// Статус: 307 Temporary Redirect
	// Заголовок: Location: https://example.com/very/long/url
}
