package noosexit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func TestNoOsExitAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		pkgPath  string
		wantMsg  string
		wantDiag bool
	}{
		{
			name:     "os.Exit in main package",
			code:     `package main; import "os"; func main() { os.Exit(1) }`,
			pkgPath:  "project/cmd/main",
			wantMsg:  "прямой вызов os.Exit в функции main запрещен",
			wantDiag: true,
		},
		{
			name:     "other package exit",
			code:     `package main; func main() { other.Exit() }`,
			pkgPath:  "project/cmd/main",
			wantDiag: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настройка тестового окружения
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Создаем минимальную информацию о типах
			info := &types.Info{
				Uses: make(map[*ast.Ident]types.Object),
				Defs: make(map[*ast.Ident]types.Object),
			}

			// Конфигурация для проверки типов
			conf := types.Config{
				Importer: new(importerStub),
			}

			// Создаем тестовый пакет
			pkg := types.NewPackage(tt.pkgPath, "main")

			// Проверяем типы (игнорируем ошибки, так как нам важно только использование os.Exit)
			_, _ = conf.Check(tt.pkgPath, fset, []*ast.File{file}, info)

			// Создаем тестовый analysis.Pass
			pass := &analysis.Pass{
				Fset:      fset,
				Files:     []*ast.File{file},
				Pkg:       pkg,
				TypesInfo: info,
				ResultOf: map[*analysis.Analyzer]interface{}{
					inspect.Analyzer: inspector.New([]*ast.File{file}),
				},
			}

			// Мокаем isCmdPackage
			oldIsCmdPackage := isCmdPackage
			isCmdPackage = func(pkgPath string) bool {
				return strings.Contains(pkgPath, "/cmd/")
			}
			defer func() { isCmdPackage = oldIsCmdPackage }()

			// Запускаем анализатор и собираем диагностики
			var diags []analysis.Diagnostic
			pass.Report = func(d analysis.Diagnostic) {
				diags = append(diags, d)
			}

			_, err = run(pass)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			// Проверяем результаты
			if tt.wantDiag {
				if len(diags) == 0 {
					t.Error("Expected diagnostic, got none")
				} else if !strings.Contains(diags[0].Message, tt.wantMsg) {
					t.Errorf("Got diagnostic %q, want %q", diags[0].Message, tt.wantMsg)
				}
			} else if len(diags) > 0 {
				t.Errorf("Unexpected diagnostic: %v", diags[0])
			}
		})
	}
}

// Простейший импортер для тестов
type importerStub struct{}

func (i *importerStub) Import(path string) (*types.Package, error) {
	return types.NewPackage(path, path), nil
}
