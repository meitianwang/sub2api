package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentAuditLogRepository struct {
	client *dbent.Client
}

func NewPaymentAuditLogRepository(client *dbent.Client) service.PaymentAuditLogRepository {
	return &paymentAuditLogRepository{client: client}
}

func (r *paymentAuditLogRepository) Create(ctx context.Context, log *service.PaymentAuditLog) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.PaymentAuditLog.Create().
		SetOrderID(log.OrderID).
		SetAction(log.Action).
		SetNillableDetail(log.Detail).
		SetNillableOperator(log.Operator).
		Save(ctx)
	if err != nil {
		return err
	}
	log.ID = created.ID
	log.CreatedAt = created.CreatedAt
	return nil
}

func (r *paymentAuditLogRepository) CountByActionAndOperatorSince(ctx context.Context, action, operator string, since time.Time) (int, error) {
	client := clientFromContext(ctx, r.client)
	return client.PaymentAuditLog.Query().
		Where(
			paymentauditlog.ActionEQ(action),
			paymentauditlog.OperatorEQ(operator),
			paymentauditlog.CreatedAtGTE(since),
		).
		Count(ctx)
}

func (r *paymentAuditLogRepository) ListByOrderID(ctx context.Context, orderID int64) ([]service.PaymentAuditLog, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(orderID)).
		Order(paymentauditlog.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]service.PaymentAuditLog, len(ms))
	for i, m := range ms {
		result[i] = service.PaymentAuditLog{
			ID:        m.ID,
			OrderID:   m.OrderID,
			Action:    m.Action,
			Detail:    m.Detail,
			Operator:  m.Operator,
			CreatedAt: m.CreatedAt,
		}
	}
	return result, nil
}
