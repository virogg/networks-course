package handlers

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/virogg/networks-course/lab06/solution/internal/transport/dto"
	"github.com/virogg/networks-course/lab06/solution/internal/transport/ftp/client"
	"github.com/virogg/networks-course/lab06/solution/pkg/ftp"
	pkghttp "github.com/virogg/networks-course/lab06/solution/pkg/http"
)

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var req dto.ConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkghttp.RespondError(w, "invalid request", http.StatusBadRequest)
		return
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(req.Host, req.Port))
	if err != nil {
		pkghttp.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := client.New(conn)

	if code, _, _ := c.ReadResponse(); code != ftp.StatusServiceReady {
		conn.Close()
		pkghttp.RespondError(w, "server not ready", http.StatusBadGateway)
		return
	}

	c.Cmd("USER " + req.User)
	code, msg, _ := c.Cmd("PASS " + req.Pass)
	if code != ftp.StatusLoginSuccessful {
		conn.Close()
		pkghttp.RespondError(w, "login failed: "+msg, http.StatusUnauthorized)
		return
	}

	h.client = c
	pkghttp.RespondJSON(w, map[string]string{"status": "connected"})
}
