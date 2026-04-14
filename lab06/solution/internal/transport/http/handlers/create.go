package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/virogg/networks-course/lab06/solution/internal/transport/dto"
	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.doStore(w, r)
}

func (h *Handler) doStore(w http.ResponseWriter, r *http.Request) {
	if !h.requireConn(w) {
		return
	}

	var req dto.FileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkghttp.RespondError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.client.Put(req.Name, []byte(req.Content)); err != nil {
		pkghttp.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkghttp.RespondJSON(w, map[string]string{"status": "ok"})
}
