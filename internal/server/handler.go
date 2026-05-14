package server

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"

	"test_case_yadro/internal/dns"
)

type Handler struct {
	manager *dns.Manager
}

func NewHandler(manager *dns.Manager) *Handler {
	return &Handler{manager: manager}
}

// Вспомгательная функция, чтобы избежать дублирования кода
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Хенлер на GET запрос от пользователя (получить список всех днс-ов)
func (h *Handler) listServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.manager.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to read DNS servers",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"servers": servers,
	})
}

// Хендлер для POST запроса пользователя (добавления нового днс-а)
func (h *Handler) addServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Валидация IP
	if net.ParseIP(req.IP) == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid IP address",
		})
		return
	}

	if err := h.manager.Add(req.IP); err != nil {
		if errors.Is(err, dns.ErrAlreadyExists) {
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
			return
		}
		log.Println("Error adding server:", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to add DNS server",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"ip": req.IP,
	})
}

// Хендлер для DELETE запроса от пользователя (удаление днс-а)
func (h *Handler) removeServer(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")

	// Валидация IP
	if net.ParseIP(ip) == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid IP address",
		})
		return
	}

	if err := h.manager.Remove(ip); err != nil {
		if errors.Is(err, dns.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to remove DNS server",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"ip": ip,
	})
}

// Регистрация роутов
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /dns", h.listServers)
	mux.HandleFunc("POST /dns", h.addServer)
	mux.HandleFunc("DELETE /dns/{ip}", h.removeServer)
}
