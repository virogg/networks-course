package handlers

import (
	"net/http"
	"sync"

	"github.com/virogg/networks-course/lab06/solution/internal/transport/ftp/client"
	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

type Handler struct {
	mu     sync.Mutex
	client *client.Client
}

func (h *Handler) requireConn(w http.ResponseWriter) bool {
	if h.client == nil {
		pkghttp.RespondError(w, "not connected", http.StatusBadRequest)
		return false
	}
	return true
}
