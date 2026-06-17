package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/internal/models"
	"room_booking_management/room_booking_service/internal/repository"
	"room_booking_management/room_booking_service/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type ReservationRepositoryTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	mock       sqlmock.Sqlmock
	repository repository.ReservationRepositoryInterface
	logger     *loggers.Logger
}

func (suite *ReservationRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	suite.NoError(err)

	dialector := sqlserver.New(sqlserver.Config{
		Conn:       db,
		DriverName: "sqlmock",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	suite.NoError(err)

	dbMock := &database.Db{
		Gorm:  gormDB,
		SqlDb: db,
	}

	logger := loggers.NewLogger(loggers.LogConfig{
		ServiceName: "test-service",
		LogDir:      suite.T().TempDir(),
		FileName:    "test.log",
	})

	suite.DB = gormDB
	suite.mock = mock
	suite.repository = repository.NewReservationRepository(dbMock, logger)
	suite.logger = logger
}

func (suite *ReservationRepositoryTestSuite) TearDownTest() {
	suite.logger.Sync()
}

func (suite *ReservationRepositoryTestSuite) TestGetCustomerByPhone() {
	phone := "+1234567890"
	columns := []string{"id", "uuid", "name", "phone_number", "is_active", "created_at", "updated_at"}
	now := time.Now()

	suite.T().Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(1, "customer-uuid", "Alice", phone, true, now, now)
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

		customer, repoErr := suite.repository.GetCustomerByPhone(context.Background(), phone)

		assert.Nil(t, repoErr)
		assert.NotNil(t, customer)
		assert.Equal(t, phone, customer.PhoneNumber)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found - Returns nil customer and nil error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(gorm.ErrRecordNotFound)

		customer, repoErr := suite.repository.GetCustomerByPhone(context.Background(), "unknown-phone")

		assert.Nil(t, repoErr)
		assert.Nil(t, customer)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(sqlmock.ErrCancelled)

		customer, repoErr := suite.repository.GetCustomerByPhone(context.Background(), phone)

		assert.Nil(t, customer)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestCreateCustomer() {
	customer := &models.Customer{
		Name:        "Alice",
		PhoneNumber: "+1234567890",
		IsActive:    true,
	}

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectBegin()
		suite.mock.ExpectQuery("INSERT INTO").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		suite.mock.ExpectCommit()

		repoErr := suite.repository.CreateCustomer(context.Background(), customer)

		assert.Nil(t, repoErr)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectBegin()
		suite.mock.ExpectQuery("INSERT INTO").WillReturnError(sqlmock.ErrCancelled)
		suite.mock.ExpectRollback()

		repoErr := suite.repository.CreateCustomer(context.Background(), customer)

		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.Equal(t, "customer", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestGetRoomByID() {
	roomID := uint(1)
	columns := []string{"id", "uuid", "name", "room_type", "price", "is_active", "created_at", "updated_at"}
	now := time.Now()

	suite.T().Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(1, "room-uuid", "Deluxe Suite", "deluxe", 250.00, true, now, now)
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

		room, repoErr := suite.repository.GetRoomByID(context.Background(), roomID)

		assert.Nil(t, repoErr)
		assert.NotNil(t, room)
		assert.Equal(t, uint(1), room.ID)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(gorm.ErrRecordNotFound)

		room, repoErr := suite.repository.GetRoomByID(context.Background(), 999)

		assert.Nil(t, room)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "NOT_FOUND", repoErr.Code)
		assert.Equal(t, "room_id", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(sqlmock.ErrCancelled)

		room, repoErr := suite.repository.GetRoomByID(context.Background(), roomID)

		assert.Nil(t, room)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestCheckRoomAvailability() {
	roomID := uint(1)
	startDate := time.Now()
	endDate := startDate.Add(48 * time.Hour)

	suite.T().Run("Available - No Overlapping Bookings", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		available, repoErr := suite.repository.CheckRoomAvailability(context.Background(), roomID, startDate, endDate)

		assert.Nil(t, repoErr)
		assert.True(t, available)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Unavailable - Overlapping Booking Exists", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		available, repoErr := suite.repository.CheckRoomAvailability(context.Background(), roomID, startDate, endDate)

		assert.Nil(t, repoErr)
		assert.False(t, available)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(sqlmock.ErrCancelled)

		available, repoErr := suite.repository.CheckRoomAvailability(context.Background(), roomID, startDate, endDate)

		assert.False(t, available)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.Equal(t, "room_id", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestCreateBooking() {
	booking := &models.Booking{
		UUID:      "booking-uuid",
		RoomID:    1,
		IsActive:  true,
		StartDate: time.Now(),
		EndDate:   time.Now().Add(48 * time.Hour),
	}

	suite.T().Run("Success", func(t *testing.T) {
		suite.mock.ExpectBegin()
		suite.mock.ExpectQuery("INSERT INTO").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		suite.mock.ExpectCommit()

		repoErr := suite.repository.CreateBooking(context.Background(), booking)

		assert.Nil(t, repoErr)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectBegin()
		suite.mock.ExpectQuery("INSERT INTO").WillReturnError(sqlmock.ErrCancelled)
		suite.mock.ExpectRollback()

		repoErr := suite.repository.CreateBooking(context.Background(), booking)

		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.Equal(t, "booking", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestCancelBooking() {
	bookingID := "booking-uuid"
	columns := []string{"id", "uuid", "room_id", "is_active", "start_date", "end_date", "created_at", "updated_at"}
	now := time.Now()

	suite.T().Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(1, bookingID, 1, true, now, now.Add(48*time.Hour), now, now)
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
		suite.mock.ExpectBegin()
		suite.mock.ExpectExec(regexp.QuoteMeta("UPDATE")).
			WillReturnResult(sqlmock.NewResult(1, 1))
		suite.mock.ExpectCommit()

		repoErr := suite.repository.CancelBooking(context.Background(), bookingID)

		assert.Nil(t, repoErr)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(gorm.ErrRecordNotFound)

		repoErr := suite.repository.CancelBooking(context.Background(), "non-existent-uuid")

		assert.NotNil(t, repoErr)
		assert.Equal(t, "NOT_FOUND", repoErr.Code)
		assert.Equal(t, "booking_id", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Lookup Database Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(sqlmock.ErrCancelled)

		repoErr := suite.repository.CancelBooking(context.Background(), bookingID)

		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Update Database Error", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(1, bookingID, 1, true, now, now.Add(48*time.Hour), now, now)
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
		suite.mock.ExpectBegin()
		suite.mock.ExpectExec(regexp.QuoteMeta("UPDATE")).
			WillReturnError(sqlmock.ErrCancelled)
		suite.mock.ExpectRollback()

		repoErr := suite.repository.CancelBooking(context.Background(), bookingID)

		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.Equal(t, "booking_id", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestGetRooms() {
	columns := []string{"id", "uuid", "name", "room_type", "price", "is_active", "created_at", "updated_at"}
	now := time.Now()

	suite.T().Run("Success - With Rooms", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*)")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnRows(sqlmock.NewRows(columns).
				AddRow(1, "room-uuid-1", "Deluxe Suite", "deluxe", 250.00, true, now, now).
				AddRow(2, "room-uuid-2", "Standard Room", "standard", 100.00, true, now, now))

		rooms, total, repoErr := suite.repository.GetRooms(context.Background())

		assert.Nil(t, repoErr)
		assert.Equal(t, int64(2), total)
		assert.Len(t, rooms, 2)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Success - Empty List", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*)")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnRows(sqlmock.NewRows(columns))

		rooms, total, repoErr := suite.repository.GetRooms(context.Background())

		assert.Nil(t, repoErr)
		assert.Equal(t, int64(0), total)
		assert.Len(t, rooms, 0)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Count Query Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*)")).
			WillReturnError(sqlmock.ErrCancelled)

		rooms, total, repoErr := suite.repository.GetRooms(context.Background())

		assert.Nil(t, rooms)
		assert.Equal(t, int64(0), total)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Select Query Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*)")).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(sqlmock.ErrCancelled)

		rooms, total, repoErr := suite.repository.GetRooms(context.Background())

		assert.Nil(t, rooms)
		assert.Equal(t, int64(0), total)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func (suite *ReservationRepositoryTestSuite) TestGetRoomByBookingID() {
	bookingID := "booking-uuid"
	columns := []string{"id", "uuid", "room_id", "is_active", "start_date", "end_date", "created_at", "updated_at"}
	roomColumns := []string{"id", "uuid", "name", "room_type", "price", "is_active", "created_at", "updated_at"}
	now := time.Now()

	suite.T().Run("Success", func(t *testing.T) {
		bookingRows := sqlmock.NewRows(columns).
			AddRow(1, bookingID, 1, true, now, now.Add(48*time.Hour), now, now)
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(bookingRows)

		roomRows := sqlmock.NewRows(roomColumns).
			AddRow(1, "room-uuid", "Deluxe Suite", "deluxe", 250.00, true, now, now)
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(roomRows)

		booking, repoErr := suite.repository.GetRoomByBookingID(context.Background(), bookingID)

		assert.Nil(t, repoErr)
		assert.NotNil(t, booking)
		assert.Equal(t, bookingID, booking.UUID)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Not Found", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(gorm.ErrRecordNotFound)

		booking, repoErr := suite.repository.GetRoomByBookingID(context.Background(), "ghost-uuid")

		assert.Nil(t, booking)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "NOT_FOUND", repoErr.Code)
		assert.Equal(t, "booking_id", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})

	suite.T().Run("Database Error", func(t *testing.T) {
		suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
			WillReturnError(sqlmock.ErrCancelled)

		booking, repoErr := suite.repository.GetRoomByBookingID(context.Background(), bookingID)

		assert.Nil(t, booking)
		assert.NotNil(t, repoErr)
		assert.Equal(t, "INTERNAL_ERROR", repoErr.Code)
		assert.Equal(t, "booking_id", repoErr.Field)
		assert.NoError(t, suite.mock.ExpectationsWereMet())
	})
}

func TestReservationRepositorySuite(t *testing.T) {
	suite.Run(t, new(ReservationRepositoryTestSuite))
}