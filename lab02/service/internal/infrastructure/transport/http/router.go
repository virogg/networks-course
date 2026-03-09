package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/virogg/networks-course/service/internal/infrastructure/transport/http/handlers/products"
	"github.com/virogg/networks-course/service/pkg/logger"
)

func NewRouter(log logger.Logger, productService productsService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	if productService != nil {
		h := products.NewHandler(productService, log)

		r.Post("/product", h.CreateProduct)
		r.Get("/product/{id}", h.GetProduct)
		r.Put("/product/{id}", h.UpdateProduct)
		r.Delete("/product/{id}", h.DeleteProduct)
		r.Get("/products", h.GetAllProducts)
		r.Post("/product/{id}/image", h.UploadImage)
		r.Get("/product/{id}/image", h.GetImage)
	}

	return r
}
