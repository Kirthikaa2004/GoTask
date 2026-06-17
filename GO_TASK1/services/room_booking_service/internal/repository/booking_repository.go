package repository

import (
	"context"
	"errors"
	"fmt"

	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/pkg/database"
	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)


type ReservationRepositoryInterface interface {
	GetCustomerByPhone(ctx context.Context, phone string) (*models.Customer, *dtos.Error)
	CreateCustomer(ctx context.Context, customer *models.Customer) *dtos.Error
	GetRoomByID(ctx context.Context, roomID uint) (*models.Room, *dtos.Error)
	CheckRoomAvailability(ctx context.Context, roomID uint, startDate, endDate interface{}) (bool, *dtos.Error)
	CreateBooking(ctx context.Context, booking *models.Booking) *dtos.Error
	CancelBooking(ctx context.Context, bookingID string) *dtos.Error
	GetRooms(ctx context.Context) ([]models.Room, int64, *dtos.Error)
	GetRoomByBookingID(ctx context.Context, bookingID string) (*models.Booking, *dtos.Error)
}


type ReservationRepository struct {
	db     *database.Db
	logger *loggers.Logger
}

func NewReservationRepository(db *database.Db, logger *loggers.Logger) ReservationRepositoryInterface {
	return &ReservationRepository{db: db, logger: logger}
}


func (r *ReservationRepository) GetCustomerByPhone(ctx context.Context, phone string) (*models.Customer, *dtos.Error) {
	var customer models.Customer
	err := r.db.Gorm.WithContext(ctx).
		Where("phone_number = ? AND is_active = ?", phone, true).
		First(&customer).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("GetCustomerByPhone failed", zap.String("phone", phone), zap.Error(err))
		return nil, &dtos.Error{Field: "phone_number", Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	return &customer, nil
}

func (r *ReservationRepository) CreateCustomer(ctx context.Context, customer *models.Customer) *dtos.Error {
	if result := r.db.Gorm.WithContext(ctx).Create(customer); result.Error != nil {
		r.logger.Error("CreateCustomer failed", zap.Error(result.Error))
		return &dtos.Error{Field: "customer", Message: result.Error.Error(), Code: "INTERNAL_ERROR_500"}
	}
	return nil
}
func (r *ReservationRepository) GetRoomByID(ctx context.Context, roomID uint) (*models.Room, *dtos.Error) {
	var room models.Room
	err := r.db.Gorm.WithContext(ctx).
		Where("id = ? AND is_active = ?", roomID, true).
		First(&room).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dtos.Error{
				Field:   "room_id",
				Message: fmt.Sprintf("room %d not found or inactive", roomID),
				Code:    "NOT_FOUND",
			}
		}
		r.logger.Error("GetRoomByID failed", zap.Uint("roomID", roomID), zap.Error(err))
		return nil, &dtos.Error{Field: "room_id", Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	return &room, nil
}


func (r *ReservationRepository) CheckRoomAvailability(ctx context.Context, roomID uint, startDate, endDate interface{}) (bool, *dtos.Error) {
	var count int64
	err := r.db.Gorm.WithContext(ctx).
		Model(&models.Booking{}).
		Where(
			"room_id = ? AND is_active = ? AND start_date < ? AND end_date > ?",
			roomID, true, endDate, startDate,
		).
		Count(&count).Error

	if err != nil {
		r.logger.Error("CheckRoomAvailability failed", zap.Uint("roomID", roomID), zap.Error(err))
		return false, &dtos.Error{Field: "room_id", Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	return count == 0, nil
}

func (r *ReservationRepository) CreateBooking(ctx context.Context, booking *models.Booking) *dtos.Error {
	if result := r.db.Gorm.WithContext(ctx).Create(booking); result.Error != nil {
		r.logger.Error("CreateBooking failed", zap.Error(result.Error))
		return &dtos.Error{Field: "booking", Message: result.Error.Error(), Code: "INTERNAL_ERROR_500"}
		
	}

	return nil
}


func (r *ReservationRepository) CancelBooking(ctx context.Context, bookingID string) *dtos.Error {
	var booking models.Booking
	err := r.db.Gorm.WithContext(ctx).
		Where("uuid = ? AND is_active = ?", bookingID, true).
		First(&booking).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dtos.Error{
				Field:   "booking_id",
				Message: fmt.Sprintf("booking %s not found or already cancelled", bookingID),
				Code:    "NOT_FOUND",
			}
		}
		r.logger.Error("CancelBooking lookup failed", zap.String("bookingID", bookingID), zap.Error(err))
		return &dtos.Error{Field: "booking_id", Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	if err := r.db.Gorm.WithContext(ctx).Model(&booking).Update("is_active", false).Error; err != nil {
		r.logger.Error("CancelBooking update failed", zap.String("bookingID", bookingID), zap.Error(err))
		return &dtos.Error{Field: "booking_id", Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	return nil
}

func (r *ReservationRepository) GetRooms(ctx context.Context) ([]models.Room, int64, *dtos.Error) {
	var rooms []models.Room
	var total int64

	if err := r.db.Gorm.WithContext(ctx).Model(&models.Room{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		r.logger.Error("GetRooms count failed", zap.Error(err))
		return nil, 0, &dtos.Error{Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	if err := r.db.Gorm.WithContext(ctx).Where("is_active = ?", true).Find(&rooms).Error; err != nil {
		r.logger.Error("GetRooms query failed", zap.Error(err))
		return nil, 0, &dtos.Error{Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	return rooms, total, nil
}

func (r *ReservationRepository) GetRoomByBookingID(ctx context.Context, bookingID string) (*models.Booking, *dtos.Error) {
	var booking models.Booking
	err := r.db.Gorm.WithContext(ctx).
		Preload("Room").
		Where("uuid = ? AND is_active = ?", bookingID, true).
		First(&booking).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &dtos.Error{
				Field:   "booking_id",
				Message: fmt.Sprintf("booking %s not found", bookingID),
				Code:    "NOT_FOUND",
			}
		}
		r.logger.Error("GetRoomByBookingID failed", zap.String("bookingID", bookingID), zap.Error(err))
		return nil, &dtos.Error{Field: "booking_id", Message: err.Error(), Code: "INTERNAL_ERROR_500"}
	}

	return &booking, nil
}