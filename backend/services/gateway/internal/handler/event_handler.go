package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventv1 "github.com/ticketbox/pkg/proto/event/v1"
)

type EventHandler struct {
	eventClient eventv1.EventServiceClient
}

func NewEventHandler(eventClient eventv1.EventServiceClient) *EventHandler {
	return &EventHandler{eventClient: eventClient}
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.eventClient.ListEvents(c.Request.Context(), &eventv1.ListEventsRequest{
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Page:     int32(page),
		PageSize: int32(pageSize),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list events"})
		return
	}

	events := make([]gin.H, 0)
	for _, e := range resp.Events {
		events = append(events, toEventJSON(e))
	}

	c.JSON(http.StatusOK, gin.H{
		"events":    events,
		"total":     resp.Total,
		"page":      resp.Page,
		"page_size": resp.PageSize,
	})
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	resp, err := h.eventClient.GetEvent(c.Request.Context(), &eventv1.GetEventRequest{
		EventId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, toEventJSON(resp))
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Category    string `json:"category" binding:"required"`
		Venue       string `json:"venue" binding:"required"`
		Location    string `json:"location" binding:"required"`
		Date        string `json:"date" binding:"required"`
		ImageURL    string `json:"image_url"`
		Tiers       []struct {
			Name          string `json:"name" binding:"required"`
			PriceCents    int64  `json:"price_cents" binding:"required"`
			TotalQuantity int32  `json:"total_quantity" binding:"required"`
		} `json:"tiers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use RFC3339"})
		return
	}

	var tiers []*eventv1.CreateTicketTierRequest
	for _, t := range req.Tiers {
		tiers = append(tiers, &eventv1.CreateTicketTierRequest{
			Name: t.Name, PriceCents: t.PriceCents, TotalQuantity: t.TotalQuantity,
		})
	}

	resp, err := h.eventClient.CreateEvent(c.Request.Context(), &eventv1.CreateEventRequest{
		Title: req.Title, Description: req.Description, Category: req.Category,
		Venue: req.Venue, Location: req.Location, Date: timestamppb.New(date),
		ImageUrl: req.ImageURL, Tiers: tiers,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, toEventJSON(resp))
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Venue       string `json:"venue"`
		Location    string `json:"location"`
		Date        string `json:"date"`
		ImageURL    string `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateReq := &eventv1.UpdateEventRequest{
		EventId: c.Param("id"), Title: req.Title, Description: req.Description,
		Category: req.Category, Venue: req.Venue, Location: req.Location, ImageUrl: req.ImageURL,
	}
	if req.Date != "" {
		date, _ := time.Parse(time.RFC3339, req.Date)
		updateReq.Date = timestamppb.New(date)
	}

	resp, err := h.eventClient.UpdateEvent(c.Request.Context(), updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update event"})
		return
	}

	c.JSON(http.StatusOK, toEventJSON(resp))
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	_, err := h.eventClient.DeleteEvent(c.Request.Context(), &eventv1.DeleteEventRequest{
		EventId: c.Param("id"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event deleted"})
}

func toEventJSON(e *eventv1.EventDetail) gin.H {
	tiers := make([]gin.H, 0)
	for _, t := range e.Tiers {
		tiers = append(tiers, gin.H{
			"id":                 t.Id,
			"event_id":           t.EventId,
			"name":               t.Name,
			"price_cents":        t.PriceCents,
			"total_quantity":     t.TotalQuantity,
			"available_quantity": t.AvailableQuantity,
			"version":            t.Version,
		})
	}
	return gin.H{
		"id":          e.Id,
		"title":       e.Title,
		"description": e.Description,
		"category":    e.Category,
		"venue":       e.Venue,
		"location":    e.Location,
		"date":        e.Date.AsTime().Format(time.RFC3339),
		"image_url":   e.ImageUrl,
		"status":      e.Status,
		"tiers":       tiers,
		"created_at":  e.CreatedAt.AsTime().Format(time.RFC3339),
	}
}
