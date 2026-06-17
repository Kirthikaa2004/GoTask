package errorcodes
type ErrorCode string

const (
	ErrCodeValidation ErrorCode ="VALIDATION_ERROR_400"
	ErrCodeNotFound   ErrorCode ="NOT_FOUND_404"
	ErrCodeConflict   ErrorCode = "CONFLICT_409"
	ErrCodeInternal   ErrorCode ="INTERNAL_ERROR_500"
	ErrCodeBadRequest ErrorCode = "BAD_REQUEST_400"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
)

var ErrorCodeStatus = map[ErrorCode] string{
	ErrCodeValidation: "Validation Failed",
	ErrCodeNotFound: "Resource Not Found",
	ErrCodeConflict: "Resource Conflict",
	ErrCodeInternal: "Internal Server Error",
	ErrCodeBadRequest: "Bad Request",
	ErrCodeUnauthorized: "Unauthorized",
}