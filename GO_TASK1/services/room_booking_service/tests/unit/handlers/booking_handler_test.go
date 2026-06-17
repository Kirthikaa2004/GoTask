package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/handlers"
	"room_booking_management/room_booking_service/internal/loggers"
	serviceMock "room_booking_management/room_booking_service/tests/unit/services/mock"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


func newReservationHandler(t *testing.T) (handlers.ReservationHandlerInterface, *serviceMock.ReservationServiceMock) {
	t.Helper()
	mockSvc := new(serviceMock.ReservationServiceMock)
	logger := loggers.NewLogger(loggers.LogConfig{
		ServiceName: "test-service",
		LogDir:      t.TempDir(),
		FileName:    "test.log",
	}) 
	t.Cleanup(func() { logger.Sync() })
	h := handlers.NewReservationHandler(mockSvc, logger)
	return h, mockSvc
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	assert.NoError(t, err)
	return bytes.NewBuffer(b)
}

func decodeResponse(t *testing.T, body *bytes.Buffer) dtos.APIResponse {
	t.Helper()
	var resp dtos.APIResponse
	assert.NoError(t, json.Unmarshal(body.Bytes(), &resp))
	return resp
}


func TestReserveRoom(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		rawBody        string
		mockResponse   *dtos.APIResponse
		setupMock      bool
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "Success - Room Reserved",
			body: map[string]interface{}{
				"name":         "Alice",
				"age":          25,
				"address":      "123 Main St",
				"phone_number": "+1234567890",
				"purpose":      "Conference",
				"room_id":      1,
				"room_floor":   2,
				"room_name":    "Boardroom A",
				"start_date":   "2026-06-01T10:00:00Z",
				"end_date":     "2026-06-01T12:00:00Z",
			},
			mockResponse: &dtos.APIResponse{
				Status:  "success",
				Code:    http.StatusOK,
				Message: "Room reserved successfully",
				Data: &dtos.RoomResponse{
					BookingID:  1,
					RoomID:     1,
					Status:     "confirmed",
					TotalPrice: 200.00,
				},
			},
			setupMock:      true,
			expectedStatus: http.StatusOK,
			expectedMsg:    "Room reserved successfully",
		},
		{
			name:           "Failure - Malformed JSON Body",
			rawBody:        `{"name": "Alice", "age":}`,
			setupMock:      false,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid request body",
		},
		{
			name: "Failure - Validation Error (missing required fields)",
			body: map[string]interface{}{
				"name": "",
			},
			setupMock:      false,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "validation failed",
		},
		{
			name: "Failure - Room Already Booked",
			body: map[string]interface{}{
				"name":         "Bob",
				"age":          30,
				"address":      "456 Elm St",
				"phone_number": "+9876543210",
				"purpose":      "Meeting",
				"room_id":      1,
				"room_floor":   2,
				"room_name":    "Boardroom A",
				"start_date":   "2026-06-01T10:00:00Z",
				"end_date":     "2026-06-01T12:00:00Z",
			},
			mockResponse: &dtos.APIResponse{
				Status:  "fail",
				Code:    http.StatusConflict,
				Message: "Room is already booked for the selected dates",
				Errors: []dtos.Error{
					{Field: "room_id", Message: "room already booked", Code: "CONFLICT"},
				},
			},
			setupMock:      true,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Room is already booked for the selected dates",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, mockSvc := newReservationHandler(t)

			var reqBody *bytes.Buffer
			if tc.body != nil {
				reqBody = toJSON(t, tc.body)
			} else {
				reqBody = bytes.NewBufferString(tc.rawBody)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/reservations", reqBody)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			if tc.setupMock {
				mockSvc.On("ReserveRoom", mock.Anything, mock.Anything).Return(tc.mockResponse)
			}

			h.ReserveRoom(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			resp := decodeResponse(t, rr.Body)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			if tc.setupMock {
				mockSvc.AssertExpectations(t)
			} else {
				mockSvc.AssertNotCalled(t, "ReserveRoom", mock.Anything, mock.Anything)
			}
		})
	}
}


