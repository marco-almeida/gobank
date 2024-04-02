package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/marco-almeida/gobank/internal"
)

type ErrorResponse struct {
	Error       string         `json:"error,omitempty"`
	Validations map[string]any `json:"validations,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, msg string, err error) {
	resp := ErrorResponse{Error: msg}
	status := http.StatusInternalServerError

	var ierr *internal.Error
	if errors.As(err, &ierr) {
		switch ierr.Code() {
		case internal.ErrorCodeNotFound:
			status = http.StatusNotFound
		case internal.ErrorCodeInvalidArgument:
			status = http.StatusBadRequest
			var verrors validator.ValidationErrors
			resp.Validations = make(map[string]any)
			if errors.As(ierr, &verrors) {
				for _, e := range verrors {
					resp.Validations[e.Field()] = e.Error()
				}
			}
		case internal.ErrorCodeDuplicate:
			status = http.StatusConflict
		case internal.ErrorCodeUnauthorized:
			status = http.StatusUnauthorized
		case internal.ErrorCodeUnknown:
			fallthrough
		default:
			status = http.StatusInternalServerError
		}
	}

	WriteJSON(w, status, resp)
}
