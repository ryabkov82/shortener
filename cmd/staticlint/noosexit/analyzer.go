// Package noosexit проверяет отсутствие прямых вызовов os.Exit в функции main пакета main
package noosexit

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = `noosexit проверяет отсутствие прямых вызовов os.Exit в функции main пакета main

Анализатор запрещает использование os.Exit() в функции main() основного пакета,
рекомендуя вместо этого возвращать ошибки или использовать log.Fatal().`

// NoOsExitAnalyzer анализатор для проверки вызовов os.Exit
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	// Проверяем, что пакет находится в директории cmd
	if !isCmdPackage(pass.Pkg.Path()) {
		return nil, nil
	}

	// Используем inspector для более эффективного обхода AST
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Фильтруем только вызовы функций
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	// Проверяем, находимся ли мы в пакете main и функции main
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// Проверяем что это вызов os.Exit
		if ident, ok := fun.X.(*ast.Ident); ok {
			if ident.Name == "os" && fun.Sel.Name == "Exit" {
				// Проверяем что находимся в функции main пакета main
				if pass.Pkg.Name() == "main" {
					// Проверяем что находимся внутри функции main
					for _, f := range pass.Files {
						for _, decl := range f.Decls {
							if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name.Name == "main" {
								pass.Reportf(call.Pos(), "прямой вызов os.Exit в функции main запрещен, используйте log.Fatal() или возврат ошибки")
							}
						}
					}
				}
			}
		}
	})

	return nil, nil
}

// isCmdPackage проверяет, находится ли пакет в директории cmd
var isCmdPackage = func(pkgPath string) bool {
	return strings.Contains(filepath.ToSlash(pkgPath), "/cmd/")
}
