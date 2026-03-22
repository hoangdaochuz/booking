package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
)

type PaymentHandler struct {
	client paymentv1.PaymentServiceClient
}

func NewPaymentHandler(client paymentv1.PaymentServiceClient) *PaymentHandler {
	return &PaymentHandler{
		client: client,
	}
}

func (p *PaymentHandler) SearchPayments(ctx *gin.Context) {
	//TODO
	// return nil
}

func (p *PaymentHandler) CreatePayment(ctx *gin.Context) {
	userId, _ := ctx.Get("user_id")
	userEmail, _ := ctx.Get("email")

	var req struct {
		BookingId     string `json:"booking_id" binding:"required"`
		Price         int32  `json:"price" binding:"required"`
		Currency      string `json:"currency" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	res, err := p.client.CreatePayment(ctx.Request.Context(), &paymentv1.CreatePaymentRequest{
		UserId:        userId.(string),
		BookingId:     req.BookingId,
		Price:         req.Price,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		UserEmail:     userEmail.(string),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to create payment: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":     res.Id,
		"status": res.Status,
	})
}

func (p *PaymentHandler) GetPaymentById(ctx *gin.Context) {
	paymentId := ctx.Param("id")
	res, err := p.client.GetPaymentById(ctx.Request.Context(), &paymentv1.GetPaymentByIdReq{
		Id: paymentId,
	})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":             res.Id,
		"booking_id":     res.BookingId,
		"user_id":        res.UserId,
		"transaction_id": res.TransactionId,
		"price":          res.Price,
		"price_method":   res.PaymentMethod,
		"status":         res.Status,
		"currency":       res.Currency,
		"created_at":     res.CreatedAt,
		"updated_at":     res.UpdatedAt,
	})
}

func (p *PaymentHandler) UpdatePayment(ctx *gin.Context) {
	var req struct {
		Status        string `json:"status"`
		OrderId       string `json:"order_id"`
		TransactionId string `json:"transaction_id"`
		Id            string `json:"id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	_, err := p.client.UpdatePayment(ctx.Request.Context(), &paymentv1.UpdatePaymentRequest{
		Id:            req.Id,
		OrderId:       req.OrderId,
		TransactionId: req.TransactionId,
		Status:        req.Status,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to update payment " + err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "payment update successfully",
	})
}

func (p *PaymentHandler) DeletePayment(ctx *gin.Context) {
	paymentId := ctx.Param("id")
	_, err := p.client.DeletePayment(ctx.Request.Context(), &paymentv1.DeletePaymentRequest{
		Id: paymentId,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to delete payment: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "delete payment successfully"})
}
