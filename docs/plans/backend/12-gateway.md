# Task 12: API Gateway — REST Endpoints

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the REST API gateway that translates HTTP requests to gRPC calls, with JWT auth middleware, CORS, and rate limiting.

**Depends on:** Tasks 8, 9, 10, 11 (all gRPC services must be defined).

**Files:**
- Create: `backend/services/gateway/internal/middleware/auth.go`
- Create: `backend/services/gateway/internal/middleware/cors.go`
- Create: `backend/services/gateway/internal/middleware/ratelimit.go`
- Create: `backend/services/gateway/internal/handler/auth_handler.go`
- Create: `backend/services/gateway/internal/handler/event_handler.go`
- Create: `backend/services/gateway/internal/handler/booking_handler.go`
- Create: `backend/services/gateway/internal/handler/user_handler.go`
- Create: `backend/services/gateway/internal/router/router.go`
- Modify: `backend/services/gateway/cmd/main.go`

---

### Step 1: Create auth middleware

`backend/services/gateway/internal/middleware/auth.go`:
```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"

    userv1 "github.com/ticketbox/pkg/proto/user/v1"
)

func AuthMiddleware(userClient userv1.UserServiceClient, redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        blacklisted, _ := redisClient.Exists(c.Request.Context(), "blacklist:"+token).Result()
        if blacklisted > 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
            return
        }

        resp, err := userClient.ValidateToken(c.Request.Context(), &userv1.ValidateTokenRequest{
            AccessToken: token,
        })
        if err != nil || !resp.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }

        c.Set("user_id", resp.UserId)
        c.Set("email", resp.Email)
        c.Set("role", resp.Role)
        c.Next()
    }
}

func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, _ := c.Get("role")
        if role != "admin" {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
            return
        }
        c.Next()
    }
}
```

### Step 2: Create CORS middleware

`backend/services/gateway/internal/middleware/cors.go`:
```go
package middleware

import "github.com/gin-gonic/gin"

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Booking-Mode")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

### Step 3: Create REST handlers

Each handler follows this pattern: parse HTTP request → call gRPC client → return JSON response.

**Auth handler** (`auth_handler.go`):
- `POST /api/auth/register` → `userClient.Register()`
- `POST /api/auth/login` → `userClient.Login()`
- `POST /api/auth/refresh` → `userClient.RefreshToken()`
- `POST /api/auth/logout` → blacklist token in Redis

**Event handler** (`event_handler.go`):
- `GET /api/events` → `eventClient.ListEvents()`
- `GET /api/events/:id` → `eventClient.GetEvent()`
- `POST /api/events` → `eventClient.CreateEvent()` (admin)
- `PUT /api/events/:id` → `eventClient.UpdateEvent()` (admin)
- `DELETE /api/events/:id` → `eventClient.DeleteEvent()` (admin)

**Booking handler** (`booking_handler.go`):
- `POST /api/bookings` → `bookingClient.CreateBooking()` — passes `X-Booking-Mode` header
- `GET /api/bookings` → `bookingClient.ListUserBookings()`
- `GET /api/bookings/:id` → `bookingClient.GetBooking()`
- `POST /api/bookings/:id/cancel` → `bookingClient.CancelBooking()`

**User handler** (`user_handler.go`):
- `GET /api/users/me` → `userClient.GetProfile()`
- `PUT /api/users/me` → `userClient.UpdateProfile()`

### Step 4: Create router

`backend/services/gateway/internal/router/router.go`:
```go
func SetupRouter(
    userClient userv1.UserServiceClient,
    eventClient eventv1.EventServiceClient,
    bookingClient bookingv1.BookingServiceClient,
    redisClient *redis.Client,
) *gin.Engine {
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(middleware.CORSMiddleware())

    authHandler := handler.NewAuthHandler(userClient, redisClient)
    eventHandler := handler.NewEventHandler(eventClient)
    bookingHandler := handler.NewBookingHandler(bookingClient)
    userHandler := handler.NewUserHandler(userClient)

    api := r.Group("/api")
    {
        auth := api.Group("/auth")
        auth.POST("/register", authHandler.Register)
        auth.POST("/login", authHandler.Login)
        auth.POST("/refresh", authHandler.RefreshToken)

        api.GET("/events", eventHandler.ListEvents)
        api.GET("/events/:id", eventHandler.GetEvent)

        protected := api.Group("")
        protected.Use(middleware.AuthMiddleware(userClient, redisClient))
        {
            protected.POST("/auth/logout", authHandler.Logout)
            protected.GET("/users/me", userHandler.GetProfile)
            protected.PUT("/users/me", userHandler.UpdateProfile)

            protected.POST("/bookings", bookingHandler.CreateBooking)
            protected.GET("/bookings", bookingHandler.ListBookings)
            protected.GET("/bookings/:id", bookingHandler.GetBooking)
            protected.POST("/bookings/:id/cancel", bookingHandler.CancelBooking)

            admin := protected.Group("")
            admin.Use(middleware.AdminOnly())
            admin.POST("/events", eventHandler.CreateEvent)
            admin.PUT("/events/:id", eventHandler.UpdateEvent)
            admin.DELETE("/events/:id", eventHandler.DeleteEvent)
        }
    }

    return r
}
```

### Step 5: Wire up main.go

- Connect to User Service (gRPC), Event Service (gRPC), Booking Service (gRPC)
- Connect to Redis
- Setup router
- Start Gin on port 8000
- Graceful shutdown

### Step 6: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/gateway
go get github.com/gin-gonic/gin
go get github.com/redis/go-redis/v9
go get google.golang.org/grpc
go mod tidy
go build ./...
```
Expected: No errors.

### Step 7: Commit

```bash
git add backend/services/gateway/
git commit -m "feat(gateway): add REST API gateway with auth, CORS, and routing"
```
