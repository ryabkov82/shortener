package testutils

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// StatusCode — универсальный enum статусов для тестов.
type StatusCode int

const (
	StatusOK StatusCode = iota
	StatusCreated
	StatusBadRequest
	StatusUnauthorized
	StatusForbidden
	StatusNotFound
	StatusConflict
	StatusTemporaryRedirect
	StatusAccepted
	StatusNoContent
	StatusServiceUnavailable
	StatusGatewayTimeout
	StatusInternalError
	StatusUnknown
)

// grpcCodeToStatusCode преобразует grpc коды в StatusCode
func GRPCCodeToStatusCode(code codes.Code) StatusCode {
	switch code {
	case codes.OK:
		return StatusCreated
	case codes.InvalidArgument:
		return StatusBadRequest
	case codes.NotFound:
		return StatusNotFound
	case codes.AlreadyExists:
		return StatusConflict
	case codes.PermissionDenied:
		return StatusForbidden
	case codes.Unauthenticated:
		return StatusUnauthorized
	case codes.Unavailable:
		return StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return StatusGatewayTimeout
	case codes.Internal:
		return StatusInternalError
	default:
		return StatusUnknown
	}
}

// httpStatusToStatusCode преобразует HTTP статус в StatusCode
func HTTPStatusToStatusCode(code int) StatusCode {
	switch code {
	case http.StatusOK:
		return StatusOK
	case http.StatusCreated:
		return StatusCreated
	case http.StatusBadRequest:
		return StatusBadRequest
	case http.StatusUnauthorized:
		return StatusUnauthorized
	case http.StatusForbidden:
		return StatusForbidden
	case http.StatusNotFound:
		return StatusNotFound
	case http.StatusConflict:
		return StatusConflict
	case http.StatusServiceUnavailable:
		return StatusServiceUnavailable
	case http.StatusGatewayTimeout:
		return StatusGatewayTimeout
	case http.StatusInternalServerError:
		return StatusInternalError
	case http.StatusTemporaryRedirect:
		return StatusTemporaryRedirect
	case http.StatusAccepted:
		return StatusAccepted
	case http.StatusNoContent:
		return StatusNoContent
	default:
		return StatusUnknown
	}
}
