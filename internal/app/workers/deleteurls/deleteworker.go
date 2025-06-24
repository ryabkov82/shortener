// Package deleteurls предоставляет асинхронный обработчик для пакетного удаления URL в сервисе сокращения ссылок.
//
// Пакет реализует паттерн "рабочий пул" с:
// - Группировкой запросов по пользователям
// - Настраиваемым размером пакета и временем ожидания
// - Параллельной обработкой с помощью пула воркеров
// - Поддержкой плавного завершения работы
package deleteurls

import (
	"errors"
	"log"
	"sync"
	"time"
)

// Repository определяет интерфейс хранилища, необходимый для работы DeleteWorker.
type Repository interface {
	// BatchMarkAsDeleted помечает несколько URL как удаленные для указанного пользователя.
	// Возвращает ошибку в случае неудачи.
	BatchMarkAsDeleted(userID string, urls []string) error
}

// DeleteTask представляет запрос на удаление нескольких сокращенных URL для пользователя.
type DeleteTask struct {
	UserID    string   // ID пользователя, инициировавшего запрос
	ShortURLs []string // Список сокращенных URL для пометки как удаленных
}

// DeleteWorker управляет жизненным циклом обработки удаления URL.
// Агрегирует запросы в пакеты и обрабатывает их асинхронно.
type DeleteWorker struct {
	repo        Repository
	taskChan    chan DeleteTask
	batchChan   chan map[string][]string
	stopChan    chan struct{}
	wg          sync.WaitGroup
	workerCount int
	batchSize   int
	batchWindow time.Duration
}

// NewDeleteWorker создает новый экземпляр DeleteWorker с заданными параметрами.
//
// Параметры:
//   - workerCount: количество воркеров для обработки пакетов
//   - batchSize: максимальный размер пакета перед обработкой
//   - batchWindow: максимальное время ожидания формирования пакета
//   - storage: реализация интерфейса Repository
func NewDeleteWorker(workerCount, batchSize int, batchWindow time.Duration, storage Repository) *DeleteWorker {
	return &DeleteWorker{
		taskChan:    make(chan DeleteTask, 10000),
		batchChan:   make(chan map[string][]string, 100),
		stopChan:    make(chan struct{}),
		workerCount: workerCount,
		batchSize:   batchSize,
		batchWindow: batchWindow,
		repo:        storage,
	}
}

// Start запускает воркеры и сборщик пакетов.
func (w *DeleteWorker) Start() {
	w.wg.Add(w.workerCount + 1) // +1 для сборщика пакетов

	go w.batchCollector() // Запускаем сборщик пакетов

	for i := 0; i < w.workerCount; i++ {
		go w.batchProcessor() // Запускаем воркеры
	}
}

// Submit добавляет новую задачу на удаление в очередь обработки.
// Возвращает ошибку если очередь переполнена.
func (w *DeleteWorker) Submit(task DeleteTask) error {
	select {
	case w.taskChan <- task:
		return nil
	default:
		return errors.New("очередь переполнена")
	}
}

// batchCollector собирает задачи в пакеты по пользователям.
// Отправляет пакеты на обработку при достижении batchSize или по истечении batchWindow.
func (w *DeleteWorker) batchCollector() {
	defer w.wg.Done()

	batch := make(map[string][]string)
	ticker := time.NewTicker(w.batchWindow)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			if len(batch) > 0 {
				w.batchChan <- batch
			}
			close(w.batchChan)
			return

		case task, ok := <-w.taskChan:
			if !ok {
				if len(batch) > 0 {
					w.batchChan <- batch
				}
				close(w.batchChan)
				return
			}

			if urls, exists := batch[task.UserID]; exists {
				batch[task.UserID] = append(urls, task.ShortURLs...)
			} else {
				batch[task.UserID] = task.ShortURLs
			}

			if len(batch) >= w.batchSize {
				w.batchChan <- batch
				batch = make(map[string][]string)
				ticker.Reset(w.batchWindow)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				w.batchChan <- batch
				batch = make(map[string][]string)
			}
		}
	}
}

// batchProcessor обрабатывает пакеты задач, используя пул воркеров.
func (w *DeleteWorker) batchProcessor() {
	defer w.wg.Done()

	for batch := range w.batchChan {
		var batchWg sync.WaitGroup
		batchWg.Add(len(batch))

		concurrencyLimit := make(chan struct{}, w.workerCount*2)

		for userID, urls := range batch {
			concurrencyLimit <- struct{}{}

			go func(userID string, urls []string) {
				defer batchWg.Done()
				defer func() { <-concurrencyLimit }()

				const subBatchSize = 50
				for i := 0; i < len(urls); i += subBatchSize {
					end := i + subBatchSize
					if end > len(urls) {
						end = len(urls)
					}
					subBatch := urls[i:end]

					if err := w.processUserBatch(userID, subBatch); err != nil {
						log.Printf("Ошибка при пометке URL как удалённых для пользователя %s: %v", userID, err)
					}
				}
			}(userID, urls)
		}

		batchWg.Wait()
	}
}

// processUserBatch выполняет пометку URL как удаленных в хранилище.
func (w *DeleteWorker) processUserBatch(userID string, urls []string) error {
	if err := w.repo.BatchMarkAsDeleted(userID, urls); err != nil {
		return err
	}
	return nil
}

// GracefulStop выполняет плавное завершение работы с заданным таймаутом.
// Дожидается завершения обработки текущих задач или истечения таймаута.
func (w *DeleteWorker) GracefulStop(timeout time.Duration) {
	close(w.stopChan)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Все воркеры завершили работу")
	case <-time.After(timeout):
		log.Println("Таймаут ожидания завершения воркеров")
	}

	close(w.taskChan)
}
