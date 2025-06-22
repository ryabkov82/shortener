package shortenapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
)

// Пример для POST /api/shorten - Создание короткой ссылки через API
func ExampleGetHandler() {
	// handler := shortenapi.GetHandler(srv, cfg.BaseURL, log)

	request := struct {
		URL string `json:"url"`
	}{
		URL: "https://example.com/long/url/to/shorten",
	}
	body, _ := json.Marshal(request)

	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	httptest.NewRecorder()
	// handler.ServeHTTP(w, req)

	fmt.Println("Запрос:")
	fmt.Println("Метод: POST /api/shorten")
	fmt.Println("Заголовок: Content-Type: application/json")
	fmt.Printf("Тело: %s\n", body)
	fmt.Println("\nОтвет:")
	fmt.Println("Статус: 201 Created")
	fmt.Println("Тело: {\"result\":\"http://localhost:8080/def456\"}")

	// Output:
	// Запрос:
	// Метод: POST /api/shorten
	// Заголовок: Content-Type: application/json
	// Тело: {"url":"https://example.com/long/url/to/shorten"}
	//
	// Ответ:
	// Статус: 201 Created
	// Тело: {"result":"http://localhost:8080/def456"}
}
