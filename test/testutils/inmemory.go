package testutils

import (
	"os"
	"path/filepath"

	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"
)

// InitializeInMemoryStorage создает и инициализирует хранилище в памяти для тестов.
//
// Функция выполняет:
//   - Поиск глобальной директории с тестовыми данными проекта
//   - Удаление предыдущего файла хранилища (если существует)
//   - Создание нового in-memory хранилища
//   - Загрузку данных из файла (если файл существует)
//
// Возвращает:
//   - *storage.InMemoryStorage: инициализированное хранилище
//   - error: ошибка, если возникла при создании хранилища
//
// Особенности:
//   - Использует глобальную тестовую директорию проекта
//   - Автоматически очищает предыдущее состояние хранилища
//   - Поддерживает загрузку данных из файла
//
// Пример использования:
//
//	st, err := InitializeInMemoryStorage()
//	if err != nil {
//	    log.Fatalf("Failed to initialize storage: %v", err)
//	}
//	defer st.Close()
func InitializeInMemoryStorage() (*storage.InMemoryStorage, error) {

	// Для общих testdata проекта
	globalData, err := GetGlobalTestDataPath()
	if err != nil {
		return nil, err
	}

	fileStorage := filepath.Join(globalData, "test.dat")

	_ = os.Remove(fileStorage)

	st, err := storage.NewInMemoryStorage(fileStorage)
	if err != nil {
		return nil, err
	}
	//st.Load(fileStorage)

	return st, nil

}
