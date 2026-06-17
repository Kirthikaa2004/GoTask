package services

import (
	"context"
	"net/http"

	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/internal/models"
	"room_booking_management/room_booking_service/internal/repository"
	"room_booking_management/room_booking_service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)


type ReservationServiceInterface interface {
	ReserveRoom(ctx context.Context, req dtos.CreateBookingRequest) *dtos.APIResponse
	CancelBooking(ctx context.Context, bookingID string) *dtos.APIResponse
	GetRooms(ctx context.Context) *dtos.APIResponse
	GetRoomByBookingID(ctx context.Context, bookingID string) *dtos.APIResponse
}


type ReservationService struct {
	reservationRepo repository.ReservationRepositoryInterface
	logger          *loggers.Logger
}

func NewReservationService(
	reservationRepo repository.ReservationRepositoryInterface,
	logger *loggers.Logger,
) ReservationServiceInterface {
	return &ReservationService{
		reservationRepo: reservationRepo,
		logger:          logger,
	}
}


func reqIDFromCtx(ctx context.Context) string {
	reqID, _ := ctx.Value("requestID").(string)
	return reqID
}


func (s *ReservationService) ReserveRoom(ctx context.Context, req dtos.CreateBookingRequest) *dtos.APIResponse {
	reqID := reqIDFromCtx(ctx)

	if req.EndDate.Before(req.StartDate) || req.EndDate.Equal(req.StartDate) {
		return utils.NewErrorResponse(http.StatusBadRequest, "end_date must be after start_date", []dtos.Error{
			{Field: "end_date", Message: "must be after start_date", Code: "VALIDATION_ERROR_400"},
		}, reqID)
	}

	room, dtoErr := s.reservationRepo.GetRoomByID(ctx, req.RoomID)
	if dtoErr != nil {
		s.logger.Error("ReserveRoom: room lookup failed", zap.String("error", dtoErr.Message))
		statusCode := http.StatusInternalServerError
		if dtoErr.Code == "NOT_FOUND_404" {
			statusCode = http.StatusNotFound
		}
		return utils.NewErrorResponse(statusCode, "room not found", []dtos.Error{*dtoErr}, reqID)
	}

	available, dtoErr := s.reservationRepo.CheckRoomAvailability(ctx, req.RoomID, req.StartDate, req.EndDate)
	if dtoErr != nil {
		s.logger.Error("ReserveRoom: availability check failed", zap.String("error", dtoErr.Message))
		return utils.NewErrorResponse(http.StatusInternalServerError, "failed to check room availability", []dtos.Error{*dtoErr}, reqID)
	}
	if !available {
		return utils.NewErrorResponse(http.StatusConflict, "room is already booked for the requested period", []dtos.Error{
			{Field: "room_id", Message: "overlapping booking exists", Code: "CONFLICT_409"},
		}, reqID)
	}

	days := int(req.EndDate.Sub(req.StartDate).Hours() / 24)
	totalPrice := room.Price * float64(days)

	s.logger.Info("Price calculated",
		zap.Int("days", days),
		zap.Float64("room_price", room.Price),
		zap.Float64("total_price", totalPrice),
	)

	customer, dtoErr := s.reservationRepo.GetCustomerByPhone(ctx, req.PhoneNumber)
	if dtoErr != nil {
		s.logger.Error("ReserveRoom: phone lookup failed",
			zap.String("phone", req.PhoneNumber),
			zap.String("error", dtoErr.Message),
		)
		return utils.NewErrorResponse(http.StatusInternalServerError, "failed to lookup customer", []dtos.Error{*dtoErr}, reqID)
	}

	if customer != nil {
		s.logger.Info("Existing customer found, reusing record",
			zap.String("phone", req.PhoneNumber),
			zap.Uint("customer_id", customer.ID),
		)
	} else {
		s.logger.Info("No existing customer, creating new record",
			zap.String("phone", req.PhoneNumber),
		)
		customer = &models.Customer{
			UUID:        uuid.NewString(),
			Name:        req.Name,
			Age:         req.Age,
			Address:     req.Address,
			PhoneNumber: req.PhoneNumber,
			IsActive:    true,
			CreatedBy:   "system",
			UpdatedBy:   "system",
		}
		if dtoErr := s.reservationRepo.CreateCustomer(ctx, customer); dtoErr != nil {
			s.logger.Error("ReserveRoom: CreateCustomer failed", zap.String("error", dtoErr.Message))
			return utils.NewErrorResponse(http.StatusInternalServerError, "failed to create customer", []dtos.Error{*dtoErr}, reqID)
		}
	}

	booking := &models.Booking{
		UUID:       uuid.NewString(),
		CustomerID: int(customer.ID),
		RoomID:     req.RoomID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Purpose:    req.Purpose,
		TotalPrice: totalPrice,
		IsActive:   true,
		CreatedBy:  "system",
		UpdatedBy:  "system",
	}
	if dtoErr := s.reservationRepo.CreateBooking(ctx, booking); dtoErr != nil {
		s.logger.Error("ReserveRoom: CreateBooking failed", zap.String("error", dtoErr.Message))
		return utils.NewErrorResponse(http.StatusInternalServerError, "failed to create booking", []dtos.Error{*dtoErr}, reqID)
	}

	s.logger.Info("Room reserved successfully",
		zap.String("booking_uuid", booking.UUID),
		zap.Uint("room_id", req.RoomID),
		zap.Uint("customer_id", customer.ID),
		zap.Float64("total_price", totalPrice),
	)

	return utils.NewSuccessResponse(http.StatusCreated, "Room reserved successfully", dtos.RoomResponse{
		BookingID:  int(booking.ID),
		RoomID:     int(req.RoomID),
		Status:     "CONFIRMED",
		TotalPrice: totalPrice,
	}, reqID)
}


func (s *ReservationService) CancelBooking(ctx context.Context, bookingID string) *dtos.APIResponse {
	reqID := reqIDFromCtx(ctx)

	if dtoErr := s.reservationRepo.CancelBooking(ctx, bookingID); dtoErr != nil {
		s.logger.Error("CancelBooking failed",
			zap.String("bookingID", bookingID),
			zap.String("error", dtoErr.Message),
		)
		statusCode := http.StatusInternalServerError
		if dtoErr.Code == "NOT_FOUND" {
			statusCode = http.StatusNotFound
		}
		return utils.NewErrorResponse(statusCode, "failed to cancel booking", []dtos.Error{*dtoErr}, reqID)
	}

	s.logger.Info("Booking cancelled", zap.String("bookingID", bookingID))
	return utils.NewSuccessResponse(http.StatusOK, "Booking cancelled successfully", dtos.CancelRoomResponse{
		Status: "CANCELLED",
	}, reqID)
}


func (s *ReservationService) GetRooms(ctx context.Context) *dtos.APIResponse {
	reqID := reqIDFromCtx(ctx)

	rooms, total, dtoErr := s.reservationRepo.GetRooms(ctx)
	if dtoErr != nil {
		s.logger.Error("GetRooms failed", zap.String("error", dtoErr.Message))
		return utils.NewErrorResponse(http.StatusInternalServerError, "failed to fetch rooms", []dtos.Error{*dtoErr}, reqID)
	}

	result := make([]dtos.GetAllRoomResponse, 0, len(rooms))
	for _, r := range rooms {
		result = append(result, dtos.GetAllRoomResponse{
			RoomID:    int(r.ID),
			RoomName:  r.RoomName,
			RoomType:  r.RoomType,
			Capacity:  r.Capacity,
			RoomFloor: r.RoomFloor,
			IsActive:  r.IsActive,
		})
	}

	resp := utils.NewSuccessResponse(http.StatusOK, "Rooms fetched successfully", result, reqID)
	resp.Meta = &dtos.Meta{Total: total}
	return resp
}


func (s *ReservationService) GetRoomByBookingID(ctx context.Context, bookingID string) *dtos.APIResponse {
	reqID := reqIDFromCtx(ctx)

	booking, dtoErr := s.reservationRepo.GetRoomByBookingID(ctx, bookingID)
	if dtoErr != nil {
		s.logger.Error("GetRoomByBookingID failed",
			zap.String("bookingID", bookingID),
			zap.String("error", dtoErr.Message),
		)
		statusCode := http.StatusInternalServerError
		if dtoErr.Code == "NOT_FOUND" {
			statusCode = http.StatusNotFound
		}
		return utils.NewErrorResponse(statusCode, "booking not found", []dtos.Error{*dtoErr}, reqID)
	}

	return utils.NewSuccessResponse(http.StatusOK, "Room fetched successfully", dtos.GetRoomResponse{
		BookingID: int(booking.ID),
		RoomName:  booking.Room.RoomName,
		RoomType:  booking.Room.RoomType,
		RoomFloor: booking.Room.RoomFloor,
		Capacity:  booking.Room.Capacity,
		Price:     booking.Room.Price, 
	}, reqID)
}