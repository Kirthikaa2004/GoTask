package services_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/internal/models"
	"room_booking_management/room_booking_service/internal/services"
	repoMock "room_booking_management/room_booking_service/tests/unit/repository/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


func newTestLogger(t *testing.T) *loggers.Logger {
	t.Helper()
	logger := loggers.NewLogger(loggers.LogConfig{
		ServiceName: "test-service",
		LogDir:      t.TempDir(),
		FileName:    "test.log",
	})
	t.Cleanup(func() { logger.Sync() })
	return logger
}


func makeRequest(startOffset, endOffset int) dtos.CreateBookingRequest {
	now := time.Now()
	return dtos.CreateBookingRequest{
		RoomID:      1,
		Name:        "John Doe",
		Age:         30,
		Address:     "123 Main St",
		PhoneNumber: "0812345678",
		Purpose:     "Meeting",
		StartDate:   now.AddDate(0, 0, startOffset),
		EndDate:     now.AddDate(0, 0, endOffset),
	}
}

func mockRoom() *models.Room {
	return &models.Room{
		ID:       1,
		RoomType: "Deluxe",
		Capacity: 2,
		Price:    500.0,
		IsActive: true,
	}
}

func mockCustomer(req dtos.CreateBookingRequest) *models.Customer {
	return &models.Customer{
		ID:          1,
		UUID:        "customer-uuid",
		Name:        req.Name,
		Age:         req.Age,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		IsActive:    true,
	}
}

func mockBooking(req dtos.CreateBookingRequest) *models.Booking {
	return &models.Booking{
		ID:        1,
		UUID:      "booking-uuid",
		RoomID:    req.RoomID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Purpose:   req.Purpose,
		IsActive:  true,
	}
}


func TestReserveRoom(t *testing.T) {
	testCases := []struct {
		name          string
		request       dtos.CreateBookingRequest
		setupMocks    func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest)
		expectedCode  int
		expectedMsg   string
		expectedError bool
		isInternal    bool
	}{
		{
			name:    "Success - New Customer",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				booking := mockBooking(req)

				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(true, nil).Once()
				r.On("GetCustomerByPhone", mock.Anything, req.PhoneNumber).Return(nil, nil).Once()
				r.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*models.Customer")).Return(nil).Once()
				r.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.Booking")).
					Run(func(args mock.Arguments) {
						b := args.Get(1).(*models.Booking)
						b.ID = booking.ID
					}).Return(nil).Once()
			},
			expectedCode:  http.StatusCreated,
			expectedMsg:   "Room reserved successfully",
			expectedError: false,
			isInternal:    false,
		},
		{
			name:    "Success - Existing Customer",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				booking := mockBooking(req)

				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(true, nil).Once()
				r.On("GetCustomerByPhone", mock.Anything, req.PhoneNumber).Return(mockCustomer(req), nil).Once()
				r.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.Booking")).
					Run(func(args mock.Arguments) {
						b := args.Get(1).(*models.Booking)
						b.ID = booking.ID
					}).Return(nil).Once()
			},
			expectedCode:  http.StatusCreated,
			expectedMsg:   "Room reserved successfully",
			expectedError: false,
			isInternal:    false,
		},
		{
			name:          "Failure - End date before start date",
			request:       makeRequest(3, 1),
			setupMocks:    func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {},
			expectedCode:  http.StatusBadRequest,
			expectedMsg:   "end_date must be after start_date",
			expectedError: true,
			isInternal:    false,
		},
		{
			name:          "Failure - End date equal to start date",
			request:       makeRequest(1, 1),
			setupMocks:    func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {},
			expectedCode:  http.StatusBadRequest,
			expectedMsg:   "end_date must be after start_date",
			expectedError: true,
			isInternal:    false,
		},
		{
			name:    "Failure - Room not found",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(nil, &dtos.Error{
					Code: "NOT_FOUND", Message: "room not found",
				}).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedMsg:   "room not found",
			expectedError: true,
			isInternal:    false,
		},
		{
			name:    "Failure - Room lookup DB error",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(nil, &dtos.Error{
					Code: "INTERNAL_ERROR", Message: "db error",
				}).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "room not found",
			expectedError: true,
			isInternal:    true,
		},
		{
			name:    "Failure - Availability check error",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(false, &dtos.Error{
					Code: "INTERNAL_ERROR", Message: "availability check failed",
				}).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "failed to check room availability",
			expectedError: true,
			isInternal:    true,
		},
		{
			name:    "Failure - Room not available (conflict)",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(false, nil).Once()
			},
			expectedCode:  http.StatusConflict,
			expectedMsg:   "room is already booked for the requested period",
			expectedError: true,
			isInternal:    false,
		},
		{
			name:    "Failure - Customer phone lookup error",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(true, nil).Once()
				r.On("GetCustomerByPhone", mock.Anything, req.PhoneNumber).Return(nil, &dtos.Error{
					Code: "INTERNAL_ERROR", Message: "failed to lookup customer",
				}).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "failed to lookup customer",
			expectedError: true,
			isInternal:    true,
		},
		{
			name:    "Failure - CreateCustomer error",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(true, nil).Once()
				r.On("GetCustomerByPhone", mock.Anything, req.PhoneNumber).Return(nil, nil).Once()
				r.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*models.Customer")).Return(&dtos.Error{
					Code: "INTERNAL_ERROR", Message: "failed to create customer",
				}).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "failed to create customer",
			expectedError: true,
			isInternal:    true,
		},
		{
			name:    "Failure - CreateBooking error",
			request: makeRequest(1, 3),
			setupMocks: func(r *repoMock.ReservationRepositoryMock, req dtos.CreateBookingRequest) {
				r.On("GetRoomByID", mock.Anything, req.RoomID).Return(mockRoom(), nil).Once()
				r.On("CheckRoomAvailability", mock.Anything, req.RoomID, req.StartDate, req.EndDate).Return(true, nil).Once()
				r.On("GetCustomerByPhone", mock.Anything, req.PhoneNumber).Return(nil, nil).Once()
				r.On("CreateCustomer", mock.Anything, mock.AnythingOfType("*models.Customer")).Return(nil).Once()
				r.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.Booking")).Return(&dtos.Error{
					Code: "INTERNAL_ERROR", Message: "failed to create booking",
				}).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "failed to create booking",
			expectedError: true,
			isInternal:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(repoMock.ReservationRepositoryMock)
			tc.setupMocks(mockRepo, tc.request)

			svc := services.NewReservationService(mockRepo, newTestLogger(t))
			resp := svc.ReserveRoom(context.Background(), tc.request)

			assert.NotNil(t, resp)
			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			if !tc.expectedError {
				assert.Equal(t, "success", resp.Status)
				roomResp, ok := resp.Data.(dtos.RoomResponse)
				assert.True(t, ok)
				assert.Equal(t, "CONFIRMED", roomResp.Status)
				assert.Equal(t, int(tc.request.RoomID), roomResp.RoomID)
				assert.Greater(t, roomResp.TotalPrice, 0.0)
			} else {
				if tc.isInternal {
					assert.Equal(t, "error", resp.Status)
				} else {
					assert.Equal(t, "fail", resp.Status)
				}
				assert.NotEmpty(t, resp.Errors)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}


