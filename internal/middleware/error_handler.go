package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/marco-almeida/mybank/internal"
	"github.com/rs/zerolog/log"
)

// ErrorResponse represents a response containing an error message.
type ErrorResponse struct {
	Error       string            `json:"error"`
	Validations []ValidationError `json:"validations,omitempty"`
}

type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, ginErr := range c.Errors {
			logLevel := log.Info()

			unwrappedErr := ginErr.Err
			var validationErrors validator.ValidationErrors
			switch {
			// check if error has validator.ValidationErrors
			case errors.As(unwrappedErr, &validationErrors):
				errorResponse := ErrorResponse{
					Error: http.StatusText(http.StatusBadRequest),
				}
				for e := range validationErrors {
					errorResponse.Validations = append(errorResponse.Validations, ValidationError{
						// field should be json representation of field
						Field: validationErrors[e].Field(),
						Tag:   validationErrors[e].Tag(),
					})
				}
				c.JSON(http.StatusBadRequest, errorResponse)
			case errors.Is(unwrappedErr, internal.ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			case errors.Is(unwrappedErr, internal.ErrInvalidToken):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			case errors.Is(unwrappedErr, internal.ErrForeignKeyConstraintViolation):
				c.JSON(http.StatusConflict, gin.H{"error": http.StatusText(http.StatusConflict)})
			case errors.Is(unwrappedErr, internal.ErrUniqueConstraintViolation):
				c.JSON(http.StatusConflict, gin.H{"error": http.StatusText(http.StatusConflict)})
			case errors.Is(unwrappedErr, internal.ErrNoRows):
				c.JSON(http.StatusNotFound, gin.H{"error": http.StatusText(http.StatusNotFound)})
			default:
				logLevel = log.Error()
				c.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			}
			logLevel.Err(unwrappedErr).Send()
		}
	}
}
