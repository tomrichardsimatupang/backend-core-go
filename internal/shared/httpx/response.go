package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
	Param string `json:"param,omitempty"`
}

type dataResponse struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

type errorResponse struct {
	Error ErrorBody `json:"error"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, dataResponse{Data: data})
}

func OKWithMeta(c *gin.Context, data, meta any) {
	c.JSON(http.StatusOK, dataResponse{Data: data, Meta: meta})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, dataResponse{Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Fail menulis response error dan menghentikan handler chain.
func Fail(c *gin.Context, status int, code, message string) {
	failWithDetails(c, status, code, message, nil)
}

// BadRequest dipakai untuk error hasil bind/validasi request.
func BadRequest(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make([]FieldError, 0, len(ve))
		for _, fe := range ve {
			details = append(details, FieldError{Field: fe.Field(), Rule: fe.Tag(), Param: fe.Param()})
		}
		failWithDetails(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "request validation failed", details)
		return
	}
	Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
}

func Unauthorized(c *gin.Context) {
	Fail(c, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required")
}

func failWithDetails(c *gin.Context, status int, code, message string, details any) {
	c.AbortWithStatusJSON(status, errorResponse{Error: ErrorBody{Code: code, Message: message, Details: details}})
}
