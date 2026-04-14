package handlers

import "net/http"

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.doStore(w, r)
}
