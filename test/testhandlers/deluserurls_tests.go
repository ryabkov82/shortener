package testhandlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ryabkov82/shortener/internal/app/models"

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

	// Тестовые данные
	cookie1, user1 := testutils.CreateSignedCookie()
	cookie2, user2 := testutils.CreateSignedCookie()
	testURLs := []models.UserURLMapping{
		{UserID: user1, OriginalURL: "https://example.com/1"},
		{UserID: user1, OriginalURL: "https://example.com/2"},
		{UserID: user1, OriginalURL: "https://example.com/3"},
		{UserID: user1, OriginalURL: "https://example.com/4"},
		{UserID: user2, OriginalURL: "https://example.com/5"},
	}

	// Заполняем хранилище
	for i, url := range testURLs {
		ctx := context.WithValue(context.Background(), jwtauth.UserIDContextKey, url.UserID)
		shortURL, err := serv.GetShortKey(ctx, url.OriginalURL)
		if err != nil {
			panic(err)
		}
		testURLs[i].ShortURL = shortURL
	}

	tests := []struct {
		cookie         *http.Cookie
		name           string
		userID         string
		codesToDelete  []string
		shouldBeMarked []string
		wantStatus     int
	}{
		{
			name:           "successful deletion",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{testURLs[0].ShortURL},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{testURLs[0].ShortURL},
		},
		{
			name:           "delete multiple",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{testURLs[1].ShortURL, testURLs[2].ShortURL},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{testURLs[1].ShortURL, testURLs[2].ShortURL},
		},
		{
			name:           "delete non-existent",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{"nonexistent"},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{},
		},
		{
			name:           "delete other user's url",
			userID:         user2,
			cookie:         cookie2,
			codesToDelete:  []string{testURLs[3].ShortURL},
			wantStatus:     http.StatusAccepted,
			shouldBeMarked: []string{}, // Не должно пометить как удаленный
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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
			assert.Equal(t, tt.wantStatus, resp.StatusCode())

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
