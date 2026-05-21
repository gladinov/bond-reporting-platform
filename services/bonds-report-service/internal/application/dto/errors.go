package dto

type ErrorCode string

const (
	// клиентские ошибки (аналог 4xx)
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrTokenInvalid ErrorCode = "TOKEN_INVALID"

	// бизнес-ошибки
	ErrReportNotFound ErrorCode = "REPORT_NOT_FOUND"
	ErrNoData         ErrorCode = "NO_DATA"

	// внешние зависимости
	ErrTinkoffAPI ErrorCode = "TINKOFF_API_ERROR"
	ErrMOEXAPI    ErrorCode = "MOEX_API_ERROR"

	// системные ошибки (аналог 5xx)
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
	ErrRenderFailed ErrorCode = "RENDER_FAILED"
)
