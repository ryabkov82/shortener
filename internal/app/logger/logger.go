// Package logger предоставляет централизованную систему логирования для приложения
// на основе zap.Logger. Реализует паттерн синглтона для глобального доступа к логеру.
package logger

import (
	"go.uber.org/zap"
)

// Log - глобальный экземпляр логера, инициализированный no-op логером по умолчанию.
// No-op логер не производит никакого вывода и не аллоцирует ресурсы.
var Log *zap.Logger = zap.NewNop()

// Initialize настраивает глобальный логер с указанным уровнем логирования.
//
// Параметры:
//   - level: строка, определяющая уровень логирования (debug, info, warn, error, dpanic, panic, fatal)
//
// Возвращает:
//   - error: ошибка, если передан некорректный уровень логирования или возникла проблема при создании логера
//
// Пример использования:
//
//	err := logger.Initialize("debug")
//	if err != nil {
//	    // обработка ошибки инициализации
//	}
//	logger.Log.Info("Логер успешно инициализирован")
func Initialize(level string) error {
	// Преобразование строкового уровня в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// Конфигурация логера в production-стиле (JSON-формат, stacktrace для ошибок)
	cfg := zap.NewProductionConfig()

	// Установка уровня логирования
	cfg.Level = lvl

	// Создание логера на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	// Замена глобального логера
	Log = zl
	return nil
}
