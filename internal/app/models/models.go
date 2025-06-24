// Package models содержит основные структуры данных приложения.
//
// Пакет определяет:
// - Модели для хранения URL
// - Модели для пакетной обработки
// - Структуры для API-ответов
package models

// URLMapping представляет соответствие между коротким и оригинальным URL.
//
// Используется в API-ответах при:
// - Получении списка URL пользователя
// - Запросе информации об отдельном URL
//
// Пример JSON:
//
//	{
//	  "short_url": "http://short.ly/abc",
//	  "original_url": "https://example.com/long/url"
//	}
type URLMapping struct {
	ShortURL    string `json:"short_url"`    // Полный сокращённый URL
	OriginalURL string `json:"original_url"` // Оригинальный длинный URL
}

// UserURLMapping расширяет URLMapping информацией о пользователе и статусе.
//
// Содержит дополнительные поля:
// - UUID - уникальный идентификатор записи
// - UserID - идентификатор пользователя-владельца
// - DeletedFlag - флаг мягкого удаления
//
// Используется в:
// - Системе хранения URL
// - Административных функциях
type UserURLMapping struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	UUID        uint64 `json:"uuid"`
	DeletedFlag bool   `json:"is_deleted"`
}

// BatchRequest представляет элемент запроса для пакетного создания URL.
//
// Используется в API:
//
//	POST /api/shorten/batch
//
// Пример JSON:
//
//	{
//	  "correlation_id": "123e4567",
//	  "original_url": "https://example.com"
//	}
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"` // Уникальный ID для сопоставления запроса/ответа
	OriginalURL   string `json:"original_url"`   // URL для сокращения
}

// BatchResponse представляет элемент ответа при пакетном создании URL.
//
// Содержит:
// - CorrelationID из исходного запроса
// - Сгенерированный короткий URL
//
// Пример JSON:
//
//	{
//	  "correlation_id": "123e4567",
//	  "short_url": "http://short.ly/abc"
//	}
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"` // Соответствует ID из запроса
	ShortURL      string `json:"short_url"`      // Полный сокращённый URL
}
