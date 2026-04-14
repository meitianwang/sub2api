package service

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
)

// PaymentOrderExpiryService expires stale pending payment orders.
type PaymentOrderExpiryService struct {
	orderRepo    PaymentOrderRepository
	auditLogRepo PaymentAuditLogRepository
	registry     *PaymentProviderRegistry
	instanceRepo PaymentProviderInstanceRepository
	encryptor    SecretEncryptor
	orderService *PaymentOrderService // used to trigger fulfillment for orders found paid at provider
	interval     time.Duration
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
}

// NewPaymentOrderExpiryService creates a new expiry service.
func NewPaymentOrderExpiryService(
	orderRepo PaymentOrderRepository,
	auditLogRepo PaymentAuditLogRepository,
	registry *PaymentProviderRegistry,
	instanceRepo PaymentProviderInstanceRepository,
	encryptor SecretEncryptor,
	interval time.Duration,
) *PaymentOrderExpiryService {
	return &PaymentOrderExpiryService{
		orderRepo:    orderRepo,
		auditLogRepo: auditLogRepo,
		registry:     registry,
		instanceRepo: instanceRepo,
		encryptor:    encryptor,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

// SetOrderService sets the order service reference for triggering fulfillment.
// Called after both services are constructed to break the circular dependency.
func (s *PaymentOrderExpiryService) SetOrderService(os *PaymentOrderService) {
	s.orderService = os
}

// Start begins the background expiry ticker.
func (s *PaymentOrderExpiryService) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		slog.Info("payment order expiry service started", "interval", s.interval)

		// Run immediately on startup to clear any orders that expired while the service was down.
		s.runOnce()

		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

// Stop halts the background ticker and waits for completion.
func (s *PaymentOrderExpiryService) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		s.wg.Wait()
		slog.Info("payment order expiry service stopped")
	})
}

func (s *PaymentOrderExpiryService) runOnce() {
	listCtx, listCancel := context.WithTimeout(context.Background(), 10*time.Second)
	orders, err := s.orderRepo.ListExpiredPending(listCtx, time.Now(), 50)
	listCancel()
	if err != nil {
		slog.Warn("failed to list expired pending orders", "error", err)
		return
	}

	for _, order := range orders {
		// Give each order its own timeout to avoid one slow provider query starving the rest.
		orderCtx, orderCancel := context.WithTimeout(context.Background(), 15*time.Second)
		s.expireOrder(orderCtx, &order)
		orderCancel()
	}
}

func (s *PaymentOrderExpiryService) expireOrder(ctx context.Context, order *PaymentOrder) {
	// Before expiring, check with the payment provider whether the order was actually paid.
	// This prevents the race condition where a user pays just as the order is about to expire.
	if paid, notification := s.queryProviderPaymentStatus(ctx, order); paid {
		slog.Info("order appears paid at provider during expiry, triggering fulfillment", "orderID", order.ID)
		if s.orderService != nil && notification != nil {
			if err := s.orderService.handleNotificationForPendingOrder(ctx, order, notification); err != nil {
				slog.Warn("failed to fulfill paid order during expiry", "orderID", order.ID, "error", err)
			}
		}
		return
	}

	ok, err := s.orderRepo.UpdateStatusCAS(ctx, order.ID, domain.PaymentOrderStatusPending, domain.PaymentOrderStatusExpired)
	if err != nil {
		slog.Warn("failed to expire order", "orderID", order.ID, "error", err)
		return
	}
	if !ok {
		return // status already changed
	}

	// Best-effort cancel at provider
	s.tryCancelAtProvider(ctx, order)

	// Write audit log
	log := &PaymentAuditLog{
		OrderID: order.ID,
		Action:  "ORDER_EXPIRED",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to write expiry audit log", "orderID", order.ID, "error", err)
	}
}

// queryProviderPaymentStatus queries the payment provider to check if the order was paid.
// Returns (false, nil) on any error (fail-open: proceed with expiry if we can't verify).
// Returns (true, notification) if the provider reports the order as paid.
func (s *PaymentOrderExpiryService) queryProviderPaymentStatus(ctx context.Context, order *PaymentOrder) (bool, *PaymentNotification) {
	if order.ProviderInstanceID == nil {
		return false, nil
	}
	providerKey := PaymentTypeToProviderKey(order.PaymentType)
	inst, err := s.instanceRepo.GetByID(ctx, *order.ProviderInstanceID)
	if err != nil {
		return false, nil
	}
	configJSON, err := s.encryptor.Decrypt(inst.Config)
	if err != nil {
		return false, nil
	}
	provider, err := s.registry.CreateProvider(providerKey, configJSON)
	if err != nil {
		return false, nil
	}

	// QueryOrder semantics differ per provider:
	//   EasyPay/Alipay/WxPay expect our order ID (out_trade_no).
	//   Stripe expects the PaymentIntent ID (stored in PaymentTradeNo).
	// Default to our order ID; only override for Stripe.
	tradeNo := strconv.FormatInt(order.ID, 10)
	if providerKey == domain.PaymentProviderStripe && order.PaymentTradeNo != nil && *order.PaymentTradeNo != "" {
		tradeNo = *order.PaymentTradeNo
	}

	result, err := provider.QueryOrder(ctx, tradeNo)
	if err != nil {
		return false, nil
	}
	if result.Status != "paid" {
		return false, nil
	}

	return true, &PaymentNotification{
		TradeNo: result.TradeNo,
		OrderID: strconv.FormatInt(order.ID, 10),
		Amount:  result.Amount,
		Status:  "success",
	}
}

func (s *PaymentOrderExpiryService) tryCancelAtProvider(ctx context.Context, order *PaymentOrder) {
	if order.PaymentTradeNo == nil || *order.PaymentTradeNo == "" || order.ProviderInstanceID == nil {
		return
	}
	providerKey := PaymentTypeToProviderKey(order.PaymentType)
	inst, err := s.instanceRepo.GetByID(ctx, *order.ProviderInstanceID)
	if err != nil {
		return
	}
	configJSON, err := s.encryptor.Decrypt(inst.Config)
	if err != nil {
		return
	}
	provider, err := s.registry.CreateProvider(providerKey, configJSON)
	if err != nil {
		return
	}
	if cancelable, ok := provider.(CancelableProvider); ok {
		_ = cancelable.CancelPayment(ctx, *order.PaymentTradeNo)
	}
}
