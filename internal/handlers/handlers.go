package handlers

import (
	"net/http"

	// "github.com/gin-gonic/gin"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}