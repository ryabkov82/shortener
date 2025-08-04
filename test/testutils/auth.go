package testutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TestSecretKey содержит тестовый ключ для подписи JWT.
// Используется только в тестовой среде.
var TestSecretKey = []byte("test-secret-key")

// CreateSignedCookie создает тестовую HTTP-куку с подписанным JWT-токеном.
// Используется для симуляции аутентифицированных запросов в тестах.
func CreateSignedCookie() (*http.Cookie, string) {
	tokenString, userID, err := jwtauth.GenerateNewToken(TestSecretKey)
	if err != nil {
		panic(err) // В тестах допустим panic
	}

	return &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}, userID
}

// CreateJWTToken создает тестовый JWT токен
func CreateCookieByUserID(userID string) (*http.Cookie, error) {
	tokenString, err := jwtauth.CreateToken(TestSecretKey, userID)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}, nil
}

// ContextWithJWT добавляет JWT токен в контекст gRPC
func ContextWithJWT(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{
		"token": token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

// ParseUserIDFromContext извлекает userID из контекста
func ParseUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not provided")
	}

	tokens := md.Get("token")
	if len(tokens) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token not provided")
	}

	claims := &jwtauth.Claims{}

	token, err := jwt.ParseWithClaims(tokens[0], claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return TestSecretKey, nil
	})

	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	userID := claims.UserID

	if userID != "" && token.Valid {
		return userID, nil
	}

	return "", status.Error(codes.Unauthenticated, "invalid token claims")
}
