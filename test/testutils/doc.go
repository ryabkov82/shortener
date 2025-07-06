// Package testutils предоставляет утилиты для тестирования сервиса сокращения URL.
//
// Пакет включает следующие основные возможности:
// - Работа с путями проекта и тестовыми данными
// - Инициализация тестовых хранилищ
// - Генерация тестовых аутентификационных данных
//
// # Работа с файловой системой
//
//	// Получение корня модуля
//	root, err := testutil.GetModuleRoot()
//
//	// Получение пути к общей папке testdata
//	testDataPath, err := testutil.GetGlobalTestDataPath()
//
// # Тестовые хранилища
//
//	// Инициализация InMemory хранилища
//	storage, err := testutil.InitializeInMemoryStorage()
//	defer os.Remove(storage.FilePath()) // Очистка
//
// # Аутентификация
//
//	// Генерация подписанной куки
//	cookie, userID := testutil.CreateSignedCookie()
//
// # Пример использования
//
//	func TestShortener(t *testing.T) {
//	    // Инициализация хранилища
//	    storage, err := testutil.InitializeInMemoryStorage()
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    defer os.Remove(storage.FilePath())
//
//	    // Создание тестового сервера
//	    srv := httptest.NewServer(handler.New(storage))
//	    defer srv.Close()
//
//	    // Создание авторизованного запроса
//	    cookie, _ := testutil.CreateSignedCookie()
//	    req := httptest.NewRequest("POST", srv.URL+"/api", nil)
//	    req.AddCookie(cookie)
//	}
//
// # Особенности
//
// - Все временные файлы должны очищаться в defer
// - Функции, возвращающие error, должны обрабатываться в тестах
// - Глобальные пути кешируются после первого вызова
package testutils
