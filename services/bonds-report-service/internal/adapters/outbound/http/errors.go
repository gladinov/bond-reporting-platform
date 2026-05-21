package httperrors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrTooManyRequests     = errors.New("too many requests")
	ErrInternalServerError = errors.New("internal server error")
	ErrServiceUnavailable  = errors.New("service unavailable")
)

func MapHTTPError(status int, body []byte) error {
	switch status {
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusInternalServerError:
		return ErrInternalServerError
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return ErrServiceUnavailable
	default:
		return fmt.Errorf("unexpected http status %d: %s", status, body)
	}
}
