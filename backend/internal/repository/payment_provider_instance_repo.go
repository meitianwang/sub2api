package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentproviderinstance"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentProviderInstanceRepository struct {
	client *dbent.Client
}

func NewPaymentProviderInstanceRepository(client *dbent.Client) service.PaymentProviderInstanceRepository {
	return &paymentProviderInstanceRepository{client: client}
}

func (r *paymentProviderInstanceRepository) Create(ctx context.Context, inst *service.PaymentProviderInstance) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.PaymentProviderInstance.Create().
		SetProviderKey(inst.ProviderKey).
		SetName(inst.Name).
		SetConfig(inst.Config).
		SetSupportedTypes(inst.SupportedTypes).
		SetEnabled(inst.Enabled).
		SetSortOrder(inst.SortOrder).
		SetNillableLimits(inst.Limits).
		SetRefundEnabled(inst.RefundEnabled).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}
	inst.ID = created.ID
	inst.CreatedAt = created.CreatedAt
	inst.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *paymentProviderInstanceRepository) GetByID(ctx context.Context, id int64) (*service.PaymentProviderInstance, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentProviderInstance.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPaymentProviderNotFound, nil)
	}
	return providerInstanceToService(m), nil
}

func (r *paymentProviderInstanceRepository) Update(ctx context.Context, inst *service.PaymentProviderInstance) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.PaymentProviderInstance.UpdateOneID(inst.ID).
		SetProviderKey(inst.ProviderKey).
		SetName(inst.Name).
		SetConfig(inst.Config).
		SetSupportedTypes(inst.SupportedTypes).
		SetEnabled(inst.Enabled).
		SetSortOrder(inst.SortOrder).
		SetNillableLimits(inst.Limits).
		SetRefundEnabled(inst.RefundEnabled).
		Save(ctx)
	return translatePersistenceError(err, service.ErrPaymentProviderNotFound, nil)
}

func (r *paymentProviderInstanceRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.PaymentProviderInstance.DeleteOneID(id).Exec(ctx)
}

func (r *paymentProviderInstanceRepository) List(ctx context.Context) ([]service.PaymentProviderInstance, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentProviderInstance.Query().
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return providerInstancesToService(ms), nil
}

func (r *paymentProviderInstanceRepository) ListEnabled(ctx context.Context) ([]service.PaymentProviderInstance, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentProviderInstance.Query().
		Where(paymentproviderinstance.EnabledEQ(true)).
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return providerInstancesToService(ms), nil
}

func (r *paymentProviderInstanceRepository) ListEnabledByProviderKey(ctx context.Context, providerKey string) ([]service.PaymentProviderInstance, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentProviderInstance.Query().
		Where(
			paymentproviderinstance.ProviderKeyEQ(providerKey),
			paymentproviderinstance.EnabledEQ(true),
		).
		Order(paymentproviderinstance.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return providerInstancesToService(ms), nil
}

func providerInstanceToService(m *dbent.PaymentProviderInstance) *service.PaymentProviderInstance {
	if m == nil {
		return nil
	}
	return &service.PaymentProviderInstance{
		ID:             m.ID,
		ProviderKey:    m.ProviderKey,
		Name:           m.Name,
		Config:         m.Config,
		SupportedTypes: m.SupportedTypes,
		Enabled:        m.Enabled,
		SortOrder:      m.SortOrder,
		Limits:         m.Limits,
		RefundEnabled:  m.RefundEnabled,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func providerInstancesToService(ms []*dbent.PaymentProviderInstance) []service.PaymentProviderInstance {
	result := make([]service.PaymentProviderInstance, len(ms))
	for i, m := range ms {
		result[i] = *providerInstanceToService(m)
	}
	return result
}
