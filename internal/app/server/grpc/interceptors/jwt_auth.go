package interceptors

import (
	"context"
	"fmt"

	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"go.uber.org/zap"

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
func JWTAutoIssueGRPC(jwtKey []byte, log *zap.Logger) grpc.UnaryServerInterceptor {
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
			log.Debug("no metadata in context",
				zap.String("method", info.FullMethod))
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey, log, info.FullMethod)
		}

		tokens := md.Get("token")
		if len(tokens) == 0 {
			log.Debug("no token provided",
				zap.String("method", info.FullMethod))
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey, log, info.FullMethod)
		}

		// Валидация токена
		claims, err := validateToken(tokens[0], jwtKey)
		if err != nil {
			log.Warn("invalid token",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey, log, info.FullMethod)
		}

		// Проверка наличия userID в claims
		if claims.UserID == "" {
			log.Warn("empty userID in token",
				zap.String("method", info.FullMethod))
			return issueNewTokenAndHandle(ctx, req, handler, jwtKey, log, info.FullMethod)
		}

		newCtx := context.WithValue(ctx, jwtauth.UserIDContextKey, claims.UserID)
		return handler(newCtx, req)
	}
}

func StrictJWTAutoIssueGRPC(jwtKey []byte, log *zap.Logger) grpc.UnaryServerInterceptor {
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
			log.Warn("no metadata in context",
				zap.String("method", info.FullMethod))

			return handleMissingToken(ctx, jwtKey, log, info.FullMethod)
		}

		tokens := md.Get("token")
		if len(tokens) == 0 {
			log.Warn("no token provided",
				zap.String("method", info.FullMethod))

			return handleMissingToken(ctx, jwtKey, log, info.FullMethod)
		}

		// Валидация токена
		claims, err := validateToken(tokens[0], jwtKey)

		if err != nil {
			log.Warn("invalid token",
				zap.String("method", info.FullMethod),
				zap.Error(err))
			return handleInvalidToken(ctx, jwtKey, log, info.FullMethod)
		}

		// Проверка наличия userID в claims
		if claims.UserID == "" {
			log.Warn("empty userID in token",
				zap.String("method", info.FullMethod))
			return handleInvalidToken(ctx, jwtKey, log, info.FullMethod)
		}

		// Добавляем userID в контекст
		newCtx := context.WithValue(ctx, jwtauth.UserIDContextKey, claims.UserID)
		return handler(newCtx, req)
	}
}

// Вспомогательные функции
func validateToken(tokenString string, jwtKey []byte) (*jwtauth.Claims, error) {
	claims := &jwtauth.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func handleMissingToken(ctx context.Context, jwtKey []byte, logger *zap.Logger, method string) (interface{}, error) {
	if err := issueNewToken(ctx, jwtKey); err != nil {
		logger.Error("failed to issue token",
			zap.String("method", method),
			zap.Error(err))
	}
	return nil, status.Error(codes.Unauthenticated, "authentication required")
}

func handleInvalidToken(ctx context.Context, jwtKey []byte, logger *zap.Logger, method string) (interface{}, error) {
	if err := issueNewToken(ctx, jwtKey); err != nil {
		logger.Error("failed to issue token",
			zap.String("method", method),
			zap.Error(err))
	}
	return nil, status.Error(codes.Unauthenticated, "invalid token")
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
	logger *zap.Logger,
	method string,
) (interface{}, error) {
	// Генерация нового токена
	token, userID, err := jwtauth.GenerateNewToken(jwtKey)
	if err != nil {
		logger.Error("failed to issue new token",
			zap.String("method", method),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "authentication error")
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
