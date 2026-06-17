package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"room_booking_management/room_booking_service/internal/dtos"
	"room_booking_management/room_booking_service/internal/loggers"
	"room_booking_management/room_booking_service/internal/services"
	"room_booking_management/room_booking_service/internal/utils"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)


type ReservationHandlerInterface interface {
	ReserveRoom(w http.ResponseWriter, r *http.Request)
	CancelBooking(w http.ResponseWriter, r *http.Request)
	GetRooms(w http.ResponseWriter, r *http.Request)
}


type ReservationHandler struct {
	reservationService services.ReservationServiceInterface
	logger             *loggers.Logger
}

func NewReservationHandler(reservationService services.ReservationServiceInterface,logger *loggers.Logger,) ReservationHandlerInterface {
	return &ReservationHandler{
		reservationService: reservationService,
		logger:             logger,
	}
}


func (h *ReservationHandler) ReserveRoom(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetOrGenerateRequestID(r)
	ctx := context.WithValue(r.Context(), "requestID", reqID)

	var req dtos.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("ReserveRoom: invalid request body", zap.Error(err))
		resp := utils.NewErrorResponse(http.StatusBadRequest, "invalid request body", []dtos.Error{
			{Field: "body", Message: err.Error(), Code: "BAD_REQUEST_400"},
		}, reqID)
		utils.WriteResponse(w, http.StatusBadRequest, resp)
		return
	}

	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		resp := utils.NewErrorResponse(http.StatusBadRequest, "validation failed", errs, reqID)
		utils.WriteResponse(w, http.StatusBadRequest, resp)
		return
	}

	resp := h.reservationService.ReserveRoom(ctx, req)
	utils.WriteResponse(w, resp.Code, resp)
}


func (h *ReservationHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetOrGenerateRequestID(r)
	ctx := context.WithValue(r.Context(), "requestID", reqID)

	vars := mux.Vars(r)
	bookingID := vars["bookingID"]
	if bookingID == "" {
		resp := utils.NewErrorResponse(http.StatusBadRequest, "booking_id is required", []dtos.Error{
			{Field: "bookingID", Message: "missing in path", Code: "BAD_REQUEST_400"},
		}, reqID)
		utils.WriteResponse(w, http.StatusBadRequest, resp)
		return
	}

	resp := h.reservationService.CancelBooking(ctx, bookingID)
	utils.WriteResponse(w, resp.Code, resp)
}

func (h *ReservationHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetOrGenerateRequestID(r)
	ctx := context.WithValue(r.Context(), "requestID", reqID)

	vars := mux.Vars(r)
	bookingID := vars["bookingID"]

	if bookingID != "" {
		resp := h.reservationService.GetRoomByBookingID(ctx, bookingID)
		utils.WriteResponse(w, resp.Code, resp)
		return
	}

	resp := h.reservationService.GetRooms(ctx)
	utils.WriteResponse(w, resp.Code, resp)
}