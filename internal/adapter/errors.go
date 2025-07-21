package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	apperrors "github.com/Snake1-1eyes/vk_task_marketplace/internal/app_errors"
	"github.com/Snake1-1eyes/vk_task_marketplace/internal/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrorCodeUserNotFound       = "USER_NOT_FOUND"
	ErrorCodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ErrorCodeListingNotFound    = "LISTING_NOT_FOUND"
	ErrorCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrorCodeInvalidToken       = "INVALID_TOKEN"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeForbidden          = "FORBIDDEN"
	ErrorCodeValidationFailed   = "VALIDATION_FAILED"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
)

// ErrorResponse представляет формат ошибки для JSON-ответа
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ExtractErrorCode извлекает код ошибки из стандартных ошибок приложения
func ExtractErrorCode(err error) string {
	switch {
	case errors.Is(err, apperrors.ErrUserNotFound):
		return ErrorCodeUserNotFound
	case errors.Is(err, apperrors.ErrUserAlreadyExists):
		return ErrorCodeUserAlreadyExists
	case errors.Is(err, apperrors.ErrListingNotFound):
		return ErrorCodeListingNotFound
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		return ErrorCodeInvalidCredentials
	case errors.Is(err, apperrors.ErrInvalidToken):
		return ErrorCodeInvalidToken
	case errors.Is(err, apperrors.ErrUnauthorized):
		return ErrorCodeUnauthorized
	case errors.Is(err, apperrors.ErrForbidden):
		return ErrorCodeForbidden
	case errors.Is(err, apperrors.ErrValidation):
		return ErrorCodeValidationFailed
	default:
		return ErrorCodeInternalError
	}
}

// MapError преобразует внутреннюю ошибку в gRPC-ошибку с деталями
func MapError(err error) error {
	code := ExtractErrorCode(err)
	message := err.Error()
	grpcCode := mapErrorCodeToGRPCCode(code)

	st := status.New(grpcCode, message)

	errorInfo := &errdetails.ErrorInfo{
		Reason:   code,
		Metadata: map[string]string{"message": message},
	}

	stWithDetails, detailErr := st.WithDetails(errorInfo)
	if detailErr != nil {
		return st.Err()
	}

	return stWithDetails.Err()
}

// mapErrorCodeToGRPCCode преобразует код ошибки в gRPC код
func mapErrorCodeToGRPCCode(code string) codes.Code {
	switch code {
	case ErrorCodeUserNotFound, ErrorCodeListingNotFound:
		return codes.NotFound
	case ErrorCodeUserAlreadyExists:
		return codes.AlreadyExists
	case ErrorCodeInvalidCredentials, ErrorCodeInvalidToken, ErrorCodeUnauthorized:
		return codes.Unauthenticated
	case ErrorCodeForbidden:
		return codes.PermissionDenied
	case ErrorCodeValidationFailed:
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

// CustomHTTPError обрабатывает gRPC ошибки и преобразует их в HTTP ответы с нужным форматом
func CustomHTTPError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	log := logger.GetLoggerFromCtx(ctx)

	s, ok := status.FromError(err)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		writeErrorResponse(ctx, w, ErrorCodeInternalError, "Внутренняя ошибка сервера", log)
		return
	}

	httpStatus := runtime.HTTPStatusFromCode(s.Code())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Пытаемся получить детали ошибки из status
	errorCode := ErrorCodeInternalError
	errorMessage := s.Message()

	// Извлекаем ErrorInfo из деталей
	for _, detail := range s.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok {
			errorCode = info.Reason
			if msg := info.Metadata["message"]; msg != "" {
				errorMessage = msg
			}
			break
		}
	}

	writeErrorResponse(ctx, w, errorCode, errorMessage, log)
}

// writeErrorResponse записывает ответ об ошибке в формате согласно контракту
func writeErrorResponse(ctx context.Context, w http.ResponseWriter, code string, message string, log *logger.Logger) {
	resp := ErrorResponse{}
	resp.Error.Code = code
	resp.Error.Message = message

	data, err := json.Marshal(resp)
	if err != nil {
		log.Error(ctx, "Ошибка при маршалинге ошибки",
			zap.Error(err),
			zap.String("errorCode", code),
		)
		if _, writeErr := w.Write([]byte(`{"error":{"code":"INTERNAL_ERROR","message":"Ошибка форматирования ответа"}}`)); writeErr != nil {
			log.Error(ctx, "Ошибка при отправке ответа",
				zap.Error(writeErr),
			)
		}
		return
	}

	if _, writeErr := w.Write(data); writeErr != nil {
		log.Error(ctx, "Ошибка при записи HTTP-ответа",
			zap.Error(writeErr),
			zap.String("errorCode", code),
		)
	}
}

func WrapValidationError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", ErrorCodeValidationFailed, err.Error())
}
