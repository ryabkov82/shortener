// Пакет auth предоставляет middleware для аутентификации через JWT.
package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ryabkov82/shortener/internal/app/jwtauth"
)

// JWTAutoIssue создает middleware для автоматической выдачи JWT токенов.
//
// Middleware проверяет наличие валидного JWT токена в cookies:
// - Если токен отсутствует или невалиден - выдает новый токен
// - Если токен валиден - извлекает userID и передает в контекст
//
// Параметры:
//
//	jwtKey - ключ для подписи JWT токенов
//
// Возвращает:
//
//	func(next http.Handler) http.Handler - middleware функцию
func JWTAutoIssue(jwtKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil || cookie == nil {
				userID := issueNewToken(w, jwtKey)
				ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, userID)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			tokenStr := cookie.Value
			claims := &jwtauth.Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				userID := issueNewToken(w, jwtKey)
				ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, userID)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, claims.UserID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// StrictJWTAutoIssue создает строгий middleware для проверки JWT токенов.
//
// В отличие от JWTAutoIssue:
// - Не выдает новый токен при отсутствии/невалидности текущего
// - Возвращает 401 Unauthorized при отсутствии валидного токена
//
// Параметры:
//
//	jwtKey - ключ для подписи JWT токенов
//
// Возвращает:
//
//	func(next http.Handler) http.Handler - middleware функцию
func StrictJWTAutoIssue(jwtKey []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil || cookie == nil {
				_ = issueNewToken(w, jwtKey)
				http.Error(w, "Status unauthorized", http.StatusUnauthorized)
				return
			}

			tokenStr := cookie.Value
			claims := &jwtauth.Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				_ = issueNewToken(w, jwtKey)
				http.Error(w, "Status unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, claims.UserID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// issueNewToken генерирует и устанавливает новый JWT токен.
//
// Параметры:
//
//	w - http.ResponseWriter для установки cookie
//	jwtKey - ключ для подписи JWT
//
// Возвращает:
//
//	string - идентификатор пользователя (userID)
func issueNewToken(w http.ResponseWriter, jwtKey []byte) string {
	token, userID, err := jwtauth.GenerateNewToken(jwtKey)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return ""
	}
	setTokenCookie(w, token)
	return userID
}

// setTokenCookie устанавливает JWT токен в cookie.
//
// Параметры:
//
//	w - http.ResponseWriter для установки cookie
//	token - JWT токен для сохранения
func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})
}
