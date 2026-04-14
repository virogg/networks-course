package main

import (
	"log"
	"net/http"

	transport "github.com/virogg/networks-course/lab06/solution/internal/transport/http"
	"github.com/virogg/networks-course/lab06/solution/internal/transport/http/handlers"
)

func main() {
	h := new(handlers.Handler)
	r := transport.NewRouter(h)
	log.Println("GUI FTP client at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
