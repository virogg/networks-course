package products

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/virogg/networks-course/service/internal/infrastructure/transport/dto"
	"github.com/virogg/networks-course/service/internal/infrastructure/transport/http/handlers"
	"github.com/virogg/networks-course/service/internal/infrastructure/transport/mappers"
	pkghttp "github.com/virogg/networks-course/service/pkg/http"
	"github.com/virogg/networks-course/service/pkg/logger"
)

type Handler struct {
	service productsService
	log     logger.Logger
}

func NewHandler(service productsService, log logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkghttp.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		pkghttp.RespondError(w, http.StatusBadRequest, "name is required")
		return
	}

	input := mappers.ToCreateProductInput(req)
	product, err := h.service.CreateProduct(r.Context(), input)
	if err != nil {
		h.log.Error("creating product: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	pkghttp.RespondJSON(w, http.StatusCreated, mappers.ToProductResponse(product))
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("updating product: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "Invalid courier ID")
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("updating product: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	input := mappers.ToUpdateProductInput(id, req)
	product, err := h.service.UpdateProduct(r.Context(), input)
	if err != nil {
		h.log.Error("updating product: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	pkghttp.RespondJSON(w, http.StatusOK, mappers.ToProductResponse(product))
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("getting product: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "Invalid courier ID")
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		h.log.Error("getting product: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	pkghttp.RespondJSON(w, http.StatusOK, mappers.ToProductResponse(product))
}

func (h *Handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetProducts(r.Context())
	if err != nil {
		h.log.Error("getting all products: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	response := make([]dto.ProductResponse, 0, len(products))
	for _, product := range products {
		response = append(response, mappers.ToProductResponse(product))
	}

	pkghttp.RespondJSON(w, http.StatusOK, response)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("deleting product: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "Invalid courier ID")
		return
	}

	product, err := h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		h.log.Error("deleting product: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	pkghttp.RespondJSON(w, http.StatusOK, mappers.ToProductResponse(product))
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("uploading image: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	file, fileHeader, err := r.FormFile("icon")
	if err != nil {
		h.log.Error("uploading image: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, fmt.Sprintf("reading file of 'icon' form data. Reason: %s\n", err))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		h.log.Error("uploading image: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	product, err := h.service.UploadProductImage(r.Context(), id, data, fileHeader.Filename)
	if err != nil {
		h.log.Error("uploading image: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	pkghttp.RespondJSON(w, http.StatusOK, mappers.ToProductResponse(product))
}

func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.log.Error("getting image: %w", logger.NewField("error", err))
		pkghttp.RespondError(w, http.StatusBadRequest, "invalid product ID")
		return
	}

	data, ct, err := h.service.GetProductImage(r.Context(), id)
	if err != nil {
		h.log.Error("getting image: %w", logger.NewField("error", err))
		handlers.HandleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", fmt.Sprintf("%d", id)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
