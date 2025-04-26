package models

// URLMapping представляет собой структуру для хранения соответствия короткого и оригинального URL.
type URLMapping struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UserURLMapping struct {
	UUID        uint64 `json:"uuid"`
	ShortURL    string `json:"short_url"`    // Короткий URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
