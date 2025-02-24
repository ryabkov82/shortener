package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// Переменная для хранения редиректов ShortURL -> OriginalURL
var shortURLs = make(map[string]string)

// Переменная для хранения значений OriginalURL -> ShortURL
var originalURLs = make(map[string]string)

func GetShortURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println(err)
		return
	}
	originalURL := string(body)

	if originalURL == "" {
		http.Error(res, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// Проверяем, что передан URL
	parsedURL, err := url.ParseRequestURI(originalURL)

	if err != nil {
		http.Error(res, "invalid request", http.StatusBadRequest)
		log.Println(err)
		return
	}

	log.Println("get URL", parsedURL)

	// Возможно, shortURL уже сгенерирован...
	shortKey, found := originalURLs[originalURL]
	if !found {
		// Генерируем короткий URL
		generated := false
		for i := 1; i < 3; i++ {
			shortKey = generateShortKey()
			// Возможно, shortURL был сгененрирован ранее
			_, found := shortURLs[shortKey]
			if !found {
				generated = true
				break
			}
		}
		if generated {
			// Cохраняем переданный URL
			shortURLs[shortKey] = originalURL
			originalURLs[originalURL] = shortKey
		} else {
			// Не удалось сгененрировать новый shortURL
			http.Error(res, "Failed to generate a new shortURL", http.StatusBadRequest)
			log.Println("Failed to generate a new shortURL")
			return
		}
	}

	res.Header().Set("content-type", "text/plain")
	// устанавливаем код 201
	res.WriteHeader(http.StatusCreated)
	// пишем тело ответа
	resp := fmt.Sprintf("http://localhost:8080/%s", shortKey)
	res.Write([]byte(resp))

}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 8

	rand.New(rand.NewSource(time.Now().UnixNano()))

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

func GetRedirectURL(res http.ResponseWriter, req *http.Request) {

	id := req.PathValue("id")

	// Получаем адрес перенаправления
	originalURL, found := shortURLs[id]
	if !found {
		http.Error(res, "Shortened key not found", http.StatusNotFound)
		log.Println("Shortened key not found", id)
		return
	}
	log.Println("Shortened key found", id, "redirect", originalURL)
	// Устанавливаем заголовок ответа Location
	res.Header().Set("Location", originalURL)
	// устанавливаем код 307
	res.WriteHeader(http.StatusTemporaryRedirect)

}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", GetShortURL)
	mux.HandleFunc("GET /{id}", GetRedirectURL)
	http.ListenAndServe("localhost:8080", mux)
}
