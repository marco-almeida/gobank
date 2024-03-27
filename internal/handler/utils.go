package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/marco-almeida/gobank/internal"
)

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
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
	if !errors.As(err, &ierr) {
		resp.Error = "internal error"
	} else {
		switch ierr.Code() {
		case internal.ErrorCodeNotFound:
			status = http.StatusNotFound
		case internal.ErrorCodeInvalidArgument:
			status = http.StatusBadRequest
		case internal.ErrorCodeDuplicate:
			status = http.StatusConflict

			// var verrors validation.Errors
			// if errors.As(ierr, &verrors) {
			// 	resp.Validations = verrors
			// }
		case internal.ErrorCodeUnknown:
			fallthrough
		default:
			status = http.StatusInternalServerError
		}
	}

	WriteJSON(w, status, resp)
}
