package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentchannel"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type paymentChannelRepository struct {
	client *dbent.Client
}

func NewPaymentChannelRepository(client *dbent.Client) service.PaymentChannelRepository {
	return &paymentChannelRepository{client: client}
}

func (r *paymentChannelRepository) Create(ctx context.Context, ch *service.PaymentChannel) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentChannel.Create().
		SetName(ch.Name).
		SetPlatform(ch.Platform).
		SetRateMultiplier(ch.RateMultiplier).
		SetNillableDescription(ch.Description).
		SetNillableModels(ch.Models).
		SetNillableFeatures(ch.Features).
		SetSortOrder(ch.SortOrder).
		SetEnabled(ch.Enabled)
	if ch.GroupID != nil {
		builder.SetGroupID(*ch.GroupID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}
	ch.ID = created.ID
	ch.CreatedAt = created.CreatedAt
	ch.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *paymentChannelRepository) GetByID(ctx context.Context, id int64) (*service.PaymentChannel, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentChannel.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPaymentChannelNotFound, nil)
	}
	return paymentChannelToService(m), nil
}

func (r *paymentChannelRepository) GetByGroupID(ctx context.Context, groupID int64) (*service.PaymentChannel, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentChannel.Query().
		Where(paymentchannel.GroupIDEQ(groupID)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPaymentChannelNotFound, nil)
	}
	return paymentChannelToService(m), nil
}

func (r *paymentChannelRepository) Update(ctx context.Context, ch *service.PaymentChannel) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentChannel.UpdateOneID(ch.ID).
		SetName(ch.Name).
		SetPlatform(ch.Platform).
		SetRateMultiplier(ch.RateMultiplier).
		SetNillableDescription(ch.Description).
		SetNillableModels(ch.Models).
		SetNillableFeatures(ch.Features).
		SetSortOrder(ch.SortOrder).
		SetEnabled(ch.Enabled)
	if ch.GroupID != nil {
		builder.SetGroupID(*ch.GroupID)
	} else {
		builder.ClearGroupID()
	}
	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrPaymentChannelNotFound, nil)
}

func (r *paymentChannelRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.PaymentChannel.DeleteOneID(id).Exec(ctx)
}

func (r *paymentChannelRepository) ListEnabled(ctx context.Context) ([]service.PaymentChannel, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentChannel.Query().
		Where(paymentchannel.EnabledEQ(true)).
		Order(paymentchannel.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return paymentChannelsToService(ms), nil
}

func (r *paymentChannelRepository) List(ctx context.Context) ([]service.PaymentChannel, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentChannel.Query().
		Order(paymentchannel.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return paymentChannelsToService(ms), nil
}

func paymentChannelToService(m *dbent.PaymentChannel) *service.PaymentChannel {
	if m == nil {
		return nil
	}
	return &service.PaymentChannel{
		ID:             m.ID,
		GroupID:        m.GroupID,
		Name:           m.Name,
		Platform:       m.Platform,
		RateMultiplier: m.RateMultiplier,
		Description:    m.Description,
		Models:         m.Models,
		Features:       m.Features,
		SortOrder:      m.SortOrder,
		Enabled:        m.Enabled,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func paymentChannelsToService(ms []*dbent.PaymentChannel) []service.PaymentChannel {
	result := make([]service.PaymentChannel, len(ms))
	for i, m := range ms {
		result[i] = *paymentChannelToService(m)
	}
	return result
}
