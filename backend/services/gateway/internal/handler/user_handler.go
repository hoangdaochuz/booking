package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	userv1 "github.com/ticketbox/pkg/proto/user/v1"
)

type UserHandler struct {
	userClient userv1.UserServiceClient
}

func NewUserHandler(userClient userv1.UserServiceClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	resp, err := h.userClient.GetProfile(c.Request.Context(), &userv1.GetProfileRequest{
		UserId: userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, toUserJSON(resp))
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.UpdateProfile(c.Request.Context(), &userv1.UpdateProfileRequest{
		UserId: userID.(string),
		Name:   req.Name,
		Email:  req.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, toUserJSON(resp))
}
