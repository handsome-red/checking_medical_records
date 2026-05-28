// internal/handlers/book_handler.go
package handlers

import (
	"net/http"

	"med_book/internal/middleware"
	"med_book/internal/service"
	"med_book/internal/templates"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookHandler struct {
	bookService *service.BookService
	template    *templates.TemplatesManager
}

func NewBookHandler(
	bookService *service.BookService,
	template *templates.TemplatesManager,
) *BookHandler {
	return &BookHandler{
		bookService: bookService,
		template:    template,
	}
}

// GetAllBooks возвращает список всех книг
func (h *BookHandler) GetAllBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.bookService.GetAllBooks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"books": books,
	}
	
	if err := h.template.ExecuteTemplate(w, "books.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// GetBookByID возвращает детали одной книги
func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	book, err := h.bookService.GetBookByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	data := map[string]any{
		"book": book,
	}
	
	if err := h.template.ExecuteTemplate(w, "book.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// StartTestPage показывает страницу выбора книги для теста
func (h *BookHandler) StartTestPage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	books, err := h.bookService.GetAllBooks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"books":   books,
		"user_id": userID,
	}
	
	if err := h.template.ExecuteTemplate(w, "select_book.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}