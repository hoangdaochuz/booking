package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	schedulerv1 "github.com/ticketbox/pkg/proto/scheduler/v1"
)

type SchedulerHandler struct {
	schedulerClient schedulerv1.SchedulerServiceClient
}

func NewSchedulerHandler(schedulerClient schedulerv1.SchedulerServiceClient) *SchedulerHandler {
	return &SchedulerHandler{schedulerClient: schedulerClient}
}

func (h *SchedulerHandler) ListSchedulerJobs(c *gin.Context) {
	resp, err := h.schedulerClient.ListActiveSchedulers(c.Request.Context(), &schedulerv1.ListActiveSchedulersRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list scheduler jobs"})
		return
	}

	jobs := make([]gin.H, 0, len(resp.Schedulers))
	for _, j := range resp.Schedulers {
		jobs = append(jobs, gin.H{
			"id":              j.Id,
			"name":            j.Name,
			"cron_expression": j.CronExpression,
			"is_enabled":      j.IsEnabled,
			"version":         j.Version,
			"timeout":         j.Timeout,
			"created_at":      j.CreatedAt.AsTime().Format(time.RFC3339),
			"updated_at":      j.UpdatedAt.AsTime().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"schedulers": jobs})
}

func (h *SchedulerHandler) UpdateSchedulerJob(c *gin.Context) {
	var req struct {
		IsEnabled      bool   `json:"is_enable"`
		CronExpression string `json:"cron_expression" binding:"required"`
		Timeout        int32  `json:"timeout" binding:"min=1,max=600"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.schedulerClient.UpdateSchedulerById(c.Request.Context(), &schedulerv1.UpdateSchedulerByIdRequest{
		Id:             c.Param("id"),
		IsEnable:       req.IsEnabled,
		CronExpression: req.CronExpression,
		Timeout:        req.Timeout,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update scheduler job", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "scheduler job updated"})
}
