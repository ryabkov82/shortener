package jwtauth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Определяем свой тип для ключа контекста.
type ContextKey string

// Константа для ключа
const UserIDContextKey ContextKey = "userID"

// Генерирует новый JWT с уникальным ID пользователя
func GenerateNewToken(jwtKey []byte) (string, string, error) {
	userID := uuid.New().String() // Генерируем уникальный ID
	//expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		//	RegisteredClaims: jwt.RegisteredClaims{
		//		ExpiresAt: jwt.NewNumericDate(expirationTime),
		//	},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	return tokenString, userID, err
}
