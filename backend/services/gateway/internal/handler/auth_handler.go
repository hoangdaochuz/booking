package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	userv1 "github.com/ticketbox/pkg/proto/user/v1"
)

type AuthHandler struct {
	userClient  userv1.UserServiceClient
	redisClient *redis.Client
}

func NewAuthHandler(userClient userv1.UserServiceClient, redisClient *redis.Client) *AuthHandler {
	return &AuthHandler{userClient: userClient, redisClient: redisClient}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.Register(c.Request.Context(), &userv1.RegisterRequest{
		Email: req.Email, Password: req.Password, Name: req.Name,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          toUserJSON(resp.User),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.Login(c.Request.Context(), &userv1.LoginRequest{
		Email: req.Email, Password: req.Password,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          toUserJSON(resp.User),
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.userClient.RefreshToken(c.Request.Context(), &userv1.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired or invalid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user":          toUserJSON(resp.User),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	h.redisClient.Set(c.Request.Context(), "blacklist:"+token, "1", 24*time.Hour)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func toUserJSON(u *userv1.UserProfile) gin.H {
	if u == nil {
		return nil
	}
	return gin.H{
		"id":         u.Id,
		"email":      u.Email,
		"name":       u.Name,
		"role":       u.Role,
		"created_at": u.CreatedAt.AsTime().Format(time.RFC3339),
	}
}
