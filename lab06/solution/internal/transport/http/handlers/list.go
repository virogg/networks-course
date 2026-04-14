package handlers

import (
	"net/http"
	"strings"

	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.requireConn(w) {
		return
	}

	data, err := h.client.List()
	if err != nil {
		pkghttp.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimRight(line, "\r")
		if line != "" {
			files = append(files, line)
		}
	}
	if files == nil {
		files = []string{}
	}

	pkghttp.RespondJSON(w, map[string]any{"files": files})
}
