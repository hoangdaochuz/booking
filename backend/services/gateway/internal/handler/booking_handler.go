package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	bookingv1 "github.com/ticketbox/pkg/proto/booking/v1"
)

type BookingHandler struct {
	bookingClient bookingv1.BookingServiceClient
}

func NewBookingHandler(bookingClient bookingv1.BookingServiceClient) *BookingHandler {
	return &BookingHandler{bookingClient: bookingClient}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		EventID string `json:"event_id" binding:"required"`
		Items   []struct {
			TicketTierID string   `json:"ticket_tier_id" binding:"required"`
			Quantity     int32    `json:"quantity" binding:"required"`
			SeatIDs      []string `json:"seat_ids"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := c.GetHeader("X-Booking-Mode")
	if mode == "" {
		mode = "pessimistic"
	}

	var items []*bookingv1.BookingItemRequest
	for _, item := range req.Items {
		items = append(items, &bookingv1.BookingItemRequest{
			TicketTierId: item.TicketTierID,
			Quantity:     item.Quantity,
			SeatIds:      item.SeatIDs,
		})
	}

	resp, err := h.bookingClient.CreateBooking(c.Request.Context(), &bookingv1.CreateBookingRequest{
		UserId:      userID.(string),
		EventId:     req.EventID,
		Items:       items,
		BookingMode: mode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "booking failed"})
		return
	}

	c.JSON(http.StatusCreated, toBookingJSON(resp))
}

func (h *BookingHandler) ListBookings(c *gin.Context) {
	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.bookingClient.ListUserBookings(c.Request.Context(), &bookingv1.ListUserBookingsRequest{
		UserId:   userID.(string),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bookings"})
		return
	}

	bookings := make([]gin.H, 0)
	for _, b := range resp.Bookings {
		bookings = append(bookings, toBookingJSON(b))
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"total":    resp.Total,
	})
}

func (h *BookingHandler) GetBooking(c *gin.Context) {
	resp, err := h.bookingClient.GetBooking(c.Request.Context(), &bookingv1.GetBookingRequest{
		BookingId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	c.JSON(http.StatusOK, toBookingJSON(resp))
}

func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resp, err := h.bookingClient.CancelBooking(c.Request.Context(), &bookingv1.CancelBookingRequest{
		BookingId: c.Param("id"),
		UserId:    userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cancel failed"})
		return
	}

	c.JSON(http.StatusOK, toBookingJSON(resp))
}

func toBookingJSON(b *bookingv1.BookingDetail) gin.H {
	items := make([]gin.H, 0)
	for _, item := range b.Items {
		items = append(items, gin.H{
			"id":               item.Id,
			"ticket_tier_id":   item.TicketTierId,
			"tier_name":        item.TierName,
			"quantity":         item.Quantity,
			"unit_price_cents": item.UnitPriceCents,
			"seat_ids":         item.SeatIds,
		})
	}
	return gin.H{
		"id":                 b.Id,
		"user_id":            b.UserId,
		"event_id":           b.EventId,
		"status":             b.Status,
		"total_amount_cents": b.TotalAmountCents,
		"items":              items,
		"created_at":         b.CreatedAt.AsTime().Format(time.RFC3339),
	}
}
