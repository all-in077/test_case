package dns

import (
	"os"
	"path/filepath"
	"testing"
)

// Вспомогательная функция (создаёт временный файл)
func newTestManager(t *testing.T, content string) *Manager {
	t.Helper()
	dir := t.TempDir() // Go сам удалит папку после теста
	path := filepath.Join(dir, "resolv.conf")

	if content != "" {
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	return NewManager(path)
}

// Тестыы дл List

func TestList_EmptyFile(t *testing.T) {
	m := newTestManager(t, "")

	servers, err := m.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("expected empty list, got %v", servers)
	}
}

func TestList_ReturnsOnlyNameservers(t *testing.T) {
	m := newTestManager(t, `# comment
search local
nameserver 8.8.8.8
nameserver 1.1.1.1
`)

	servers, err := m.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0] != "8.8.8.8" || servers[1] != "1.1.1.1" {
		t.Fatalf("unexpected servers: %v", servers)
	}
}

// Тесты для Add

func TestAdd_Success(t *testing.T) {
	m := newTestManager(t, "")

	if err := m.Add("8.8.8.8"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	servers, _ := m.List()
	if len(servers) != 1 || servers[0] != "8.8.8.8" {
		t.Fatalf("expected [8.8.8.8], got %v", servers)
	}
}

func TestAdd_Duplicate(t *testing.T) {
	m := newTestManager(t, "nameserver 8.8.8.8\n")

	err := m.Add("8.8.8.8")
	if err != ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAdd_PreservesOtherLines(t *testing.T) {
	m := newTestManager(t, "# comment\nsearch local\n")

	m.Add("8.8.8.8")

	// читаем файл напрямую
	data, _ := os.ReadFile(m.filePath)
	content := string(data)

	if content[:9] != "# comment" {
		t.Fatal("comment line was lost after Add")
	}
}

// Тесты для Remove

func TestRemove_Success(t *testing.T) {
	m := newTestManager(t, "nameserver 8.8.8.8\nnameserver 1.1.1.1\n")

	if err := m.Remove("8.8.8.8"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	servers, _ := m.List()
	if len(servers) != 1 || servers[0] != "1.1.1.1" {
		t.Fatalf("expected [1.1.1.1], got %v", servers)
	}
}

func TestRemove_NotFound(t *testing.T) {
	m := newTestManager(t, "nameserver 1.1.1.1\n")

	err := m.Remove("8.8.8.8")
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRemove_PreservesOtherLines(t *testing.T) {
	m := newTestManager(t, "# comment\nsearch local\nnameserver 8.8.8.8\n")

	m.Remove("8.8.8.8")

	servers, _ := m.List()
	if len(servers) != 0 {
		t.Fatalf("expected empty list, got %v", servers)
	}

	// Проверяем что comment b search остались
	data, _ := os.ReadFile(m.filePath)
	content := string(data)
	if !contains(content, "# comment") || !contains(content, "search local") {
		t.Fatal("other lines were lost after Remove")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
