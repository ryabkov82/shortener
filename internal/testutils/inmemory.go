package testutils

import (
	"os"
	"path/filepath"

	storage "github.com/ryabkov82/shortener/internal/app/storage/inmemory"
)

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
	st.Load(fileStorage)

	return st, nil

}
