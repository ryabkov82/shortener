// Package staticlint представляет собой комплексный статический анализатор для проекта Shortener.
// Он объединяет:
// - стандартные анализаторы go/analysis
// - анализаторы staticcheck
// - дополнительные сторонние анализаторы
//
// Для запуска:
//
//	go run cmd/staticlint/main.go ./...
//
// Для установки и использования как standalone-инструмента:
//
//	go install github.com/ryabkov82/shortener/cmd/staticlint
//	staticlint ./...
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/timakin/bodyclose/passes/bodyclose"

	"github.com/ryabkov82/shortener/cmd/staticlint/noosexit"
)

func main() {
	analyzers := setupAnalyzers()
	multichecker.Main(analyzers...)
}

func setupAnalyzers() []*analysis.Analyzer {

	// Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	standardAnalyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,          // проверяет корректность объявлений ассемблерного кода
		assign.Analyzer,           // обнаруживает бесполезные присваивания
		atomic.Analyzer,           // проверяет правильность использования sync/atomic
		bools.Analyzer,            // обнаруживает распространенные ошибки с булевыми операторами
		buildtag.Analyzer,         // проверяет корректность build тегов
		cgocall.Analyzer,          // проверяет корректность вызовов CGO
		composite.Analyzer,        // проверяет композитные литералы без ключей
		copylock.Analyzer,         // проверяет копирование мьютексов
		deepequalerrors.Analyzer,  // проверяет использование deep equal с ошибками
		errorsas.Analyzer,         // проверяет правильность использования errors.As
		fieldalignment.Analyzer,   // предлагает оптимальное выравнивание полей структур
		httpresponse.Analyzer,     // проверяет закрытие HTTP response bodies
		ifaceassert.Analyzer,      // обнаруживает бессмысленные type assertions
		loopclosure.Analyzer,      // проверяет захват переменных в замыканиях
		lostcancel.Analyzer,       // проверяет утечку контекста
		nilfunc.Analyzer,          // обнаруживает сравнения функций с nil
		printf.Analyzer,           // проверяет формат строки в Printf-функциях
		shadow.Analyzer,           // обнаруживает затенение переменных
		shift.Analyzer,            // проверяет сдвиги превышающие размер типа
		sigchanyzer.Analyzer,      // проверяет неправильное использование каналов в signal.Notify
		stdmethods.Analyzer,       // проверяет соответствие стандартным интерфейсам
		structtag.Analyzer,        // проверяет корректность тегов структур
		testinggoroutine.Analyzer, // обнаруживает утечку горутин в тестах
		unmarshal.Analyzer,        // проверяет правильность передачи указателей в Unmarshal
		unreachable.Analyzer,      // обнаруживает недостижимый код
		unsafeptr.Analyzer,        // проверяет корректность преобразований unsafe.Pointer
		unusedresult.Analyzer,     // проверяет неиспользованные результаты функций
	}

	// Анализаторы класса SA из staticcheck.io (Static Analysis)
	var saAnalyzers []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[:2] == "SA" {
			saAnalyzers = append(saAnalyzers, v.Analyzer)
		}
	}

	// Дополнительные анализаторы из других классов staticcheck
	var otherStaticcheckAnalyzers []*analysis.Analyzer

	// Добавляем анализаторы из stylecheck
	for _, a := range stylecheck.Analyzers {
		if a.Analyzer.Name == "ST1000" { // проверка документации пакета
			otherStaticcheckAnalyzers = append(otherStaticcheckAnalyzers, a.Analyzer)
			break
		}
	}

	// Добавляем анализаторы из simple
	for _, a := range simple.Analyzers {
		if a.Analyzer.Name == "S1002" { // предлагает упрощение булевых выражений
			otherStaticcheckAnalyzers = append(otherStaticcheckAnalyzers, a.Analyzer)
			break
		}
	}

	// Добавляем анализаторы из quickfix
	for _, a := range quickfix.Analyzers {
		if a.Analyzer.Name == "QF1001" { // применяет законы Де Моргана
			otherStaticcheckAnalyzers = append(otherStaticcheckAnalyzers, a.Analyzer)
			break
		}
	}

	// Сторонние анализаторы
	externalAnalyzers := []*analysis.Analyzer{
		bodyclose.Analyzer, // проверяет закрытие response.Body
		analyzer.Analyzer,  // go-critic, выявляет потенциальные ошибки, неэффективности и плохие практики программирования
	}

	// Собственные анализаторы
	customAnalyzers := []*analysis.Analyzer{
		noosexit.NoOsExitAnalyzer,
	}

	// Объединяем все анализаторы
	var analyzers []*analysis.Analyzer
	analyzers = append(analyzers, standardAnalyzers...)
	analyzers = append(analyzers, saAnalyzers...)
	analyzers = append(analyzers, otherStaticcheckAnalyzers...)
	analyzers = append(analyzers, externalAnalyzers...)
	analyzers = append(analyzers, customAnalyzers...)

	return analyzers
}
