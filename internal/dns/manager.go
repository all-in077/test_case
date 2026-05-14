package dns

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	ErrAlreadyExists = errors.New("DNS server already exists")
	ErrNotFound      = errors.New("DNS server not found")
)

type Manager struct {
	filePath string
	mu       sync.Mutex // Мьютекс для конкуретного доступа к файлу
}

// Констурктор
func NewManager(filePath string) *Manager {
	return &Manager{filePath: filePath}
}

// Метод чтения днс-ов из файла, возвращает слайс строк
func (m *Manager) readServers() ([]string, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // Если нет файла - просто пустой список вернем
		}
		return nil, err
	}

	var servers []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver ") { // Если есть нужный префикс заходим в строку и вычлиняем dns
			ip := strings.TrimPrefix(line, "nameserver ")
			servers = append(servers, strings.TrimSpace(ip))
		}
	}
	return servers, nil
}

// Метод обертка над readServers, дополнтильно обезопасли мьютексом конкурентный доступ
func (m *Manager) List() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.readServers()
}

// Метод записи в файл нового днс-а
func (m *Manager) Add(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	servers, err := m.readServers()
	if err != nil {
		return err
	}

	// Проверяем дубликат
	for _, s := range servers {
		if s == ip {
			return ErrAlreadyExists
		}
	}

	// Дописываем в файл, если не нашли дубликат
	f, err := os.OpenFile(m.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "nameserver %s\n", ip)
	return err
}

// Метод удаление днс-а
func (m *Manager) Remove(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	var newLines []string
	found := false

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "nameserver "+ip {
			found = true
			continue // Пропускаем эту строку, если нашли искомый IP
		}
		newLines = append(newLines, line)
	}

	if !found {
		return ErrNotFound
	}

	return os.WriteFile(m.filePath, []byte(strings.Join(newLines, "\n")), 0644)
}
