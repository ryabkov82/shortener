// auth_testutils.go
package testutils

import (
	"net/http"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
)

var (
	TestSecretKey = []byte("test-secret-key")
)

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
