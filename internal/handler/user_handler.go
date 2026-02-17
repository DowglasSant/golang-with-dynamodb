package handler

import (
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dowglassantana/golang-with-dynamodb/internal/entity"
	"github.com/dowglassantana/golang-with-dynamodb/internal/service"
)

//go:embed static/index.html
var indexHTML embed.FS

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, _ := indexHTML.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
	mux.HandleFunc("POST /users", h.Create)
	mux.HandleFunc("GET /users", h.GetAll)
	mux.HandleFunc("GET /users/{id}", h.GetByID)
	mux.HandleFunc("PUT /users/{id}", h.Update)
	mux.HandleFunc("DELETE /users/{id}", h.Delete)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input entity.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo da requisicao invalido"})
		return
	}

	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name e email sao obrigatorios"})
		return
	}

	user, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "usuario nao encontrado"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAll(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var input entity.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo da requisicao invalido"})
		return
	}

	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Email) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name e email sao obrigatorios"})
		return
	}

	if err := h.service.Update(r.Context(), id, input); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "usuario atualizado com sucesso"})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
