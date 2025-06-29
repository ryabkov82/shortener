package testutils

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GetModuleRoot возвращает абсолютный путь к корневой директории Go-модуля.
// Использует команду 'go list -m'. Может возвращать ошибку, если команда
// не выполнена или модуль не инициализирован.
func GetModuleRoot() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetGlobalTestDataPath возвращает путь к общей папке testdata проекта
func GetGlobalTestDataPath() (string, error) {
	root, err := GetModuleRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "test", "testdata"), nil
}
