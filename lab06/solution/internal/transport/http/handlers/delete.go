package handlers

import (
	"net/http"

	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.requireConn(w) {
		return
	}
	name := r.URL.Query().Get("name")
	code, msg, _ := h.client.Cmd("DELE " + name)
	if code != ftp.StatusFileActionOK {
		pkghttp.RespondError(w, msg, http.StatusInternalServerError)
		return
	}
	pkghttp.RespondJSON(w, map[string]string{"status": "deleted"})
}