func TestCancelBooking(t *testing.T) {
	tests := []struct {
		name           string
		bookingID      string
		setMuxVar      bool
		mockResponse   *dtos.APIResponse
		setupMock      bool
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:      "Success - Booking Cancelled",
			bookingID: "booking-uuid",
			setMuxVar: true,
			mockResponse: &dtos.APIResponse{
				Status:  "success",
				Code:    http.StatusOK,
				Message: "Booking cancelled successfully",
				Data: &dtos.CancelRoomResponse{
					Status: "cancelled",
				},
			},
			setupMock:      true,
			expectedStatus: http.StatusOK,
			expectedMsg:    "Booking cancelled successfully",
		},
		{
			name:           "Failure - Missing BookingID in Path",
			bookingID:      "",
			setMuxVar:      false,
			setupMock:      false,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "booking_id is required",
		},
		{
			name:      "Failure - Booking Not Found",
			bookingID: "non-existent-uuid",
			setMuxVar: true,
			mockResponse: &dtos.APIResponse{
				Status:  "fail",
				Code:    http.StatusNotFound,
				Message: "Booking not found",
				Errors: []dtos.Error{
					{Field: "bookingID", Message: "booking not found", Code: "NOT_FOUND"},
				},
			},
			setupMock:      true,
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Booking not found",
		},
		{
			name:      "Failure - Booking Already Cancelled",
			bookingID: "already-cancelled-uuid",
			setMuxVar: true,
			mockResponse: &dtos.APIResponse{
				Status:  "fail",
				Code:    http.StatusConflict,
				Message: "Booking is already cancelled",
				Errors: []dtos.Error{
					{Field: "bookingID", Message: "already cancelled", Code: "CONFLICT"},
				},
			},
			setupMock:      true,
			expectedStatus: http.StatusConflict,
			expectedMsg:    "Booking is already cancelled",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, mockSvc := newReservationHandler(t)

			req, err := http.NewRequest(http.MethodPut,
				"/api/v1/reservations/"+tc.bookingID+"/cancel", nil)
			assert.NoError(t, err)

			if tc.setMuxVar {
				req = mux.SetURLVars(req, map[string]string{"bookingID": tc.bookingID})
			}

			rr := httptest.NewRecorder()

			if tc.setupMock {
				mockSvc.On("CancelBooking", mock.Anything, tc.bookingID).Return(tc.mockResponse)
			}

			h.CancelBooking(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			resp := decodeResponse(t, rr.Body)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			if tc.setupMock {
				mockSvc.AssertExpectations(t)
			} else {
				mockSvc.AssertNotCalled(t, "CancelBooking", mock.Anything, mock.Anything)
			}
		})
	}
}


func TestGetRooms(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		bookingID      string
		setMuxVar      bool
		mockResponse   *dtos.APIResponse
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "Success - Get All Rooms",
			path: "/api/v1/rooms",
			mockResponse: &dtos.APIResponse{
				Status:  "success",
				Code:    http.StatusOK,
				Message: "All rooms fetched successfully",
				Data: []dtos.GetAllRoomResponse{
					{
						RoomID:    1,
						RoomName:  "Boardroom A",
						RoomType:  "conference",
						Capacity:  10,
						RoomFloor: 2,
						IsActive:  true,
					},
					{
						RoomID:    2,
						RoomName:  "Suite B",
						RoomType:  "suite",
						Capacity:  4,
						RoomFloor: 3,
						IsActive:  true,
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "All rooms fetched successfully",
		},
		{
			name: "Success - Get All Rooms Empty List",
			path: "/api/v1/rooms",
			mockResponse: &dtos.APIResponse{
				Status:  "success",
				Code:    http.StatusOK,
				Message: "All rooms fetched successfully",
				Data:    []dtos.GetAllRoomResponse{},
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "All rooms fetched successfully",
		},
		{
			name:      "Success - Get Room By BookingID",
			path:      "/api/v1/rooms/1",
			bookingID: "1",
			setMuxVar: true,
			mockResponse: &dtos.APIResponse{
				Status:  "success",
				Code:    http.StatusOK,
				Message: "Room fetched successfully",
				Data: &dtos.GetRoomResponse{
					BookingID: 1,
					Capacity:  10,
					RoomType:  "conference",
					RoomName:  "Boardroom A",
					RoomFloor: 2,
					Price:     200.00,
				},
			},
			expectedStatus: http.StatusOK,
			expectedMsg:    "Room fetched successfully",
		},
		{
			name:      "Failure - Room Not Found For BookingID",
			path:      "/api/v1/rooms/999",
			bookingID: "999",
			setMuxVar: true,
			mockResponse: &dtos.APIResponse{
				Status:  "fail",
				Code:    http.StatusNotFound,
				Message: "Room not found for booking",
				Errors: []dtos.Error{
					{Field: "bookingID", Message: "no room found", Code: "NOT_FOUND"},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "Room not found for booking",
		},
		{
			name: "Failure - DB Error On Get All Rooms",
			path: "/api/v1/rooms",
			mockResponse: &dtos.APIResponse{
				Status:  "error",
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
				Errors: []dtos.Error{
					{Field: "db", Message: "connection failed", Code: "INTERNAL_ERROR"},
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "Internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, mockSvc := newReservationHandler(t)

			req, err := http.NewRequest(http.MethodGet, tc.path, nil)
			assert.NoError(t, err)

			if tc.setMuxVar && tc.bookingID != "" {
				req = mux.SetURLVars(req, map[string]string{"bookingID": tc.bookingID})
				mockSvc.On("GetRoomByBookingID", mock.Anything, tc.bookingID).
					Return(tc.mockResponse)
			} else {
				mockSvc.On("GetRooms", mock.Anything).Return(tc.mockResponse)
			}

			rr := httptest.NewRecorder()
			h.GetRooms(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			resp := decodeResponse(t, rr.Body)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			mockSvc.AssertExpectations(t)
		})
	}
}