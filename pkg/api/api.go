package api

import (
	"GoNews/pkg/storage"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Программный интерфейс сервера GoNews
type API struct {
	db     storage.Interface
	router *mux.Router
}

type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// Конструктор объекта API
func New(db storage.Interface) *API {
	api := API{
		db: db,
	}
	api.router = mux.NewRouter()
	api.endpoints()
	return &api
}

// Регистрация обработчиков API.
func (api *API) endpoints() {
	api.router.HandleFunc("/posts", api.postsHandler).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/posts", api.addPostHandler).Methods(http.MethodPost, http.MethodOptions)
	api.router.HandleFunc("/posts", api.updatePostHandler).Methods(http.MethodPut, http.MethodOptions)
	api.router.HandleFunc("/posts", api.deletePostHandler).Methods(http.MethodDelete, http.MethodOptions)
}

// Получение маршрутизатора запросов.
// Требуется для передачи маршрутизатора веб-серверу.
func (api *API) Router() *mux.Router {
	return api.router
}

func (api *API) validatePost(p storage.Post) []string {
	var validationErrors []string

	if p.Title == "" {
		validationErrors = append(validationErrors, "заголовок публикации не может быть пустым")
	}
	if p.Content == "" {
		validationErrors = append(validationErrors, "содержание публикации не может быть пустым")
	}
	if p.AuthorID <= 0 {
		validationErrors = append(validationErrors, "ID автора должен быть положительным")
	}

	return validationErrors
}

func (api *API) validatePostUpdate(p storage.Post) []string {
	var validationErrors []string

	if p.ID <= 0 {
		validationErrors = append(validationErrors, "ID публикации не может быть пустым")
	}

	return validationErrors
}

// Получение всех публикаций.
func (api *API) postsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := api.db.Posts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// Добавление публикации.
func (api *API) addPostHandler(w http.ResponseWriter, r *http.Request) {
	var p storage.Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	validationErrors := api.validatePost(p)
	if len(validationErrors) > 0 {
		response := ErrorResponse{Errors: validationErrors}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = api.db.AddPost(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Обновление публикации.
func (api *API) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	var p storage.Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	validationErrors := api.validatePostUpdate(p)
	if len(validationErrors) > 0 {
		response := ErrorResponse{Errors: validationErrors}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = api.db.UpdatePost(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Удаление публикации.
func (api *API) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	var p storage.Post
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = api.db.DeletePost(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
