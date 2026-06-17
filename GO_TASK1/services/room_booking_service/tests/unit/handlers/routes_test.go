package handlers_test

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"room_booking_management/room_booking_service/internal/config"
	"room_booking_management/room_booking_service/internal/handlers"
	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/pkg/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func setupTestGormDB() (*gorm.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	dialector := sqlserver.New(sqlserver.Config{
		Conn:       db,
		DriverName: "sqlmock",
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	return gormDB, mock, nil
}

func setupMockSqlDB() (*sql.DB, sqlmock.Sqlmock, error) {
	return sqlmock.New()
}

type testRouter struct {
	router *mux.Router
	mock   sqlmock.Sqlmock
}

func newTestRouter(t *testing.T) *testRouter {
	t.Helper()

	gormDB, mock, err := setupTestGormDB()
	assert.NoError(t, err)

	sqlDB, _, err := setupMockSqlDB()
	assert.NoError(t, err)

	db := &database.Db{Gorm: gormDB, SqlDb: sqlDB}

	logger := loggers.NewLogger(loggers.LogConfig{
		ServiceName: "test-service",
		LogDir:      t.TempDir(),
		FileName:    "test.log",
	})

	cfg := &config.Config{
		DatabaseURL: "mock_db_url",
		Port:        "8080",
		BaseUrl:     "http://localhost:8080",
	}

	router := mux.NewRouter()
	handlers.SetupRoutes(router, db, logger, cfg)

	return &testRouter{router: router, mock: mock}
}

func (tr *testRouter) serve(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	tr.router.ServeHTTP(rr, req)
	return rr
}

func TestSetupRoutes_ReserveRoom(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name: "Success - Valid Reservation",
			payload: `{
				"name":"Alice","age":25,"address":"123 Main St",
				"phone_number":"+1234567890","purpose":"Meeting",
				"room_id":1,"room_floor":2,"room_name":"Boardroom A",
				"start_date":"2026-06-01T10:00:00Z",
				"end_date":"2026-06-05T10:00:00Z"
			}`,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("INSERT")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Failure - Malformed JSON",
			payload:        `{"name": "Alice", "age":}`,
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Failure - Missing Required Fields",
			payload:        `{"name": ""}`,
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTestRouter(t)
			tc.mockSetup(tr.mock)

			req, err := http.NewRequest(
				http.MethodPost,
				"/api/v1/reservations",
				bytes.NewBufferString(tc.payload),
			)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := tr.serve(req)
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestSetupRoutes_CancelBooking(t *testing.T) {
	tests := []struct {
		name           string
		bookingID      string
		method         string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name:      "Success - Booking Cancelled",
			bookingID: "booking-uuid",
			method:    http.MethodPut,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnRows(sqlmock.NewRows([]string{"id", "uuid", "room_id", "is_active"}).
						AddRow(1, "booking-uuid", 1, true))
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("UPDATE")).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Failure - Booking Not Found",
			bookingID: "non-existent-uuid",
			method:    http.MethodPut,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Failure - Wrong HTTP Method",
			bookingID:      "booking-uuid",
			method:         http.MethodPost,
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTestRouter(t)
			tc.mockSetup(tr.mock)

			req, err := http.NewRequest(
				tc.method,
				"/api/v1/reservations/"+tc.bookingID+"/cancel",
				nil,
			)
			assert.NoError(t, err)

			rr := tr.serve(req)
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestSetupRoutes_GetRooms(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
	}{
		{
			name:   "Success - Get All Rooms",
			path:   "/api/v1/rooms",
			method: http.MethodGet,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "uuid", "room_type", "capacity", "room_floor", "room_name", "is_active"}).
						AddRow(1, "uuid-1", "Deluxe", 2, 1, "Room A", true).
						AddRow(2, "uuid-2", "Suite", 4, 2, "Room B", true))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Success - Get Room By BookingID",
			path:   "/api/v1/rooms/booking-uuid",
			method: http.MethodGet,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "uuid", "room_id", "is_active"}).
						AddRow(1, "booking-uuid", 1, true))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Failure - Room Not Found For BookingID",
			path:   "/api/v1/rooms/ghost-uuid",
			method: http.MethodGet,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Failure - Wrong HTTP Method",
			path:           "/api/v1/rooms",
			method:         http.MethodPost,
			mockSetup:      func(mock sqlmock.Sqlmock) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTestRouter(t)
			tc.mockSetup(tr.mock)

			req, err := http.NewRequest(tc.method, tc.path, nil)
			assert.NoError(t, err)

			rr := tr.serve(req)
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestSetupRoutes_RouteRegistration(t *testing.T) {
	tr := newTestRouter(t)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/reservations"},
		{http.MethodPut, "/api/v1/reservations/any-id/cancel"},
		{http.MethodGet, "/api/v1/rooms"},
		{http.MethodGet, "/api/v1/rooms/any-id"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req, err := http.NewRequest(route.method, route.path, nil)
			assert.NoError(t, err)

			var match mux.RouteMatch
			matched := tr.router.Match(req, &match)
			assert.True(t, matched, "expected route %s %s to be registered", route.method, route.path)
		})
	}
}