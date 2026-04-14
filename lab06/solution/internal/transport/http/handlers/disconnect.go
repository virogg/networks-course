package handlers

import (
	"net/http"

	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.client != nil {
		h.client.Close()
		h.client = nil
	}
	pkghttp.RespondJSON(w, map[string]string{"status": "disconnected"})
}
