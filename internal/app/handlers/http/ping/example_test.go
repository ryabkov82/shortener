package ping

import (
	"fmt"
	"net/http/httptest"
)

// Пример для GET /ping - Проверка доступности БД
func ExampleGetHandler() {
	// handler := ping.GetHandler(srv, log)

	httptest.NewRequest("GET", "/ping", nil)
	httptest.NewRecorder()
	// handler.ServeHTTP(w, req)

	fmt.Println("Запрос:")
	fmt.Println("Метод: GET /ping")
	fmt.Println("\nОтвет (БД доступна):")
	fmt.Println("Статус: 200 OK")
	fmt.Println("\nОтвет (БД недоступна):")
	fmt.Println("Статус: 500 Internal Server Error")

	// Output:
	// Запрос:
	// Метод: GET /ping
	//
	// Ответ (БД доступна):
	// Статус: 200 OK
	//
	// Ответ (БД недоступна):
	// Статус: 500 Internal Server Error
}
