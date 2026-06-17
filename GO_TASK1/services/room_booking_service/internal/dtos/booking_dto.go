package dtos
 
import "time"
 
type CreateBookingRequest struct {
	Name        string    `json:"name"         validate:"required"`
	Age         int       `json:"age"          validate:"required,min=1"`
	Address     string    `json:"address"      validate:"required"`
	PhoneNumber string    `json:"phone_number" validate:"required"`
	Purpose     string    `json:"purpose"      validate:"required"`
	RoomID      uint      `json:"room_id"      validate:"required"`
	RoomFloor   uint      `json:"room_floor"   validate:"required"`
	RoomName    string    `json:"room_name"    validate:"required"`
	StartDate   time.Time `json:"start_date"   validate:"required"`
	EndDate     time.Time `json:"end_date"     validate:"required"`
}
 
type CancelBookingRequest struct {
	RoomID    int  `json:"room_id" validate:"required"`
}
 
 
type RoomResponse struct {
	BookingID  int     `json:"booking_id"`
	RoomID     int     `json:"room_id"`
	Status     string  `json:"status"`
	TotalPrice float64 `json:"total_price"`
}
 
type CancelRoomResponse struct {
	Status string `json:"status"`
}
 
type GetAllRoomResponse struct {
	RoomID   int    `json:"room_id"`
	RoomName string  `json:"room_name"`
	RoomType string `json:"room_type"`
	Capacity int    `json:"capacity"`
	RoomFloor int    `json:"room_floor"`
	IsActive bool   `json:"is_active"`
}
 
type GetRoomResponse struct {
	BookingID    int     `json:"booking_id"`
	Capacity     int     `json:"capacity"`
	RoomType     string  `json:"room_type"`
	RoomName     string   `json:"room_name"`
    RoomFloor    int     `json:"room_floor"`
	Price        float64 `json:"price"`
}
 