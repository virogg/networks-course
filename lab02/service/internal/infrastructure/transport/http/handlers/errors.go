package handlers

import (
	"errors"
	"net/http"

	svcerr "github.com/virogg/networks-course/service/internal/service/errors"
	pkghttp "github.com/virogg/networks-course/service/pkg/http"
)

func HandleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svcerr.ErrNotFound):
		pkghttp.RespondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, svcerr.ErrAlreadyExists), errors.Is(err, svcerr.ErrNoAvailableCouriers):
		pkghttp.RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, svcerr.ErrValidation), errors.Is(err, svcerr.ErrInvalidInput):
		pkghttp.RespondError(w, http.StatusBadRequest, err.Error())
	default:
		pkghttp.RespondError(w, http.StatusInternalServerError, "Internal server error")
	}
}
