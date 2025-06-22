// Пакет jwtauth предоставляет функционал для работы с JWT токенами аутентификации.
package jwtauth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims представляет кастомные claims JWT токена.
// Содержит идентификатор пользователя и стандартные claims.
type Claims struct {
	UserID               string `json:"user_id"` // Уникальный идентификатор пользователя
	jwt.RegisteredClaims        // Стандартные claims JWT
}

// ContextKey тип для ключей контекста.
// Используется для безопасного доступа к значениям в context.Context.
type ContextKey string

// UserIDContextKey ключ для хранения ID пользователя в контексте.
const UserIDContextKey ContextKey = "userID"

// GenerateNewToken генерирует новый JWT токен для пользователя.
//
// Параметры:
//   - jwtKey: секретный ключ для подписи токена
//
// Возвращает:
//   - string: подписанный JWT токен
//   - string: сгенерированный ID пользователя
//   - error: ошибка генерации токена
//
// Пример использования:
//
//	token, userID, err := GenerateNewToken([]byte("secret"))
func GenerateNewToken(jwtKey []byte) (string, string, error) {
	userID := uuid.New().String() // Генерация уникального ID пользователя

	claims := &Claims{
		UserID: userID,
		// При необходимости можно добавить время экспирации:
		// RegisteredClaims: jwt.RegisteredClaims{
		//     ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		// },
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, userID, err
}
