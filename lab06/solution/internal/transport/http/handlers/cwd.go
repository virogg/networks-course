package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) Cwd(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.requireConn(w) {
		return
	}

	var req struct {
		Dir string `json:"dir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkghttp.RespondError(w, "invalid request", http.StatusBadRequest)
		return
	}

	code, msg, _ := h.client.Cmd("CWD " + req.Dir)
	if code != ftp.StatusFileActionOK {
		pkghttp.RespondError(w, msg, http.StatusInternalServerError)
		return
	}
	pkghttp.RespondJSON(w, map[string]string{"status": "ok"})
}
