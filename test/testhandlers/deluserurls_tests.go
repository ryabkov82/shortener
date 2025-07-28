package testhandlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/status"

	pb "github.com/ryabkov82/shortener/api"
	"github.com/ryabkov82/shortener/internal/app/jwtauth"
	"github.com/ryabkov82/shortener/internal/app/service"
	"github.com/ryabkov82/shortener/test/testutils"

	"github.com/go-resty/resty/v2"
)

// TestDelUserUrls тестирует обработчик удаления пользовательских URL (/api/user/urls).
//
// Проверяет следующие сценарии:
//   - Успешную пометку URL как удаленных (StatusAccepted)
//   - Пакетное удаление нескольких URL
//   - Попытку удаления несуществующих URL
//   - Попытку удаления URL другого пользователя
//   - Асинхронную природу удаления (проверка через последующий запрос)
//   - Работу сжатия gzip на входящих данных
//   - Авторизацию через JWT cookie
//
// Тест создает:
//   - Тестовое хранилище с URL для двух пользователей
//   - HTTP-сервер с middleware:
//   - Логирование
//   - Gzip сжатие
//   - JWT авторизация
//   - Набор тестовых случаев с разными сценариями удаления
//
// Примеры тест-кейсов:
//   - Удаление одного URL (ожидается 202 Accepted)
//   - Пакетное удаление (ожидается 202 Accepted)
//   - Попытка удаления URL другого пользователя
//   - Проверка статуса 410 Gone после удаления
//
// Особенности:
//   - Использует асинхронную модель удаления
//   - Проверяет реальное изменение статуса через последующий GET-запрос
//   - Поддерживает сжатие gzip в запросах
func TestDelUserUrls(t *testing.T, serv *service.Service, client *resty.Client) {

	user1URLs, _ := prepareTestURLs(serv)

	tests := CommonDelUserURLsCases(user1URLs)
	for _, tt := range tests {
		t.Run("HTTP"+tt.name, func(t *testing.T) {

			// Подготовка запроса
			body, _ := json.Marshal(tt.codesToDelete)
			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(body))
			assert.NoError(t, err)
			err = zb.Close()
			assert.NoError(t, err)

			// Запрос
			resp, err := client.R().
				SetBody(buf).
				SetCookie(tt.cookie).
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Accept-Encoding", "gzip").
				Delete("/api/user/urls")

			// Проверки
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, testutils.HTTPStatusToStatusCode(resp.StatusCode()))

			time.Sleep(500 * time.Millisecond)
			// Проверка что URL помечены как удаленные
			for _, code := range tt.codesToDelete {

				resp, err := client.R().
					SetCookie(tt.cookie).
					Get("/" + code)

				// Проверки
				assert.NoError(t, err)
				if resp.StatusCode() == http.StatusGone {
					assert.Contains(t, tt.shouldBeMarked, code)
				} else {
					assert.NotContains(t, tt.shouldBeMarked, code)
				}

			}
		})
	}
}

func TestDelUserUrlsGRPC(t *testing.T, serv *service.Service, grpcClient pb.ShortenerClient) {

	user1URLs, _ := prepareTestURLs(serv)

	tests := CommonDelUserURLsCases(user1URLs)
	for _, tt := range tests {
		t.Run("gRPC"+tt.name, func(t *testing.T) {

			token := tt.cookie.Value
			ctx := testutils.ContextWithJWT(context.Background(), token)

			_, err := grpcClient.DeleteUserURLs(ctx, &pb.DeleteRequest{ShortUrls: tt.codesToDelete})
			assert.NoError(t, err)

			time.Sleep(500 * time.Millisecond)
			// Проверка что URL помечены как удаленные
			for _, code := range tt.codesToDelete {

				_, err := grpcClient.GetOriginalURL(ctx, &pb.GetRequest{ShortUrl: code})

				var redirectStatus testutils.StatusCode
				if err != nil {
					if s, ok := status.FromError(err); ok {
						redirectStatus = testutils.GRPCCodeToStatusCode(s.Code())
					} else {
						redirectStatus = testutils.StatusInternalError
					}
				} else {
					redirectStatus = testutils.StatusTemporaryRedirect
				}

				// Проверки
				if redirectStatus == testutils.StatusNotFound {
					if len(tt.shouldBeMarked) > 0 {
						assert.Contains(t, tt.shouldBeMarked, code)
					}
				} else {
					assert.NotContains(t, tt.shouldBeMarked, code)
				}
			}

		})
	}
}

func prepareTestURLs(serv *service.Service) (map[string]string, map[string]string) {

	user1 := "user1"
	user2 := "user2"

	// Исходные данные
	user1Data := map[string]string{
		"url1": "https://example.com/1",
		"url2": "https://example.com/2",
		"url3": "https://example.com/3",
		"url4": "https://example.com/4",
	}
	user2Data := map[string]string{
		"url5": "https://example.com/5",
	}

	user1Shorts := map[string]string{}
	user2Shorts := map[string]string{}

	// Заполняем хранилище
	for key, original := range user1Data {
		ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, user1)
		shortKey, err := serv.GetShortKey(ctx, original)
		if err != nil {
			panic(fmt.Sprintf("failed to shorten URL %s: %v", original, err))
		}
		user1Shorts[key] = shortKey
	}

	for key, original := range user2Data {
		ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, user2)
		shortKey, err := serv.GetShortKey(ctx, original)
		if err != nil {
			panic(fmt.Sprintf("failed to shorten URL %s: %v", original, err))
		}
		user2Shorts[key] = shortKey
	}

	return user1Shorts, user2Shorts

}
