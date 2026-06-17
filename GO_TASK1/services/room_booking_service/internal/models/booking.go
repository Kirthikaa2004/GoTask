package models
 
import (
	"time"
)
 
 
type Customer struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string    `gorm:"type:uuid;uniqueIndex;not null" json:"uuid"`
	Name        string    `gorm:"type:varchar(150);not null" json:"name"`
	Age         int       `gorm:"not null" json:"age"`
	Address     string    `gorm:"type:text" json:"address"`
	PhoneNumber string    `gorm:"type:varchar(20);not null" json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   string    `gorm:"type:varchar(100)" json:"created_by"`
	UpdatedBy   string    `gorm:"type:varchar(100)" json:"updated_by"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
}
 
func (Customer) TableName() string {
	return "customers"
}
 
 
type Room struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      string    `gorm:"type:uuid;uniqueIndex;not null" json:"uuid"`
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	RoomType  string    `gorm:"type:varchar(100);not null" json:"room_type"`
	Capacity  int       `gorm:"not null" json:"capacity"`
	RoomFloor int       `gorm:"not null" json:"room_floor"`
	RoomName  string    `gorm:"type:varchar(100);not null" json:"room_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `gorm:"type:varchar(100)" json:"created_by"`
	UpdatedBy string    `gorm:"type:varchar(100)" json:"updated_by"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}
 
func (Room) TableName() string {
	return "rooms"
}
 
 
type Booking struct {
    ID         uint
    UUID       string
    CustomerID int
    RoomID     uint
    StartDate  time.Time
    EndDate    time.Time
    Purpose    string
	TotalPrice float64
    CreatedAt  time.Time
    UpdatedAt  time.Time
    CreatedBy  string
    UpdatedBy  string
    IsActive   bool
    Room       *Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
}

 
func (Booking) TableName() string {
	return "bookings"
}
 