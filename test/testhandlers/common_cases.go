package testhandlers

import (
	"net/http"

	"github.com/ryabkov82/shortener/test/testutils"
)

type ShortenURLTestCase struct {
	Name        string
	OriginalURL string
	Want        testutils.StatusCode
	Cookie      *http.Cookie
}

// ShortenResult — общий результат для обоих клиентов
type ShortenResult struct {
	ShortURL string
	Status   testutils.StatusCode
}

func CommonShortenURLTestCases() []ShortenURLTestCase {

	cookie, _ := testutils.CreateSignedCookie()

	return []ShortenURLTestCase{
		{
			Name:        "valid URL",
			OriginalURL: "https://example.com",
			Want:        testutils.StatusCreated,
			Cookie:      cookie,
		},
		{
			Name:        "valid URL",
			OriginalURL: "https://example.com",
			Want:        testutils.StatusConflict,
			Cookie:      cookie,
		},
		/*
			{
				Name:        "invalid JWT",
				OriginalURL: "https://example.com",
				Want:        testutils.StatusUnauthorized,
				Setup: func() (context.Context, error) {
					return testutils.ContextWithJWT(context.Background(), "bad.token"), nil
				},
				Cookie: &http.Cookie{},
			},
		*/
		{
			Name:        "bad URL",
			OriginalURL: "not-a-url",
			Want:        testutils.StatusBadRequest,
			Cookie:      cookie,
		},
	}
}

type RedirectResult struct {
	Location string
	Status   testutils.StatusCode
}

type RedirectTestCase struct {
	Name           string
	ShortKey       string
	ExpectedStatus testutils.StatusCode
	ExpectedURL    string
}

func CommonRedirectTestCases(shortKey string, originalURL string) []RedirectTestCase {
	return []RedirectTestCase{
		{
			Name:           "valid redirect",
			ShortKey:       shortKey,
			ExpectedStatus: testutils.StatusTemporaryRedirect,
			ExpectedURL:    originalURL,
		},
		{
			Name:           "not found",
			ShortKey:       "not_existing_key",
			ExpectedStatus: testutils.StatusNotFound,
			ExpectedURL:    "",
		},
	}
}

type BatchTestCase struct {
	Name           string
	Request        string
	WantStatus     testutils.StatusCode
	ExpectedLength int // сколько объектов в ответе
}

func CommonBatchTestCases() []BatchTestCase {
	return []BatchTestCase{
		{
			Name: "valid batch request",
			Request: `[
			{"correlation_id": "123", "original_url": "https://example.com/page1"},
			{"correlation_id": "456", "original_url": "https://example.com/page2"}
		]`,
			WantStatus:     testutils.StatusCreated,
			ExpectedLength: 2,
		},
		{
			Name:           "invalid json request",
			Request:        `{}`,
			WantStatus:     testutils.StatusBadRequest,
			ExpectedLength: 0,
		},
	}
}

type DelUserURLsTestCase struct {
	cookie         *http.Cookie
	name           string
	userID         string
	codesToDelete  []string
	shouldBeMarked []string
	wantStatus     testutils.StatusCode
}

func CommonDelUserURLsCases(user1URLs map[string]string) []DelUserURLsTestCase {

	user1 := "user1"
	user2 := "user2"
	cookie1, err := testutils.CreateCookieByUserID(user1)

	if err != nil {
		panic(err)
	}
	cookie2, err := testutils.CreateCookieByUserID(user2)

	if err != nil {
		panic(err)
	}

	return []DelUserURLsTestCase{
		{
			name:           "successful deletion",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{user1URLs["url1"]},
			wantStatus:     testutils.StatusAccepted,
			shouldBeMarked: []string{user1URLs["url1"]},
		},
		{
			name:           "delete multiple",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{user1URLs["url2"], user1URLs["url3"]},
			wantStatus:     testutils.StatusAccepted,
			shouldBeMarked: []string{user1URLs["url2"], user1URLs["url3"]},
		},
		{
			name:           "delete non-existent",
			userID:         user1,
			cookie:         cookie1,
			codesToDelete:  []string{"nonexistent"},
			wantStatus:     testutils.StatusAccepted,
			shouldBeMarked: []string{},
		},
		{
			name:           "delete other user's url",
			userID:         user2,
			cookie:         cookie2,
			codesToDelete:  []string{user1URLs["url4"]},
			wantStatus:     testutils.StatusAccepted,
			shouldBeMarked: []string{}, // Не должно пометить как удаленный
		},
	}
}
