package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/virogg/networks-course/lab06/solution/internal/transport/http/handlers"
)

func NewRouter(h *handlers.Handler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", handlers.Index)
	r.Post("/connect", h.Connect)
	r.Post("/disconnect", h.Disconnect)
	r.Get("/list", h.List)
	r.Post("/cwd", h.Cwd)
	r.Post("/create", h.Create)
	r.Get("/read", h.Read)
	r.Put("/update", h.Update)
	r.Delete("/delete", h.Delete)
	return r
}
