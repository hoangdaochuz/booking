package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ticketbox/payment/internal/domain"
	"github.com/ticketbox/payment/internal/repository"
	"github.com/ticketbox/payment/internal/service"
	paymentv1 "github.com/ticketbox/pkg/proto/payment/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	logger  *zap.Logger
	service *service.PaymentService
}

func NewPaymentServer(service *service.PaymentService, logger *zap.Logger) *PaymentServer {
	return &PaymentServer{
		service: service,
		logger:  logger,
	}
}

func (p *PaymentServer) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid user id")
	}
	bookingId, err := uuid.Parse(req.BookingId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid booking id")
	}
	paymentId := uuid.New()
	payment := domain.CreatePaymentRequest{
		Payment: domain.Payment{
			ID:            paymentId,
			UserId:        userId,
			BookingId:     &bookingId,
			Price:         req.Price,
			Currency:      req.Currency,
			PaymentMethod: req.PaymentMethod,
			Status:        domain.PaymentPending,
		},
		UserEmail: req.UserEmail,
	}
	res, err := p.service.CreatePayment(ctx, &payment)
	if err != nil {
		p.logger.Sugar().Errorf("fail to create payment: %w", err)
		return nil, err
	}
	return &paymentv1.CreatePaymentResponse{
		Id:                        paymentId.String(),
		Status:                    string(domain.PaymentPending),
		PaymentIntentId:           res.PaymentItentId,
		PaymentIntentClientSecret: res.PaymentIntentClientSecret,
	}, nil
}

func toPaymentEntry(payment *domain.Payment) paymentv1.PaymentEntry {
	bookingId := uuid.UUID{}
	if payment.BookingId != nil {
		bookingId = *payment.BookingId
	}
	orderId := uuid.UUID{}
	if payment.OrderId != nil {
		orderId = *payment.OrderId
	}
	transactionId := uuid.UUID{}
	if payment.Transaction_id != nil {
		transactionId = *payment.Transaction_id
	}
	return paymentv1.PaymentEntry{
		Id:            payment.ID.String(),
		UserId:        payment.UserId.String(),
		BookingId:     bookingId.String(),
		OrderId:       orderId.String(),
		Status:        string(payment.Status),
		Price:         payment.Price,
		Currency:      payment.Currency,
		TransactionId: transactionId.String(),
		PaymentMethod: payment.PaymentMethod,
		CreatedAt: &timestamppb.Timestamp{
			Seconds: int64(payment.CreatedAt.Second()),
		},
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: int64(payment.UpdateAt.Second()),
		},
	}
}

func (p *PaymentServer) GetPaymentById(ctx context.Context, req *paymentv1.GetPaymentByIdReq) (*paymentv1.PaymentEntry, error) {
	paymentId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid payment id")
	}
	res, err := p.service.GetPaymentById(ctx, paymentId)
	if err != nil {
		p.logger.Sugar().Errorf("get payment by id fail: %w", err)
		if errors.Is(err, repository.RecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, err
	}
	if res == nil {
		return nil, status.Error(codes.Internal, "get payment by id return nil")
	}
	payment := toPaymentEntry(res)
	return &payment, nil
}

func (p *PaymentServer) GetPayments(ctx context.Context, req *paymentv1.GetPaymentsReqeust) (*paymentv1.PaymentList, error) {
	return nil, nil
}

func (p *PaymentServer) UpdatePaymentStatus(ctx context.Context, req *paymentv1.UpdatePaymentStatusRequest) (*emptypb.Empty, error) {
	paymentId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid payment id")
	}
	err = p.service.UpdatePaymentStatus(ctx, paymentId, domain.PaymentStatus(req.Status))
	if err != nil {
		p.logger.Sugar().Errorf("fail to update payment status: %w", err)
		return nil, err
	}
	return nil, nil
}

func (p *PaymentServer) UpdatePayment(ctx context.Context, req *paymentv1.UpdatePaymentRequest) (*emptypb.Empty, error) {

	orderId, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Order Id")
	}
	paymentId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid payment id")
	}
	transactionId, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid transaction id")
	}
	payment := domain.Payment{
		OrderId:        &orderId,
		Transaction_id: &transactionId,
		ID:             paymentId,
	}
	err = p.service.UpdatePayment(ctx, &payment)
	if err != nil {
		p.logger.Sugar().Errorf("fail to update payment: %w", err)
		return nil, err
	}
	return nil, nil
}

func (p *PaymentServer) DeletePayment(ctx context.Context, req *paymentv1.DeletePaymentRequest) (*emptypb.Empty, error) {
	paymentId, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid payment id")
	}
	err = p.service.SoftDeletePayment(ctx, paymentId)
	if err != nil {
		p.logger.Sugar().Errorf("fail to soft delete payment: %w", err)
		return nil, err
	}
	return nil, nil
}
