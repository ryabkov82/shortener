package main

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestSetupAnalyzers(t *testing.T) {
	analyzers := setupAnalyzers()

	if len(analyzers) == 0 {
		t.Error("expected non-empty analyzers list")
	}

	// Проверяем наличие конкретных анализаторов
	checkAnalyzerExists(t, analyzers, "noosexit")
	checkAnalyzerExists(t, analyzers, "bodyclose")
}

func checkAnalyzerExists(t *testing.T, analyzers []*analysis.Analyzer, name string) {
	for _, a := range analyzers {
		if a.Name == name {
			return
		}
	}
	t.Errorf("analyzer %q not found", name)
}
