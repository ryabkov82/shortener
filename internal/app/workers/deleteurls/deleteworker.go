package deleteurls

import (
	"context"
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
	taskChan    chan DeleteTask
	batchChan   chan map[string][]string
	workerCount int
	batchSize   int
	batchWindow time.Duration
	repo        Repository
}

func NewDeleteWorker(workerCount, batchSize int, batchWindow time.Duration, storage Repository) *DeleteWorker {
	return &DeleteWorker{
		taskChan:    make(chan DeleteTask, 10000),
		batchChan:   make(chan map[string][]string, 100),
		workerCount: workerCount,
		batchSize:   batchSize,
		batchWindow: batchWindow,
		repo:        storage,
	}
}

func (w *DeleteWorker) Start(ctx context.Context) {
	// Запускаем сборщик батчей
	go w.batchCollector(ctx)

	// Запускаем воркеров для обработки батчей
	for i := 0; i < w.workerCount; i++ {
		go w.batchProcessor(ctx)
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

func (w *DeleteWorker) batchCollector(ctx context.Context) {
	batch := make(map[string][]string)
	ticker := time.NewTicker(w.batchWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// При завершении отправляем оставшиеся задачи
			if len(batch) > 0 {
				w.batchChan <- batch
			}
			close(w.batchChan)
			return

		case task := <-w.taskChan:
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

func (w *DeleteWorker) batchProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case batch, ok := <-w.batchChan:
			if !ok {
				return
			}

			var wg sync.WaitGroup
			for userID, urls := range batch {
				wg.Add(1)
				go func(uid string, u []string) {
					defer wg.Done()
					if err := w.processUserBatch(uid, u); err != nil {
						log.Printf("Failed to process batch for user %s: %v", uid, err)
					}
				}(userID, urls)
			}
			wg.Wait()
		}
	}
}

func (w *DeleteWorker) processUserBatch(userID string, urls []string) error {
	// 1. Обновляем в БД
	if err := w.repo.BatchMarkAsDeleted(userID, urls); err != nil {
		return err
	}

	return nil
}
