package mock

import (
	"context"

	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/models"

	"github.com/stretchr/testify/mock"
)

type ReservationRepositoryMock struct {
	mock.Mock
}

func (m *ReservationRepositoryMock) GetCustomerByPhone(ctx context.Context, phone string) (*models.Customer, *dtos.Error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, castDtoError(args.Get(1))
	}
	return args.Get(0).(*models.Customer), castDtoError(args.Get(1))
}

func (m *ReservationRepositoryMock) CreateCustomer(ctx context.Context, customer *models.Customer) *dtos.Error {
	args := m.Called(ctx, customer)
	return castDtoError(args.Get(0))
}

func (m *ReservationRepositoryMock) GetRoomByID(ctx context.Context, roomID uint) (*models.Room, *dtos.Error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, castDtoError(args.Get(1))
	}
	return args.Get(0).(*models.Room), castDtoError(args.Get(1))
}

func (m *ReservationRepositoryMock) CheckRoomAvailability(ctx context.Context, roomID uint, startDate, endDate interface{}) (bool, *dtos.Error) {
	args := m.Called(ctx, roomID, startDate, endDate)
	return args.Bool(0), castDtoError(args.Get(1))
}

func (m *ReservationRepositoryMock) CreateBooking(ctx context.Context, booking *models.Booking) *dtos.Error {
	args := m.Called(ctx, booking)
	return castDtoError(args.Get(0))
}

func (m *ReservationRepositoryMock) CancelBooking(ctx context.Context, bookingID string) *dtos.Error {
	args := m.Called(ctx, bookingID)
	return castDtoError(args.Get(0))
}

func (m *ReservationRepositoryMock) GetRooms(ctx context.Context) ([]models.Room, int64, *dtos.Error) {
	args := m.Called(ctx)
	total := args.Get(1).(int64)
	if args.Get(0) == nil {
		return nil, total, castDtoError(args.Get(2))
	}
	return args.Get(0).([]models.Room), total, castDtoError(args.Get(2))
}

func (m *ReservationRepositoryMock) GetRoomByBookingID(ctx context.Context, bookingID string) (*models.Booking, *dtos.Error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, castDtoError(args.Get(1))
	}
	return args.Get(0).(*models.Booking), castDtoError(args.Get(1))
}

// castDtoError safely casts interface{} to *dtos.Error
func castDtoError(v interface{}) *dtos.Error {
	if v == nil {
		return nil
	}
	return v.(*dtos.Error)
}