func TestCancelBooking(t *testing.T) {
	testCases := []struct {
		name          string
		bookingID     string
		mockRepoError *dtos.Error
		expectedCode  int
		expectedMsg   string
		expectedError bool
		isInternal    bool
	}{
		{
			name:          "Success - Cancel Booking",
			bookingID:     "valid-booking-id",
			mockRepoError: nil,
			expectedCode:  http.StatusOK,
			expectedMsg:   "Booking cancelled successfully",
			expectedError: false,
			isInternal:    false,
		},
		{
			name:      "Failure - Booking Not Found",
			bookingID: "invalid-booking-id",
			mockRepoError: &dtos.Error{
				Code:    "NOT_FOUND",
				Message: "booking not found",
			},
			expectedCode:  http.StatusNotFound,
			expectedMsg:   "failed to cancel booking",
			expectedError: true,
			isInternal:    false,
		},
		{
			name:      "Failure - Database Error",
			bookingID: "valid-booking-id",
			mockRepoError: &dtos.Error{
				Code:    "INTERNAL_ERROR",
				Message: "db error",
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "failed to cancel booking",
			expectedError: true,
			isInternal:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(repoMock.ReservationRepositoryMock)
			mockRepo.On("CancelBooking", mock.Anything, tc.bookingID).Return(tc.mockRepoError).Once()

			svc := services.NewReservationService(mockRepo, newTestLogger(t))
			resp := svc.CancelBooking(context.Background(), tc.bookingID)

			assert.NotNil(t, resp)
			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			if !tc.expectedError {
				assert.Equal(t, "success", resp.Status)
				cancelResp, ok := resp.Data.(dtos.CancelRoomResponse)
				assert.True(t, ok)
				assert.Equal(t, "CANCELLED", cancelResp.Status)
			} else {
				if tc.isInternal {
					assert.Equal(t, "error", resp.Status)
				} else {
					assert.Equal(t, "fail", resp.Status)
				}
				assert.NotEmpty(t, resp.Errors)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}


func TestGetRooms(t *testing.T) {
	mockRooms := []models.Room{
		{ID: 1, RoomType: "Deluxe", Capacity: 2, IsActive: true},
		{ID: 2, RoomType: "Suite", Capacity: 4, IsActive: true},
	}

	testCases := []struct {
		name          string
		mockRooms     []models.Room
		mockTotal     int64
		mockError     *dtos.Error
		expectedCode  int
		expectedMsg   string
		expectedError bool
	}{
		{
			name:          "Success - Rooms fetched",
			mockRooms:     mockRooms,
			mockTotal:     2,
			mockError:     nil,
			expectedCode:  http.StatusOK,
			expectedMsg:   "Rooms fetched successfully",
			expectedError: false,
		},
		{
			name:          "Success - Empty rooms list",
			mockRooms:     []models.Room{},
			mockTotal:     0,
			mockError:     nil,
			expectedCode:  http.StatusOK,
			expectedMsg:   "Rooms fetched successfully",
			expectedError: false,
		},
		{
			name:      "Failure - Database Error",
			mockRooms: nil,
			mockTotal: 0,
			mockError: &dtos.Error{
				Code:    "INTERNAL_ERROR",
				Message: "failed to fetch rooms",
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "failed to fetch rooms",
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(repoMock.ReservationRepositoryMock)
			mockRepo.On("GetRooms", mock.Anything).
				Return(tc.mockRooms, tc.mockTotal, tc.mockError).Once()

			svc := services.NewReservationService(mockRepo, newTestLogger(t))
			resp := svc.GetRooms(context.Background())

			assert.NotNil(t, resp)
			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			if !tc.expectedError {
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Meta)
				assert.Equal(t, tc.mockTotal, resp.Meta.Total)

				rooms, ok := resp.Data.([]dtos.GetAllRoomResponse)
				assert.True(t, ok)
				assert.Len(t, rooms, len(tc.mockRooms))

				for i, r := range tc.mockRooms {
					assert.Equal(t, int(r.ID), rooms[i].RoomID)
					assert.Equal(t, r.RoomType, rooms[i].RoomType)
					assert.Equal(t, r.Capacity, rooms[i].Capacity)
					assert.Equal(t, r.IsActive, rooms[i].IsActive)
				}
			} else {
				assert.Equal(t, "error", resp.Status)
				assert.NotEmpty(t, resp.Errors)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}


func TestGetRoomByBookingID(t *testing.T) {
	mockBookingResult := &models.Booking{
		ID: 1,
		Room: &models.Room{
			RoomType: "Deluxe",
			Capacity: 2,
			Price:    500.0,
		},
	}

	testCases := []struct {
		name          string
		bookingID     string
		mockBooking   *models.Booking
		mockError     *dtos.Error
		expectedCode  int
		expectedMsg   string
		expectedError bool
		isInternal    bool
	}{
		{
			name:          "Success - Room fetched by booking ID",
			bookingID:     "valid-booking-id",
			mockBooking:   mockBookingResult,
			mockError:     nil,
			expectedCode:  http.StatusOK,
			expectedMsg:   "Room fetched successfully",
			expectedError: false,
			isInternal:    false,
		},
		{
			name:        "Failure - Booking Not Found",
			bookingID:   "invalid-booking-id",
			mockBooking: nil,
			mockError: &dtos.Error{
				Code:    "NOT_FOUND",
				Message: "booking not found",
			},
			expectedCode:  http.StatusNotFound,
			expectedMsg:   "booking not found",
			expectedError: true,
			isInternal:    false,
		},
		{
			name:        "Failure - Database Error",
			bookingID:   "valid-booking-id",
			mockBooking: nil,
			mockError: &dtos.Error{
				Code:    "INTERNAL_ERROR",
				Message: "db error",
			},
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "booking not found",
			expectedError: true,
			isInternal:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(repoMock.ReservationRepositoryMock)
			mockRepo.On("GetRoomByBookingID", mock.Anything, tc.bookingID).
				Return(tc.mockBooking, tc.mockError).Once()

			svc := services.NewReservationService(mockRepo, newTestLogger(t))
			resp := svc.GetRoomByBookingID(context.Background(), tc.bookingID)

			assert.NotNil(t, resp)
			assert.Equal(t, tc.expectedCode, resp.Code)
			assert.Equal(t, tc.expectedMsg, resp.Message)

			if !tc.expectedError {
				assert.Equal(t, "success", resp.Status)
				roomResp, ok := resp.Data.(dtos.GetRoomResponse)
				assert.True(t, ok)
				assert.Equal(t, int(tc.mockBooking.ID), roomResp.BookingID)
				assert.Equal(t, tc.mockBooking.Room.RoomType, roomResp.RoomType)
				assert.Equal(t, tc.mockBooking.Room.Capacity, roomResp.Capacity)
				assert.Equal(t, tc.mockBooking.Room.Price, roomResp.Price)
			} else {
				if tc.isInternal {
					assert.Equal(t, "error", resp.Status)
				} else {
					assert.Equal(t, "fail", resp.Status)
				}
				assert.NotEmpty(t, resp.Errors)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}