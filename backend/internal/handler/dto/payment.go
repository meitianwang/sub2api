package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

// --- Payment Order DTOs ---

// UserPaymentOrderDTO is the user-facing order DTO.
// It excludes internal fields (recharge_code, provider_instance_id, payment_trade_no, user_notes).
type UserPaymentOrderDTO struct {
	ID                  int64   `json:"id"`
	Amount              string  `json:"amount"`
	PayAmount           *string `json:"pay_amount,omitempty"`
	FeeRate             *string `json:"fee_rate,omitempty"`
	Status              string  `json:"status"`
	PaymentType         string  `json:"payment_type"`
	OrderType           string  `json:"order_type"`
	PlanID              *int64  `json:"plan_id,omitempty"`
	SubscriptionGroupID *int64  `json:"subscription_group_id,omitempty"`
	SubscriptionDays    *int    `json:"subscription_days,omitempty"`
	RefundAmount        *string `json:"refund_amount,omitempty"`
	RefundReason        *string `json:"refund_reason,omitempty"`
	RefundAt            *string `json:"refund_at,omitempty"`
	RefundRequestedAt   *string `json:"refund_requested_at,omitempty"`
	RefundRequestReason *string `json:"refund_request_reason,omitempty"`
	FailedReason        *string `json:"failed_reason,omitempty"`
	ExpiresAt           string  `json:"expires_at"`
	PaidAt              *string `json:"paid_at,omitempty"`
	CompletedAt         *string `json:"completed_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
}

// PaymentOrderDTO is the admin-facing order DTO with all fields.
type PaymentOrderDTO struct {
	ID                  int64    `json:"id"`
	UserID              int64    `json:"user_id"`
	UserEmail           *string  `json:"user_email,omitempty"`
	UserName            *string  `json:"user_name,omitempty"`
	Amount              string   `json:"amount"`
	PayAmount           *string  `json:"pay_amount,omitempty"`
	FeeRate             *string  `json:"fee_rate,omitempty"`
	RechargeCode        string   `json:"recharge_code"`
	Status              string   `json:"status"`
	PaymentType         string   `json:"payment_type"`
	PaymentTradeNo      *string  `json:"payment_trade_no,omitempty"`
	OrderType           string   `json:"order_type"`
	PlanID              *int64   `json:"plan_id,omitempty"`
	SubscriptionGroupID *int64   `json:"subscription_group_id,omitempty"`
	SubscriptionDays    *int     `json:"subscription_days,omitempty"`
	ProviderInstanceID  *int64   `json:"provider_instance_id,omitempty"`
	RefundAmount        *string  `json:"refund_amount,omitempty"`
	RefundReason        *string  `json:"refund_reason,omitempty"`
	RefundAt            *string  `json:"refund_at,omitempty"`
	RefundRequestedAt   *string  `json:"refund_requested_at,omitempty"`
	RefundRequestReason *string  `json:"refund_request_reason,omitempty"`
	FailedReason        *string  `json:"failed_reason,omitempty"`
	SrcHost             *string  `json:"src_host,omitempty"`
	ExpiresAt           string   `json:"expires_at"`
	PaidAt              *string  `json:"paid_at,omitempty"`
	CompletedAt         *string  `json:"completed_at,omitempty"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type CreateOrderResponse struct {
	OrderID      int64  `json:"order_id"`
	Amount       string `json:"amount"`
	PayAmount    string `json:"pay_amount"`
	FeeRate      string `json:"fee_rate"`
	Status       string `json:"status"`
	PaymentType  string `json:"payment_type"`
	OrderType    string `json:"order_type"`
	PayURL       string `json:"pay_url,omitempty"`
	QrCode       string `json:"qr_code,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	ExpiresAt    string `json:"expires_at"`
	AccessToken  string `json:"access_token,omitempty"`
}

type PaymentAuditLogDTO struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	Action    string  `json:"action"`
	Detail    *string `json:"detail,omitempty"`
	Operator  *string `json:"operator,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type AdminOrderDetailDTO struct {
	Order     PaymentOrderDTO       `json:"order"`
	AuditLogs []PaymentAuditLogDTO  `json:"audit_logs"`
}

// --- Payment Config DTOs ---

type PaymentConfigDTO struct {
	EnabledPaymentTypes    []string          `json:"enabled_payment_types"`
	MinRechargeAmount      string            `json:"min_recharge_amount"`
	MaxRechargeAmount      string            `json:"max_recharge_amount"`
	MaxDailyRechargeAmount string            `json:"max_daily_recharge_amount"`
	BalancePaymentDisabled bool              `json:"balance_payment_disabled"`
	MaxPendingOrders       int               `json:"max_pending_orders"`
	PendingCount           int               `json:"pending_count"`
	MethodLimits           []MethodLimitDTO  `json:"method_limits"`
}

type MethodLimitDTO struct {
	PaymentType string `json:"payment_type"`
	Available   bool   `json:"available"`
	DailyLimit  string `json:"daily_limit"`
	DailyUsed   string `json:"daily_used"`
	Remaining   string `json:"remaining"`
	SingleMin   string `json:"single_min"`
	SingleMax   string `json:"single_max"`
	FeeRate     string `json:"fee_rate"`
}

// --- Channel / Plan / Instance DTOs ---

type PaymentChannelDTO struct {
	ID             int64   `json:"id"`
	GroupID        *int64  `json:"group_id,omitempty"`
	Name           string  `json:"name"`
	Platform       string  `json:"platform"`
	RateMultiplier string  `json:"rate_multiplier"`
	Description    *string `json:"description,omitempty"`
	Models         *string `json:"models,omitempty"`
	Features       *string `json:"features,omitempty"`
	SortOrder      int     `json:"sort_order"`
	Enabled        bool    `json:"enabled"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type SubscriptionPlanDTO struct {
	ID            int64   `json:"id"`
	GroupID       *int64  `json:"group_id,omitempty"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	Price         string  `json:"price"`
	OriginalPrice *string `json:"original_price,omitempty"`
	ValidityDays  int     `json:"validity_days"`
	ValidityUnit  string  `json:"validity_unit"`
	Features      *string `json:"features,omitempty"`
	ProductName   *string `json:"product_name,omitempty"`
	ForSale       bool    `json:"for_sale"`
	SortOrder     int     `json:"sort_order"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type PaymentDashboardDTO struct {
	TodayAmount     string                       `json:"today_amount"`
	TodayOrderCount int                          `json:"today_order_count"`
	TotalAmount     string                       `json:"total_amount"`
	TotalOrderCount int                          `json:"total_order_count"`
	DailySeries     []service.DailySeriesPoint   `json:"daily_series"`
	PaymentMethods  []service.PaymentMethodStat  `json:"payment_methods"`
	Leaderboard     []service.LeaderboardEntry   `json:"leaderboard,omitempty"`
}

// --- Mapper functions ---

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

func decimalPtrToStringPtr(d *decimal.Decimal) *string {
	if d == nil {
		return nil
	}
	s := d.String()
	return &s
}

// UserPaymentOrderFromService converts a service PaymentOrder to a user-facing DTO,
// excluding internal fields like recharge_code, provider_instance_id, payment_trade_no.
func UserPaymentOrderFromService(o *service.PaymentOrder) *UserPaymentOrderDTO {
	if o == nil {
		return nil
	}
	return &UserPaymentOrderDTO{
		ID:                  o.ID,
		Amount:              o.Amount.String(),
		PayAmount:           decimalPtrToStringPtr(o.PayAmount),
		FeeRate:             decimalPtrToStringPtr(o.FeeRate),
		Status:              o.Status,
		PaymentType:         o.PaymentType,
		OrderType:           o.OrderType,
		PlanID:              o.PlanID,
		SubscriptionGroupID: o.SubscriptionGroupID,
		SubscriptionDays:    o.SubscriptionDays,
		RefundAmount:        decimalPtrToStringPtr(o.RefundAmount),
		RefundReason:        o.RefundReason,
		RefundAt:            formatTimePtr(o.RefundAt),
		RefundRequestedAt:   formatTimePtr(o.RefundRequestedAt),
		RefundRequestReason: o.RefundRequestReason,
		FailedReason:        o.FailedReason,
		ExpiresAt:           formatTime(o.ExpiresAt),
		PaidAt:              formatTimePtr(o.PaidAt),
		CompletedAt:         formatTimePtr(o.CompletedAt),
		CreatedAt:           formatTime(o.CreatedAt),
	}
}

// PaymentOrderFromService converts a service PaymentOrder to an admin-facing DTO with all fields.
func PaymentOrderFromService(o *service.PaymentOrder) *PaymentOrderDTO {
	if o == nil {
		return nil
	}
	return &PaymentOrderDTO{
		ID:                  o.ID,
		UserID:              o.UserID,
		UserEmail:           o.UserEmail,
		UserName:            o.UserName,
		Amount:              o.Amount.String(),
		PayAmount:           decimalPtrToStringPtr(o.PayAmount),
		FeeRate:             decimalPtrToStringPtr(o.FeeRate),
		RechargeCode:        o.RechargeCode,
		Status:              o.Status,
		PaymentType:         o.PaymentType,
		PaymentTradeNo:      o.PaymentTradeNo,
		OrderType:           o.OrderType,
		PlanID:              o.PlanID,
		SubscriptionGroupID: o.SubscriptionGroupID,
		SubscriptionDays:    o.SubscriptionDays,
		ProviderInstanceID:  o.ProviderInstanceID,
		RefundAmount:        decimalPtrToStringPtr(o.RefundAmount),
		RefundReason:        o.RefundReason,
		RefundAt:            formatTimePtr(o.RefundAt),
		RefundRequestedAt:   formatTimePtr(o.RefundRequestedAt),
		RefundRequestReason: o.RefundRequestReason,
		FailedReason:        o.FailedReason,
		SrcHost:             o.SrcHost,
		ExpiresAt:           formatTime(o.ExpiresAt),
		PaidAt:              formatTimePtr(o.PaidAt),
		CompletedAt:         formatTimePtr(o.CompletedAt),
		CreatedAt:           formatTime(o.CreatedAt),
		UpdatedAt:           formatTime(o.UpdatedAt),
	}
}

func PaymentAuditLogFromService(l *service.PaymentAuditLog) *PaymentAuditLogDTO {
	if l == nil {
		return nil
	}
	return &PaymentAuditLogDTO{
		ID:        l.ID,
		OrderID:   l.OrderID,
		Action:    l.Action,
		Detail:    l.Detail,
		Operator:  l.Operator,
		CreatedAt: formatTime(l.CreatedAt),
	}
}

func PaymentChannelFromService(c *service.PaymentChannel) *PaymentChannelDTO {
	if c == nil {
		return nil
	}
	return &PaymentChannelDTO{
		ID:             c.ID,
		GroupID:        c.GroupID,
		Name:           c.Name,
		Platform:       c.Platform,
		RateMultiplier: c.RateMultiplier.String(),
		Description:    c.Description,
		Models:         c.Models,
		Features:       c.Features,
		SortOrder:      c.SortOrder,
		Enabled:        c.Enabled,
		CreatedAt:      formatTime(c.CreatedAt),
		UpdatedAt:      formatTime(c.UpdatedAt),
	}
}

func SubscriptionPlanFromService(p *service.SubscriptionPlan) *SubscriptionPlanDTO {
	if p == nil {
		return nil
	}
	return &SubscriptionPlanDTO{
		ID:            p.ID,
		GroupID:       p.GroupID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         p.Price.String(),
		OriginalPrice: decimalPtrToStringPtr(p.OriginalPrice),
		ValidityDays:  p.ValidityDays,
		ValidityUnit:  p.ValidityUnit,
		Features:      p.Features,
		ProductName:   p.ProductName,
		ForSale:       p.ForSale,
		SortOrder:     p.SortOrder,
		CreatedAt:     formatTime(p.CreatedAt),
		UpdatedAt:     formatTime(p.UpdatedAt),
	}
}
