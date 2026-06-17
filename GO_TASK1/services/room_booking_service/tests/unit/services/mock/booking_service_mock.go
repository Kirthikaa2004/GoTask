
package mock

import (
    "context"

    "room_booking_management/room_booking_service/internal/dtos"

    "github.com/stretchr/testify/mock"
)

type ReservationServiceMock struct {
    mock.Mock
}

func (m *ReservationServiceMock) ReserveRoom(ctx context.Context, req dtos.CreateBookingRequest) *dtos.APIResponse {
    args := m.Called(ctx, req)
    return args.Get(0).(*dtos.APIResponse)
}

func (m *ReservationServiceMock) CancelBooking(ctx context.Context, bookingID string) *dtos.APIResponse {
    args := m.Called(ctx, bookingID)
    return args.Get(0).(*dtos.APIResponse)
}

func (m *ReservationServiceMock) GetRooms(ctx context.Context) *dtos.APIResponse {
    args := m.Called(ctx)
    return args.Get(0).(*dtos.APIResponse)
}

func (m *ReservationServiceMock) GetRoomByBookingID(ctx context.Context, bookingID string) *dtos.APIResponse {
    args := m.Called(ctx, bookingID)
    return args.Get(0).(*dtos.APIResponse)
}