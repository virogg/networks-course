package handlers

import (
	"net/http"

	"github.com/virogg/networks-course/lab06/solution/static"
)

func Index(w http.ResponseWriter, r *http.Request) {
	data, _ := static.FS.ReadFile("index.html")
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(data)
}
