package models

// URLMapping представляет собой структуру для хранения соответствия короткого и оригинального URL.
type URLMapping struct {
	ShortURL    string // Короткий URL
	OriginalURL string // Оригинальный URL
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
