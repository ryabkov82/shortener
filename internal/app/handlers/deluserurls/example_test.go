package deluserurls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
)

// Пример для DELETE /api/user/urls - Удаление ссылок пользователя
func Example_deleteUserURLs() {
	// handler := deluserurls.GetHandler(srv, cfg.BaseURL, log)

	request := []string{"abc123", "def456"}
	body, _ := json.Marshal(request)

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Здесь бы был заголовок авторизации

	httptest.NewRecorder()
	// handler.ServeHTTP(w, req)

	fmt.Println("Запрос:")
	fmt.Println("Метод: DELETE /api/user/urls")
	fmt.Println("Заголовок: Content-Type: application/json")
	fmt.Printf("Тело: %s\n", body)
	fmt.Println("\nОтвет (успешный):")
	fmt.Println("Статус: 202 Accepted")
	fmt.Println("\nОтвет (ошибка):")
	fmt.Println("Статус: 500 Internal Server Error")

	// Output:
	// Запрос:
	// Метод: DELETE /api/user/urls
	// Заголовок: Content-Type: application/json
	// Тело: ["abc123","def456"]
	//
	// Ответ (успешный):
	// Статус: 202 Accepted
	//
	// Ответ (ошибка):
	// Статус: 500 Internal Server Error
}
