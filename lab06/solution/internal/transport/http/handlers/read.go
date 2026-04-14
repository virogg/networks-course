package handlers

import (
	"net/http"

	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) Read(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.requireConn(w) {
		return
	}

	name := r.URL.Query().Get("name")
	data, err := h.client.Get(name)
	if err != nil {
		pkghttp.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkghttp.RespondJSON(w, map[string]string{"content": string(data)})
}
