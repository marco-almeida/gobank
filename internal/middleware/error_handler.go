package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/marco-almeida/mybank/internal"
	"github.com/rs/zerolog/log"
)

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
				errorResponse := internal.ErrorResponse{
					Error: http.StatusText(http.StatusBadRequest),
				}
				for e := range validationErrors {
					errorResponse.Validations = append(errorResponse.Validations, internal.ValidationError{
						// field should be json representation of field
						Field:   validationErrors[e].Field(),
						Tag:     validationErrors[e].Tag(),
						Message: validationErrorToText(validationErrors[e]),
					})
				}
				c.JSON(http.StatusBadRequest, errorResponse)
			case errors.Is(unwrappedErr, internal.ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			case errors.Is(unwrappedErr, internal.ErrAccountAlreadyExists):
				c.JSON(http.StatusBadRequest, gin.H{"error": "account already exists"})
			case errors.Is(unwrappedErr, internal.ErrInvalidToken):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			case errors.Is(unwrappedErr, internal.ErrCurrencyMismatch):
				c.JSON(http.StatusBadRequest, gin.H{"error": "currency mismatch"})
			case errors.Is(unwrappedErr, internal.ErrForbidden):
				c.JSON(http.StatusForbidden, gin.H{"error": http.StatusText(http.StatusForbidden)})
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

func validationErrorToText(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", e.Field(), e.Param())
	case "min":
		return fmt.Sprintf("%s must be longer than %s", e.Field(), e.Param())
	case "email":
		return fmt.Sprintf("Invalid email format")
	case "len":
		return fmt.Sprintf("%s must be %s characters long", e.Field(), e.Param())
	}
	return fmt.Sprintf("%s is not valid", e.Field())
}
