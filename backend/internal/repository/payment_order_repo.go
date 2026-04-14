package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

type paymentOrderRepository struct {
	client *dbent.Client
	sql    *sql.DB
}

func NewPaymentOrderRepository(client *dbent.Client, sqlDB *sql.DB) service.PaymentOrderRepository {
	return &paymentOrderRepository{client: client, sql: sqlDB}
}

func (r *paymentOrderRepository) Create(ctx context.Context, order *service.PaymentOrder) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentOrder.Create().
		SetUserID(order.UserID).
		SetNillableUserEmail(order.UserEmail).
		SetNillableUserName(order.UserName).
		SetNillableUserNotes(order.UserNotes).
		SetAmount(order.Amount).
		SetNillablePayAmount(order.PayAmount).
		SetNillableFeeRate(order.FeeRate).
		SetRechargeCode(order.RechargeCode).
		SetStatus(order.Status).
		SetPaymentType(order.PaymentType).
		SetNillablePaymentTradeNo(order.PaymentTradeNo).
		SetNillablePayURL(order.PayURL).
		SetNillableQrCode(order.QrCode).
		SetExpiresAt(order.ExpiresAt).
		SetNillableClientIP(order.ClientIP).
		SetNillableSrcHost(order.SrcHost).
		SetNillableSrcURL(order.SrcURL).
		SetOrderType(order.OrderType).
		SetNillablePlanID(order.PlanID).
		SetNillableSubscriptionGroupID(order.SubscriptionGroupID).
		SetNillableSubscriptionDays(order.SubscriptionDays).
		SetNillableProviderInstanceID(order.ProviderInstanceID)

	created, err := builder.Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}
	applyPaymentOrderEntity(order, created)
	return nil
}

func (r *paymentOrderRepository) Delete(ctx context.Context, id int64) error {
	client := clientFromContext(ctx, r.client)
	return client.PaymentOrder.DeleteOneID(id).Exec(ctx)
}

func (r *paymentOrderRepository) GetByID(ctx context.Context, id int64) (*service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentOrder.Get(ctx, id)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPaymentOrderNotFound, nil)
	}
	return paymentOrderToService(m), nil
}

func (r *paymentOrderRepository) GetByRechargeCode(ctx context.Context, code string) (*service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.PaymentOrder.Query().
		Where(paymentorder.RechargeCodeEQ(code)).
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrPaymentOrderNotFound, nil)
	}
	return paymentOrderToService(m), nil
}

func (r *paymentOrderRepository) Update(ctx context.Context, order *service.PaymentOrder) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentOrder.UpdateOneID(order.ID).
		SetStatus(order.Status).
		SetNillablePaymentTradeNo(order.PaymentTradeNo).
		SetNillablePayURL(order.PayURL).
		SetNillableQrCode(order.QrCode).
		SetNillableRefundAmount(order.RefundAmount).
		SetNillableRefundReason(order.RefundReason).
		SetNillableRefundAt(order.RefundAt).
		SetForceRefund(order.ForceRefund).
		SetNillableRefundRequestedAt(order.RefundRequestedAt).
		SetNillableRefundRequestReason(order.RefundRequestReason).
		SetNillableRefundRequestedBy(order.RefundRequestedBy).
		SetNillablePaidAt(order.PaidAt).
		SetNillableCompletedAt(order.CompletedAt).
		SetNillableFailedAt(order.FailedAt).
		SetNillableFailedReason(order.FailedReason).
		SetNillableProviderInstanceID(order.ProviderInstanceID)

	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrPaymentOrderNotFound, nil)
}

func (r *paymentOrderRepository) UpdateRechargeCode(ctx context.Context, id int64, code string) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.PaymentOrder.UpdateOneID(id).
		SetRechargeCode(code).
		Save(ctx)
	return translatePersistenceError(err, service.ErrPaymentOrderNotFound, nil)
}

