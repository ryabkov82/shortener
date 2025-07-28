package interceptors

import (
	"context"
	"fmt"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTAutoIssueGRPC возвращает gRPC-интерцептор для JWT-аутентификации с автоматической выдачей токенов.
//
// Интерцептор выполняет:
//   - Проверку JWT-токена из метаданных gRPC-запроса (ключ "token")
//   - Автоматическую генерацию и выдачу нового токена, если:
//   - Токен отсутствует
//   - Токен невалиден
//   - Токен не содержит UserID
//   - Добавление UserID в контекст вызова при успешной аутентификации
//   - Пропуск проверки для публичных методов (см. isPublicMethod)
//
// Параметры:
//   - jwtKey: секретный ключ для подписи JWT-токенов
//
// Возвращает:
//   - grpc.UnaryServerInterceptor: настроенный интерцептор аутентификации
//
// Особенности работы:
//   - Для извлечения токена используются метаданные gRPC (metadata)
//   - Новый токен устанавливается в заголовки ответа (grpc.SetHeader)
//   - Публичные методы определяются функцией isPublicMethod
//   - Контекст обогащается UserID (ключ jwtauth.UserIDContextKey)
//
// Пример использования:
//
//	server := grpc.NewServer(
//	    grpc.ChainUnaryInterceptor(
//	        interceptor.JWTAutoIssueGRPC([]byte("secret")),
//	    ),
//	)
//
// Схема работы:
//  1. Проверка метода в isPublicMethod
//  2. Извлечение токена из metadata["token"]
//  3. Валидация токена и проверка claims
//  4. Выдача нового токена при необходимости
//  5. Добавление UserID в контекст
//  6. Вызов основного обработчика
func JWTAutoIssueGRPC(jwtKey []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Пропускаем аутентификацию для публичных методов
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Аналог r.Cookie() в HTTP - получаем токен из метаданных
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey)
		}

		tokens := md.Get("token")
		if len(tokens) == 0 {
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey)
		}

		// Парсинг и валидация токена (аналогично HTTP)
		claims := &jwtauth.Claims{}

		token, err := jwt.ParseWithClaims(tokens[0], claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey)
		}

		// Если токен валиден, добавляем userID в контекст
		userID := claims.UserID
		if userID == "" {
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey)
		}

		newCtx := context.WithValue(ctx, jwtauth.UserIDContextKey, claims.UserID)
		return handler(newCtx, req)
	}
}

func StrictJWTAutoIssueGRPC(jwtKey []byte) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Пропускаем аутентификацию для публичных методов
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {

			err := issueNewToken(ctx, jwtKey)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate token")
			}
			return nil, status.Error(codes.Unauthenticated, "status unauthenticated")
		}

		tokens := md.Get("token")
		if len(tokens) == 0 {
			err := issueNewToken(ctx, jwtKey)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate token")
			}
			return nil, status.Error(codes.Unauthenticated, "status unauthenticated")
		}

		// Парсинг и валидация токена (аналогично HTTP)
		claims := &jwtauth.Claims{}

		token, err := jwt.ParseWithClaims(tokens[0], claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			err := issueNewToken(ctx, jwtKey)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate token")
			}
			return nil, status.Error(codes.Unauthenticated, "status unauthenticated")
		}

		// Если токен валиден, добавляем userID в контекст
		userID := claims.UserID
		if userID == "" {
			err := issueNewToken(ctx, jwtKey)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate token")
			}
			return nil, status.Error(codes.Unauthenticated, "status unauthenticated")
		}

		newCtx := context.WithValue(ctx, jwtauth.UserIDContextKey, claims.UserID)
		return handler(newCtx, req)
	}
}

func issueNewToken(
	ctx context.Context,
	jwtKey []byte,
) error {
	// Генерация нового токена
	token, _, err := jwtauth.GenerateNewToken(jwtKey)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create token")
	}

	// Установка токена в заголовки ответа (аналог Set-Cookie)
	header := metadata.Pairs("token", token)
	grpc.SetHeader(ctx, header)

	return nil
}

// Аналог issueNewToken в HTTP-версии
func issueNewTokenAndHandle(
	ctx context.Context,
	req interface{},
	handler grpc.UnaryHandler,
	jwtKey []byte,
) (interface{}, error) {
	// Генерация нового токена
	token, userID, err := jwtauth.GenerateNewToken(jwtKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create token")
	}

	// Установка токена в заголовки ответа (аналог Set-Cookie)
	header := metadata.Pairs("token", token)
	grpc.SetHeader(ctx, header)

	// Добавление userID в контекст
	newCtx := context.WithValue(ctx, jwtauth.UserIDContextKey, userID)
	return handler(newCtx, req)
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/grpc.health.v1.Health/Check": true,
		// Другие публичные методы
	}
	return publicMethods[method]
}
