package handlers

import (
	"net/http"

	"room_booking_management/room_booking_service/internal/config"
	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/internal/repository"
	"room_booking_management/room_booking_service/internal/services"
	"room_booking_management/room_booking_service/pkg/database"

	"github.com/gorilla/mux"
)

func SetupRoutes(router *mux.Router, db *database.Db, logger *loggers.Logger, cfg *config.Config) {

	reservationRepo := repository.NewReservationRepository(db, logger)

	reservationService := services.NewReservationService(reservationRepo, logger)

	reservationHandler := NewReservationHandler(reservationService, logger)

	api := router.PathPrefix("/api/v1").Subrouter()

	
	api.HandleFunc("/reservations", reservationHandler.ReserveRoom).Methods(http.MethodPost)
	api.HandleFunc("/reservations/{bookingID}/cancel", reservationHandler.CancelBooking).Methods(http.MethodPut)
	api.HandleFunc("/rooms", reservationHandler.GetRooms).Methods(http.MethodGet)
	api.HandleFunc("/rooms/{bookingID}", reservationHandler.GetRooms).Methods(http.MethodGet)
}