func (r *paymentOrderRepository) UpdateStatusCAS(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	client := clientFromContext(ctx, r.client)
	n, err := client.PaymentOrder.Update().
		Where(paymentorder.IDEQ(id), paymentorder.StatusEQ(fromStatus)).
		SetStatus(toStatus).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *paymentOrderRepository) ConfirmPaidCAS(ctx context.Context, id int64, tradeNo string, paidAmount decimal.Decimal, paidAt time.Time, graceDeadline time.Time) (bool, error) {
	client := clientFromContext(ctx, r.client)
	n, err := client.PaymentOrder.Update().
		Where(
			paymentorder.IDEQ(id),
			paymentorder.Or(
				paymentorder.StatusEQ("pending"),
				paymentorder.And(
					paymentorder.StatusEQ("expired"),
					paymentorder.UpdatedAtGTE(graceDeadline),
				),
			),
		).
		SetStatus("paid").
		SetPaymentTradeNo(tradeNo).
		SetPaidAt(paidAt).
		ClearFailedAt().
		ClearFailedReason().
		Save(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *paymentOrderRepository) UpdatePaymentResult(ctx context.Context, id int64, tradeNo string, payURL, qrCode *string, instanceID *int64) error {
	client := clientFromContext(ctx, r.client)
	builder := client.PaymentOrder.UpdateOneID(id).
		SetPaymentTradeNo(tradeNo)
	if payURL != nil {
		builder.SetPayURL(*payURL)
	}
	if qrCode != nil {
		builder.SetQrCode(*qrCode)
	}
	if instanceID != nil {
		builder.SetProviderInstanceID(*instanceID)
	}
	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrPaymentOrderNotFound, nil)
}

func (r *paymentOrderRepository) CountPendingByUserID(ctx context.Context, userID int64) (int, error) {
	client := clientFromContext(ctx, r.client)
	return client.PaymentOrder.Query().
		Where(paymentorder.UserIDEQ(userID), paymentorder.StatusEQ("pending")).
		Count(ctx)
}

// activeOrderStatuses are the statuses that indicate an order is still in-flight.
var activeOrderStatuses = []string{"pending", "paid", "recharging"}

func (r *paymentOrderRepository) CountActiveByProviderInstanceID(ctx context.Context, instanceID int64) (int, error) {
	client := clientFromContext(ctx, r.client)
	return client.PaymentOrder.Query().
		Where(
			paymentorder.ProviderInstanceIDEQ(instanceID),
			paymentorder.StatusIn(activeOrderStatuses...),
		).
		Count(ctx)
}

func (r *paymentOrderRepository) CountActiveByPlanID(ctx context.Context, planID int64) (int, error) {
	client := clientFromContext(ctx, r.client)
	return client.PaymentOrder.Query().
		Where(
			paymentorder.PlanIDEQ(planID),
			paymentorder.StatusIn(activeOrderStatuses...),
		).
		Count(ctx)
}

// SumDailyPaidByUserID sums base recharge amounts (amount, not pay_amount) for user daily limit checks.
// User-facing limits are denominated in recharge value, not the fee-inclusive amount charged at the gateway.
func (r *paymentOrderRepository) SumDailyPaidByUserID(ctx context.Context, userID int64, bizDayStart time.Time) (decimal.Decimal, error) {
	return r.sumAmount(ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM payment_orders WHERE user_id = $1 AND paid_at >= $2 AND status IN ('paid', 'recharging', 'completed') ",
		userID, bizDayStart,
	)
}

// SumDailyPaidByPaymentType sums base recharge amounts for global per-method daily limit checks.
func (r *paymentOrderRepository) SumDailyPaidByPaymentType(ctx context.Context, paymentType string, bizDayStart time.Time) (decimal.Decimal, error) {
	return r.sumAmount(ctx,
		"SELECT COALESCE(SUM(amount), 0) FROM payment_orders WHERE payment_type = $1 AND paid_at >= $2 AND status IN ('paid', 'recharging', 'completed') ",
		paymentType, bizDayStart,
	)
}

// SumDailyPaidByInstanceID sums fee-inclusive amounts (pay_amount) for provider instance daily limit checks.
// Instance limits track actual money flowing through the payment gateway, which includes fees.
func (r *paymentOrderRepository) SumDailyPaidByInstanceID(ctx context.Context, instanceID int64, bizDayStart time.Time) (decimal.Decimal, error) {
	return r.sumAmount(ctx,
		"SELECT COALESCE(SUM(pay_amount), 0) FROM payment_orders WHERE provider_instance_id = $1 AND paid_at >= $2 AND status IN ('paid', 'recharging', 'completed') ",
		instanceID, bizDayStart,
	)
}

// SumDailyPaidByInstanceIDs is the batch version of SumDailyPaidByInstanceID (uses pay_amount).
func (r *paymentOrderRepository) SumDailyPaidByInstanceIDs(ctx context.Context, instanceIDs []int64, bizDayStart time.Time) (map[int64]decimal.Decimal, error) {
	result := make(map[int64]decimal.Decimal, len(instanceIDs))
	if len(instanceIDs) == 0 {
		return result, nil
	}

	// Build placeholders: $1, $2, ..., $N for instance IDs; $N+1 for bizDayStart
	placeholders := make([]string, len(instanceIDs))
	args := make([]any, 0, len(instanceIDs)+1)
	for i, id := range instanceIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, id)
	}
	args = append(args, bizDayStart)
	tsPlaceholder := fmt.Sprintf("$%d", len(instanceIDs)+1)

	query := fmt.Sprintf(
		"SELECT provider_instance_id, COALESCE(SUM(pay_amount), 0) FROM payment_orders WHERE provider_instance_id IN (%s) AND paid_at >= %s AND status IN ('paid', 'recharging', 'completed') GROUP BY provider_instance_id",
		strings.Join(placeholders, ","), tsPlaceholder,
	)

	rows, err := r.sql.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var instID int64
		var s string
		if err := rows.Scan(&instID, &s); err != nil {
			return nil, err
		}
		d, _ := decimal.NewFromString(s)
		result[instID] = d
	}
	return result, rows.Err()
}

