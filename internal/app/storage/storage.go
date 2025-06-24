// Package storage определяет интерфейсы хранилища и общие ошибки для работы с URL.
package storage

import "errors"

// Общие ошибки хранилища URL
var (
	// ErrURLNotFound возвращается, когда запрошенный URL не найден в хранилище.
	ErrURLNotFound = errors.New("url not found")

	// ErrURLExists возвращается при попытке сохранить URL, который уже существует.
	// Используется, когда оригинальный URL уже имеет сокращенную версию.
	ErrURLExists = errors.New("url exists")

	// ErrShortURLExists возвращается при попытке использовать уже занятый короткий URL.
	// Отличается от ErrURLExists тем, что указывает на конфликт именно по короткому URL.
	ErrShortURLExists = errors.New("short URL already exists")

	// ErrURLDeleted возвращается при попытке доступа к URL, помеченному как удаленный.
	ErrURLDeleted = errors.New("URL has been deleted")
)
