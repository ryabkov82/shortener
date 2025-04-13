package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ryabkov82/shortener/internal/app/jwtauth"
)

func JWTAutoIssue(jwtKey []byte) func(next http.Handler) http.Handler {
	// Middleware: проверяет JWT или выдаёт новый

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("token")
			if err != nil || cookie == nil {
				//userID := issueNewToken(w, jwtKey)
				ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, jwtauth.UserDefault) // для поддержки обратной совместимости
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
				//userID := issueNewToken(w, jwtKey)
				ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, jwtauth.UserDefault) // для поддержки обратной совместимости
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), jwtauth.UserIDContextKey, claims.UserID)
			r = r.WithContext(ctx)
			// Передаем управление следующему обработчику в цепочке middleware
			next.ServeHTTP(w, r)
		}

		// Возвращаем созданный выше обработчик, приведя его к типу http.HandlerFunc
		return http.HandlerFunc(fn)

	}
}

func StrictJWTAutoIssue(jwtKey []byte) func(next http.Handler) http.Handler {
	// Middleware: проверяет JWT или выдаёт новый

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
			// Передаем управление следующему обработчику в цепочке middleware
			next.ServeHTTP(w, r)
		}

		// Возвращаем созданный выше обработчик, приведя его к типу http.HandlerFunc
		return http.HandlerFunc(fn)

	}
}

// Выдаёт новый токен и куку
func issueNewToken(w http.ResponseWriter, jwtKey []byte) string {
	token, userID, err := jwtauth.GenerateNewToken(jwtKey)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return ""
	}
	setTokenCookie(w, token)
	//log.Printf("Issued new JWT for user: %s", userID)
	return userID
}

// Устанавливает JWT в куки
func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		//Secure:   true, // HTTPS-only
		SameSite: http.SameSiteStrictMode,
	})
}
