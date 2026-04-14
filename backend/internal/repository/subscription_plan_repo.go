package repository

import (
	"context"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionplan"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionPlanRepository struct {
	client *dbent.Client
}

func NewSubscriptionPlanRepository(client *dbent.Client) service.SubscriptionPlanRepository {
	return &subscriptionPlanRepository{client: client}
}

func (r *subscriptionPlanRepository) Create(ctx context.Context, plan *service.SubscriptionPlan) error {
	client := clientFromContext(ctx, r.client)
	builder := client.SubscriptionPlan.Create().
		SetName(plan.Name).
		SetNillableDescription(plan.Description).
		SetPrice(plan.Price).
		SetNillableOriginalPrice(plan.OriginalPrice).
		SetValidityDays(plan.ValidityDays).
		SetValidityUnit(plan.ValidityUnit).
		SetNillableFeatures(plan.Features).
		SetNillableProductName(plan.ProductName).
		SetForSale(plan.ForSale).
		SetSortOrder(plan.SortOrder)
	if plan.GroupID != nil {
		builder.SetGroupID(*plan.GroupID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}
	plan.ID = created.ID
	plan.CreatedAt = created.CreatedAt
	plan.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *subscriptionPlanRepository) GetByID(ctx context.Context, id int64) (*service.SubscriptionPlan, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrSubscriptionPlanNotFound, nil)
	}
	return subscriptionPlanToService(m), nil
}

func (r *subscriptionPlanRepository) Update(ctx context.Context, plan *service.SubscriptionPlan) error {
	client := clientFromContext(ctx, r.client)
	builder := client.SubscriptionPlan.UpdateOneID(plan.ID).
		SetName(plan.Name).
		SetNillableDescription(plan.Description).
		SetPrice(plan.Price).
		SetNillableOriginalPrice(plan.OriginalPrice).
		SetValidityDays(plan.ValidityDays).
		SetValidityUnit(plan.ValidityUnit).
		SetNillableFeatures(plan.Features).
		SetNillableProductName(plan.ProductName).
		SetForSale(plan.ForSale).
		SetSortOrder(plan.SortOrder)
	if plan.GroupID != nil {
		builder.SetGroupID(*plan.GroupID)
	} else {
		builder.ClearGroupID()
	}
	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrSubscriptionPlanNotFound, nil)
}

func (r *subscriptionPlanRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.SubscriptionPlan.DeleteOneID(id).Exec(ctx)
}

func (r *subscriptionPlanRepository) ListForSale(ctx context.Context) ([]service.SubscriptionPlan, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.SubscriptionPlan.Query().
		Where(subscriptionplan.ForSaleEQ(true)).
		Order(subscriptionplan.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return subscriptionPlansToService(ms), nil
}

func (r *subscriptionPlanRepository) List(ctx context.Context) ([]service.SubscriptionPlan, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.SubscriptionPlan.Query().
		Order(subscriptionplan.BySortOrder()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return subscriptionPlansToService(ms), nil
}

func subscriptionPlanToService(m *dbent.SubscriptionPlan) *service.SubscriptionPlan {
	if m == nil {
		return nil
	}
	return &service.SubscriptionPlan{
		ID:            m.ID,
		GroupID:       m.GroupID,
		Name:          m.Name,
		Description:   m.Description,
		Price:         m.Price,
		OriginalPrice: m.OriginalPrice,
		ValidityDays:  m.ValidityDays,
		ValidityUnit:  m.ValidityUnit,
		Features:      m.Features,
		ProductName:   m.ProductName,
		ForSale:       m.ForSale,
		SortOrder:     m.SortOrder,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func subscriptionPlansToService(ms []*dbent.SubscriptionPlan) []service.SubscriptionPlan {
	result := make([]service.SubscriptionPlan, len(ms))
	for i, m := range ms {
		result[i] = *subscriptionPlanToService(m)
	}
	return result
}
