package deleteurls

import (
	"errors"
	"log"
	"sync"
	"time"
)

type Repository interface {
	BatchMarkAsDeleted(userID string, urls []string) error
}

type DeleteTask struct {
	UserID    string
	ShortURLs []string
}

type DeleteWorker struct {
	taskChan    chan DeleteTask          // канал задач на удаление сокращенных url
	batchChan   chan map[string][]string // агрерированные в батчи задачи на удаление сокращенных url в разрезе пользователей
	stopChan    chan struct{}            // канал завершения
	wg          sync.WaitGroup           // для ожидания завершения воркеров
	workerCount int
	batchSize   int
	batchWindow time.Duration
	repo        Repository
}

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

func (w *DeleteWorker) Start() {

	w.wg.Add(w.workerCount + 1) // +1 для сборщика батчей

	// Запускаем сборщик батчей
	go w.batchCollector()

	// Запускаем воркеров для обработки батчей
	for i := 0; i < w.workerCount; i++ {
		go w.batchProcessor()
	}
}

func (w *DeleteWorker) Submit(task DeleteTask) error {

	select {
	case w.taskChan <- task:
		return nil
	default:
		return errors.New("очередь переполнена") // Очередь переполнена
	}
}

func (w *DeleteWorker) batchCollector() {

	defer w.wg.Done()

	batch := make(map[string][]string)
	ticker := time.NewTicker(w.batchWindow)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			// При завершении отправляем оставшиеся задачи
			if len(batch) > 0 {
				w.batchChan <- batch
			}
			close(w.batchChan)
			return

		case task, ok := <-w.taskChan:

			if !ok {
				// Канал закрыт, отправляем оставшиеся данные
				if len(batch) > 0 {
					w.batchChan <- batch
				}
				close(w.batchChan)
				return
			}

			// Добавляем URL в батч для данного пользователя
			if urls, exists := batch[task.UserID]; exists {
				batch[task.UserID] = append(urls, task.ShortURLs...)
			} else {
				batch[task.UserID] = task.ShortURLs
			}

			// Если батч достиг размера - отправляем на обработку
			if len(batch) >= w.batchSize {
				w.batchChan <- batch
				batch = make(map[string][]string)
				ticker.Reset(w.batchWindow)
			}

		case <-ticker.C:
			// По таймеру отправляем собранные задачи
			if len(batch) > 0 {
				w.batchChan <- batch
				batch = make(map[string][]string)
			}
		}
	}
}

func (w *DeleteWorker) batchProcessor() {

	defer w.wg.Done()

	for batch := range w.batchChan {
		// Используем WaitGroup для ожидания завершения всех горутин
		var batchWg sync.WaitGroup
		batchWg.Add(len(batch))

		// Создаем канал для ограничения количества одновременно работающих горутин
		concurrencyLimit := make(chan struct{}, w.workerCount*2)

		for userID, urls := range batch {
			// Захватываем слот в канале (ограничиваем параллелизм)
			concurrencyLimit <- struct{}{}

			go func(userID string, urls []string) {
				defer batchWg.Done()
				defer func() { <-concurrencyLimit }() // Освобождаем слот

				// Разбиваем на под-батчи для очень больших списков URL
				const subBatchSize = 50
				for i := 0; i < len(urls); i += subBatchSize {
					end := i + subBatchSize
					if end > len(urls) {
						end = len(urls)
					}
					subBatch := urls[i:end]

					if err := w.processUserBatch(userID, subBatch); err != nil {
						log.Printf("Ошибка при пометке URL как удалённых для пользователя %s: %v", userID, err)
						// Можно добавить retry логику здесь при необходимости
					}
				}
			}(userID, urls)
		}

		// Ожидаем завершения обработки всего батча
		batchWg.Wait()
	}
}

func (w *DeleteWorker) processUserBatch(userID string, urls []string) error {
	// 1. Обновляем в БД
	if err := w.repo.BatchMarkAsDeleted(userID, urls); err != nil {
		return err
	}

	return nil
}

// GracefulStop реализует graceful shutdown с таймаутом
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
