package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/errorcodes"
	"room_booking_management/room_booking_service/internal/utils"

	"github.com/stretchr/testify/assert"
)

// ── WriteResponse ─────────────────────────────────────────────────────────────

func TestWriteResponse(t *testing.T) {
	t.Run("Success response", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		utils.WriteResponse(w, http.StatusOK, data)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"message":"success"}`, w.Body.String())
	})

	t.Run("Database Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"message": "internal server"}

		utils.WriteResponse(w, http.StatusInternalServerError, data)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"message":"internal server"}`, w.Body.String())
	})

	t.Run("Item Not Found Error", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"message": "data not found"}

		utils.WriteResponse(w, http.StatusNotFound, data)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"message":"data not found"}`, w.Body.String())
	})

	t.Run("Error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"error": "bad request"}

		utils.WriteResponse(w, http.StatusBadRequest, data)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"error":"bad request"}`, w.Body.String())
	})

	t.Run("Json Encode Error response", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := make(chan int) // channels cannot be JSON-encoded

		utils.WriteResponse(w, http.StatusInternalServerError, data)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ── ValidateStruct ────────────────────────────────────────────────────────────

func TestValidateStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"required,email"`
	}

	t.Run("Valid struct", func(t *testing.T) {
		input := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
		}

		errs := utils.ValidateStruct(&input)
		assert.Nil(t, errs)
	})

	t.Run("Missing name", func(t *testing.T) {
		input := TestStruct{
			Email: "john@example.com",
		}

		expected := []dtos.Error{
			{
				Field:   "Name",
				Message: "Validation failed for the field 'Name' on the 'required' constraint.",
				Code:    string(errorcodes.ErrCodeValidation),
			},
		}

		errs := utils.ValidateStruct(&input)
		assert.Equal(t, expected, errs)
	})

	t.Run("Invalid email", func(t *testing.T) {
		input := TestStruct{
			Name:  "John Doe",
			Email: "invalid-email",
		}

		expected := []dtos.Error{
			{
				Field:   "Email",
				Message: "Validation failed for the field 'Email' on the 'email' constraint.",
				Code:    string(errorcodes.ErrCodeValidation),
			},
		}

		errs := utils.ValidateStruct(&input)
		assert.Equal(t, expected, errs)
	})

	t.Run("Empty struct - both fields missing", func(t *testing.T) {
		input := TestStruct{}

		expected := []dtos.Error{
			{
				Field:   "Name",
				Message: "Validation failed for the field 'Name' on the 'required' constraint.",
				Code:    string(errorcodes.ErrCodeValidation),
			},
			{
				Field:   "Email",
				Message: "Validation failed for the field 'Email' on the 'required' constraint.",
				Code:    string(errorcodes.ErrCodeValidation),
			},
		}

		errs := utils.ValidateStruct(&input)
		assert.ElementsMatch(t, expected, errs)
	})

	t.Run("Non-struct input returns internal error", func(t *testing.T) {
		input := "not a struct"

		errs := utils.ValidateStruct(input)
		assert.Len(t, errs, 1)
		assert.Equal(t, "unknown", errs[0].Field)
		assert.Equal(t, string(errorcodes.ErrCodeInternal), errs[0].Code)
	})
}

// ── MapErrorCode ──────────────────────────────────────────────────────────────

func TestMapErrorCode(t *testing.T) {
	t.Run("No errors returns 200", func(t *testing.T) {
		input := &dtos.APIResponse{Errors: []dtos.Error{}}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusOK, result.Code)
	})

	t.Run("Not Found Error", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: string(errorcodes.ErrCodeNotFound)}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusNotFound, result.Code)
	})

	t.Run("Validation Error", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: string(errorcodes.ErrCodeValidation)}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusBadRequest, result.Code)
	})

	t.Run("Internal Error", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: string(errorcodes.ErrCodeInternal)}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusInternalServerError, result.Code)
	})

	t.Run("Unauthorized Error", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: string(errorcodes.ErrCodeUnauthorized)}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusUnauthorized, result.Code)
	})

	t.Run("Bad Request Error", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: string(errorcodes.ErrCodeBadRequest)}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusBadRequest, result.Code)
	})

	t.Run("Conflict Error", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: string(errorcodes.ErrCodeConflict)}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusConflict, result.Code)
	})

	t.Run("Unknown Error defaults to 500", func(t *testing.T) {
		input := &dtos.APIResponse{
			Errors: []dtos.Error{{Code: "TOTALLY_UNKNOWN"}},
		}

		result := utils.MapErrorCode(input)
		assert.Equal(t, http.StatusInternalServerError, result.Code)
	})
}

// ── NewErrorResponse ──────────────────────────────────────────────────────────

func TestNewErrorResponse(t *testing.T) {
	t.Run("4xx sets status to fail", func(t *testing.T) {
		resp := utils.NewErrorResponse(http.StatusBadRequest, "bad input", []dtos.Error{
			{Field: "name", Message: "required", Code: string(errorcodes.ErrCodeValidation)},
		}, "req-001")

		assert.Equal(t, "fail", resp.Status)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "bad input", resp.Message)
		assert.Equal(t, "req-001", resp.RequestID)
		assert.Len(t, resp.Errors, 1)
		assert.NotEmpty(t, resp.Timestamp)
	})

	t.Run("5xx sets status to error", func(t *testing.T) {
		resp := utils.NewErrorResponse(http.StatusInternalServerError, "something broke", []dtos.Error{
			{Field: "db", Message: "unreachable", Code: string(errorcodes.ErrCodeInternal)},
		}, "req-002")

		assert.Equal(t, "error", resp.Status)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Equal(t, "something broke", resp.Message)
		assert.Equal(t, "req-002", resp.RequestID)
		assert.Len(t, resp.Errors, 1)
		assert.NotEmpty(t, resp.Timestamp)
	})

	t.Run("Empty errors slice is preserved", func(t *testing.T) {
		resp := utils.NewErrorResponse(http.StatusNotFound, "not found", []dtos.Error{}, "req-003")

		assert.Equal(t, "fail", resp.Status)
		assert.Equal(t, "req-003", resp.RequestID)
		assert.Empty(t, resp.Errors)
	})

	t.Run("Empty requestID is allowed", func(t *testing.T) {
		resp := utils.NewErrorResponse(http.StatusBadRequest, "bad input", []dtos.Error{}, "")

		assert.Equal(t, "fail", resp.Status)
		assert.Equal(t, "", resp.RequestID)
	})
}

// ── NewSuccessResponse ────────────────────────────────────────────────────────

func TestNewSuccessResponse(t *testing.T) {
	t.Run("Returns success status and data", func(t *testing.T) {
		payload := map[string]string{"id": "123"}
		resp := utils.NewSuccessResponse(http.StatusOK, "created", payload, "req-001")

		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "created", resp.Message)
		assert.Equal(t, payload, resp.Data)
		assert.Equal(t, "req-001", resp.RequestID)
		assert.NotEmpty(t, resp.Timestamp)
	})

	t.Run("Nil data is allowed", func(t *testing.T) {
		resp := utils.NewSuccessResponse(http.StatusNoContent, "deleted", nil, "req-002")

		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "req-002", resp.RequestID)
		assert.Nil(t, resp.Data)
	})

	t.Run("Empty requestID is allowed", func(t *testing.T) {
		resp := utils.NewSuccessResponse(http.StatusOK, "ok", nil, "")

		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "", resp.RequestID)
	})
}

// ── GetOrGenerateRequestID ────────────────────────────────────────────────────

func TestGetOrGenerateRequestID(t *testing.T) {
	t.Run("Returns existing X-Request-ID header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "my-custom-id")

		reqID := utils.GetOrGenerateRequestID(req)

		assert.Equal(t, "my-custom-id", reqID)
	})

	t.Run("Generates UUID when header is absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// No X-Request-ID header set

		reqID := utils.GetOrGenerateRequestID(req)

		assert.NotEmpty(t, reqID)
		// UUID v4 format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, reqID)
	})

	t.Run("Generates UUID when header is empty string", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "")

		reqID := utils.GetOrGenerateRequestID(req)

		assert.NotEmpty(t, reqID)
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, reqID)
	})

	t.Run("Two calls without header produce different UUIDs", func(t *testing.T) {
		req1 := httptest.NewRequest(http.MethodGet, "/", nil)
		req2 := httptest.NewRequest(http.MethodGet, "/", nil)

		id1 := utils.GetOrGenerateRequestID(req1)
		id2 := utils.GetOrGenerateRequestID(req2)

		assert.NotEqual(t, id1, id2)
	})
}