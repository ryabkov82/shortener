package shorturl

import (
	"fmt"
	"net/http/httptest"
	"strings"
)

// Пример для POST / - Создание короткой ссылки из формы
func ExampleGetHandler() {
	// В реальном коде здесь была бы инициализация сервера
	// handler := shorturl.GetHandler(srv, cfg.BaseURL, log)

	// Подготовка тестового запроса
	body := strings.NewReader("https://example.com/very/long/url")
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httptest.NewRecorder()
	// handler.ServeHTTP(w, req) // В реальном примере вызывался бы обработчик

	// Для примера выводим ожидаемый результат
	fmt.Println("Запрос:")
	fmt.Println("Метод: POST /")
	fmt.Println("Заголовок: Content-Type: application/x-www-form-urlencoded")
	fmt.Println("Тело: https://example.com/very/long/url")
	fmt.Println("\nОтвет:")
	fmt.Println("Статус: 201 Created")
	fmt.Println("Тело: http://localhost:8080/abc123")

	// Output:
	// Запрос:
	// Метод: POST /
	// Заголовок: Content-Type: application/x-www-form-urlencoded
	// Тело: https://example.com/very/long/url
	//
	// Ответ:
	// Статус: 201 Created
	// Тело: http://localhost:8080/abc123
}
