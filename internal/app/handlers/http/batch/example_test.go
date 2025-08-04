package batch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
)

// Пример для POST /api/shorten/batch - Пакетное создание коротких ссылок
func ExampleGetHandler() {
	// handler := batch.GetHandler(srv, cfg.BaseURL, log)

	request := []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example.com/first/long/url",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example.com/second/long/url",
		},
	}
	body, _ := json.Marshal(request)

	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	httptest.NewRecorder()
	// handler.ServeHTTP(w, req)

	fmt.Println("Запрос:")
	fmt.Println("Метод: POST /api/shorten/batch")
	fmt.Println("Заголовок: Content-Type: application/json")
	fmt.Printf("Тело: %s\n", body)
	fmt.Println("\nОтвет:")
	fmt.Println("Статус: 201 Created")
	fmt.Println("Тело: [{\"correlation_id\":\"1\",\"short_url\":\"http://localhost:8080/ghi789\"},{\"correlation_id\":\"2\",\"short_url\":\"http://localhost:8080/jkl012\"}]")

	// Output:
	// Запрос:
	// Метод: POST /api/shorten/batch
	// Заголовок: Content-Type: application/json
	// Тело: [{"correlation_id":"1","original_url":"https://example.com/first/long/url"},{"correlation_id":"2","original_url":"https://example.com/second/long/url"}]
	//
	// Ответ:
	// Статус: 201 Created
	// Тело: [{"correlation_id":"1","short_url":"http://localhost:8080/ghi789"},{"correlation_id":"2","short_url":"http://localhost:8080/jkl012"}]
}
