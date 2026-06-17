package dtos

import "room_booking_management/room_booking_service/internal/errorcodes"

type APIResponse struct {
	Status    string      `json:"status"`          
	Code      int         `json:"code"`             
	Message   string      `json:"message"`          
	Data      interface{} `json:"data,omitempty"`   
	Errors    []Error     `json:"errors,omitempty"` 
	Meta      *Meta       `json:"meta,omitempty"`   
	RequestID string      `json:"request_id"`       
	Timestamp string      `json:"timestamp"`       
}

type Error struct {
	Field   string `json:"field,omitempty"` 
	Message string `json:"message"`         
	Code    string `json:"code"`            
}

func (e *Error) Error() string {
	return e.Message
}

type Meta struct {
	Page     int `json:"page"`      
	PageSize int `json:"page_size"` 
	Total    int64 `json:"total"`    
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ValidationErrorToAPIError(err *ValidationError) Error {
	return Error{
		Field:   err.Field,
		Message: err.Message,
		Code:    string(errorcodes.ErrCodeValidation),
	}
}