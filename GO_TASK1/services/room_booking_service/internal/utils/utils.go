package utils

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "github.com/google/uuid"
    "room_booking_management/room_booking_service/internal/dtos"
    "room_booking_management/room_booking_service/internal/errorcodes"

    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

func WriteResponse(w http.ResponseWriter, code int, data interface{}) {
    if w.Header().Get("Content-Type") == "" {
        w.Header().Set("Content-Type", "application/json")
    }
    w.WriteHeader(code)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func ValidateStruct(s interface{}) []dtos.Error {
    var validationErrors []dtos.Error

    if err := validate.Struct(s); err != nil {
        if validationErrs, ok := err.(validator.ValidationErrors); ok {
            for _, err := range validationErrs {
                validationErrors = append(validationErrors, dtos.Error{
                    Field:   err.Field(),
                    Message: fmt.Sprintf("Validation failed for the field '%s' on the '%s' constraint.", err.Field(), err.Tag()),
                    Code:    string(errorcodes.ErrCodeValidation),
                })
            }
        } else {
            validationErrors = append(validationErrors, dtos.Error{
                Field:   "unknown",
                Message: err.Error(),
                Code:    string(errorcodes.ErrCodeInternal),
            })
        }
    }

    return validationErrors
}

func MapErrorCode(apiResponse *dtos.APIResponse) *dtos.APIResponse {
    if len(apiResponse.Errors) == 0 {
        apiResponse.Code = http.StatusOK
        return apiResponse
    }

    switch apiResponse.Errors[0].Code {
    case string(errorcodes.ErrCodeNotFound):
        apiResponse.Code = http.StatusNotFound
    case string(errorcodes.ErrCodeValidation):
        apiResponse.Code = http.StatusBadRequest
    case string(errorcodes.ErrCodeInternal):
        apiResponse.Code = http.StatusInternalServerError
    case string(errorcodes.ErrCodeUnauthorized):
        apiResponse.Code = http.StatusUnauthorized
    case string(errorcodes.ErrCodeBadRequest):
        apiResponse.Code = http.StatusBadRequest
    case string(errorcodes.ErrCodeConflict):
        apiResponse.Code = http.StatusConflict
    default:
        apiResponse.Code = http.StatusInternalServerError
    }

    return apiResponse
}
func GetOrGenerateRequestID(r *http.Request) string{
	reqID := r.Header.Get("X-Request-ID")
	if reqID == ""{
		reqID = uuid.NewString()
	}
	return reqID
}

func NewErrorResponse(code int, message string, errs []dtos.Error,reqID string) *dtos.APIResponse {
    status := "fail"
    if code >= 500 {
        status = "error"
    }
    return &dtos.APIResponse{
        Status:    status,
        Code:      code,
        Message:   message,
        Errors:    errs,
		RequestID: reqID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}


func NewSuccessResponse(code int, message string, data interface{},reqID string) *dtos.APIResponse {
    return &dtos.APIResponse{
        Status:    "success",
        Code:      code,
        Message:   message,
        Data:      data,
		RequestID: reqID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}