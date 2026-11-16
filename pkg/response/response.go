// Package response keeps one JSON envelope for every endpoint, so clients never
// have to special-case success and failure shapes.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the only shape this service writes.
type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error *Error `json:"error,omitempty"`
}

// Error is machine-readable: code is stable, message is for humans, fields is
// for form validation.
type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// OK writes a success envelope.
func OK(c *gin.Context, status int, data any) {
	c.JSON(status, Envelope{Data: data})
}

// Created writes 201 with the created resource.
func Created(c *gin.Context, data any) {
	OK(c, http.StatusCreated, data)
}

// Fail writes an error envelope and aborts the handler chain.
func Fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, Envelope{Error: &Error{Code: code, Message: message}})
}

// Invalid writes a 422 with per-field messages.
func Invalid(c *gin.Context, fields map[string]string) {
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, Envelope{
		Error: &Error{
			Code:    "validation_failed",
			Message: "one or more fields are invalid",
			Fields:  fields,
		},
	})
}