func (r *paymentOrderRepository) ListExpiredPending(ctx context.Context, now time.Time, limit int) ([]service.PaymentOrder, error) {
	client := clientFromContext(ctx, r.client)
	ms, err := client.PaymentOrder.Query().
		Where(
			paymentorder.StatusEQ("pending"),
			paymentorder.ExpiresAtLT(now),
		).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return paymentOrdersToService(ms), nil
}

func (r *paymentOrderRepository) List(ctx context.Context, params pagination.PaginationParams) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.PaymentOrder.Query().Order(dbent.Desc(paymentorder.FieldID))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	ms, err := query.
		Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return paymentOrdersToService(ms), &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (r *paymentOrderRepository) ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.PaymentOrder.Query().
		Where(paymentorder.UserIDEQ(userID)).
		Order(dbent.Desc(paymentorder.FieldID))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	ms, err := query.
		Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return paymentOrdersToService(ms), &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (r *paymentOrderRepository) ListFiltered(ctx context.Context, filter service.PaymentOrderListFilter, params pagination.PaginationParams) ([]service.PaymentOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	query := client.PaymentOrder.Query().Order(dbent.Desc(paymentorder.FieldID))

	if filter.UserID != nil {
		query = query.Where(paymentorder.UserIDEQ(*filter.UserID))
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where(paymentorder.StatusEQ(*filter.Status))
	}
	if filter.OrderType != nil && *filter.OrderType != "" {
		query = query.Where(paymentorder.OrderTypeEQ(*filter.OrderType))
	}
	if filter.PaymentType != nil && *filter.PaymentType != "" {
		query = query.Where(paymentorder.PaymentTypeEQ(*filter.PaymentType))
	}
	if filter.DateFrom != nil {
		query = query.Where(paymentorder.CreatedAtGTE(*filter.DateFrom))
	}
	if filter.DateTo != nil {
		query = query.Where(paymentorder.CreatedAtLTE(*filter.DateTo))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	ms, err := query.
		Offset((params.Page - 1) * params.PageSize).
		Limit(params.PageSize).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return paymentOrdersToService(ms), &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func (r *paymentOrderRepository) GetDashboardStats(ctx context.Context, since time.Time) (*service.RawDashboardData, error) {
	now := time.Now()
	// Use Asia/Shanghai timezone for business day
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*3600)
	}
	todayStart := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc).UTC()

	// Paid statuses used for revenue calculations (hardcoded, not user input).
	paidStatusList := []any{"paid", "recharging", "completed", "refunding", "refunded", "refund_failed"}
	statusPlaceholders := "$2, $3, $4, $5, $6, $7"

	data := &service.RawDashboardData{}

	// Today stats — $1=todayStart, $2..$7=statuses
	todayArgs := append([]any{todayStart}, paidStatusList...)
	row := r.sql.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM payment_orders WHERE paid_at >= $1 AND status IN ("+statusPlaceholders+")",
		todayArgs...)
	var todayAmountStr string
	if err := row.Scan(&todayAmountStr, &data.TodayOrderCount); err != nil {
		return nil, err
	}
	data.TodayAmount, _ = decimal.NewFromString(todayAmountStr)

	// Total stats (since date)
	sinceArgs := append([]any{since}, paidStatusList...)
	row = r.sql.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM payment_orders WHERE paid_at >= $1 AND status IN ("+statusPlaceholders+")",
		sinceArgs...)
	var totalAmountStr string
	if err := row.Scan(&totalAmountStr, &data.TotalOrderCount); err != nil {
		return nil, err
	}
	data.TotalAmount, _ = decimal.NewFromString(totalAmountStr)

	// Daily series
	rows, err := r.sql.QueryContext(ctx,
		`SELECT DATE(paid_at AT TIME ZONE 'Asia/Shanghai') as d, COALESCE(SUM(amount), 0), COUNT(*)
		FROM payment_orders
		WHERE paid_at >= $1 AND status IN (`+statusPlaceholders+`)
		GROUP BY d ORDER BY d`,
		sinceArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pt service.DailySeriesPoint
		var amtStr string
		if err := rows.Scan(&pt.Date, &amtStr, &pt.Count); err != nil {
			return nil, err
		}
		pt.Amount, _ = decimal.NewFromString(amtStr)
		data.DailySeries = append(data.DailySeries, pt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Payment methods breakdown (with success count and rate)
	rows2, err := r.sql.QueryContext(ctx,
		`SELECT payment_type, COALESCE(SUM(amount), 0), COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('completed', 'refunding', 'refunded', 'refund_failed', 'partially_refunded'))
		FROM payment_orders
		WHERE paid_at >= $1 AND status IN (`+statusPlaceholders+`)
		GROUP BY payment_type`,
		sinceArgs...)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var pm service.PaymentMethodStat
		var amtStr string
		if err := rows2.Scan(&pm.PaymentType, &amtStr, &pm.Count, &pm.SuccessCount); err != nil {
			return nil, err
		}
		pm.Amount, _ = decimal.NewFromString(amtStr)
		if pm.Count > 0 {
			pm.SuccessRate = float64(pm.SuccessCount) / float64(pm.Count)
		}
		data.PaymentMethods = append(data.PaymentMethods, pm)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	// Leaderboard: top 10 users by amount
	rows3, err := r.sql.QueryContext(ctx,
		`SELECT po.user_id, po.user_email, COALESCE(SUM(po.amount), 0), COUNT(*)
		FROM payment_orders po
		WHERE po.paid_at >= $1 AND po.status IN (`+statusPlaceholders+`)
		GROUP BY po.user_id, po.user_email
		ORDER BY SUM(po.amount) DESC
		LIMIT 10`,
		sinceArgs...)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var entry service.LeaderboardEntry
		var amtStr string
		if err := rows3.Scan(&entry.UserID, &entry.UserEmail, &amtStr, &entry.Count); err != nil {
			return nil, err
		}
		entry.Amount, _ = decimal.NewFromString(amtStr)
		data.Leaderboard = append(data.Leaderboard, entry)
	}
	if err := rows3.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

// --- helpers ---

func (r *paymentOrderRepository) sumAmount(ctx context.Context, query string, args ...any) (decimal.Decimal, error) {
	row := r.sql.QueryRowContext(ctx, query, args...)
	var s string
	if err := row.Scan(&s); err != nil {
		return decimal.Zero, err
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, nil
	}
	return d, nil
}

func paymentOrderToService(m *dbent.PaymentOrder) *service.PaymentOrder {
	if m == nil {
		return nil
	}
	return &service.PaymentOrder{
		ID:                  m.ID,
		UserID:              m.UserID,
		UserEmail:           m.UserEmail,
		UserName:            m.UserName,
		UserNotes:           m.UserNotes,
		Amount:              m.Amount,
		PayAmount:           m.PayAmount,
		FeeRate:             m.FeeRate,
		RechargeCode:        m.RechargeCode,
		Status:              m.Status,
		PaymentType:         m.PaymentType,
		PaymentTradeNo:      m.PaymentTradeNo,
		PayURL:              m.PayURL,
		QrCode:              m.QrCode,
		QrCodeImg:           m.QrCodeImg,
		RefundAmount:        m.RefundAmount,
		RefundReason:        m.RefundReason,
		RefundAt:            m.RefundAt,
		ForceRefund:         m.ForceRefund,
		RefundRequestedAt:   m.RefundRequestedAt,
		RefundRequestReason: m.RefundRequestReason,
		RefundRequestedBy:   m.RefundRequestedBy,
		ExpiresAt:           m.ExpiresAt,
		PaidAt:              m.PaidAt,
		CompletedAt:         m.CompletedAt,
		FailedAt:            m.FailedAt,
		FailedReason:        m.FailedReason,
		ClientIP:            m.ClientIP,
		SrcHost:             m.SrcHost,
		SrcURL:              m.SrcURL,
		OrderType:           m.OrderType,
		PlanID:              m.PlanID,
		SubscriptionGroupID: m.SubscriptionGroupID,
		SubscriptionDays:    m.SubscriptionDays,
		ProviderInstanceID:  m.ProviderInstanceID,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}

func applyPaymentOrderEntity(dst *service.PaymentOrder, src *dbent.PaymentOrder) {
	dst.ID = src.ID
	dst.CreatedAt = src.CreatedAt
	dst.UpdatedAt = src.UpdatedAt
}

func paymentOrdersToService(ms []*dbent.PaymentOrder) []service.PaymentOrder {
	result := make([]service.PaymentOrder, len(ms))
	for i, m := range ms {
		result[i] = *paymentOrderToService(m)
	}
	return result
}
