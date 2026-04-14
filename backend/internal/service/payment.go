package service

import (
	"time"

	"github.com/shopspring/decimal"
)

// PaymentOrder is the service-layer domain model for a payment order.
type PaymentOrder struct {
	ID                  int64
	UserID              int64
	UserEmail           *string
	UserName            *string
	UserNotes           *string
	Amount              decimal.Decimal
	PayAmount           *decimal.Decimal
	FeeRate             *decimal.Decimal
	RechargeCode        string
	Status              string
	PaymentType         string
	PaymentTradeNo      *string
	PayURL              *string
	QrCode              *string
	QrCodeImg           *string
	RefundAmount        *decimal.Decimal
	RefundReason        *string
	RefundAt            *time.Time
	ForceRefund         bool
	RefundRequestedAt   *time.Time
	RefundRequestReason *string
	RefundRequestedBy   *int64
	ExpiresAt           time.Time
	PaidAt              *time.Time
	CompletedAt         *time.Time
	FailedAt            *time.Time
	FailedReason        *string
	ClientIP            *string
	SrcHost             *string
	SrcURL              *string
	OrderType           string
	PlanID              *int64
	SubscriptionGroupID *int64
	SubscriptionDays    *int
	ProviderInstanceID  *int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// PaymentAuditLog is the service-layer domain model for payment audit logs.
type PaymentAuditLog struct {
	ID        int64
	OrderID   int64
	Action    string
	Detail    *string
	Operator  *string
	CreatedAt time.Time
}

// PaymentChannel is the service-layer domain model for a payment channel display config.
type PaymentChannel struct {
	ID             int64
	GroupID        *int64
	Name           string
	Platform       string
	RateMultiplier decimal.Decimal
	Description    *string
	Models         *string
	Features       *string
	SortOrder      int
	Enabled        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// PaymentProviderInstance is the service-layer domain model for a provider instance.
type PaymentProviderInstance struct {
	ID             int64
	ProviderKey    string
	Name           string
	Config         string // AES-256-GCM encrypted JSON
	SupportedTypes string // comma-separated PaymentType values
	Enabled        bool
	SortOrder      int
	Limits         *string // JSON: per-type daily/single limits
	RefundEnabled  bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SubscriptionPlan is the service-layer domain model for a subscription plan.
type SubscriptionPlan struct {
	ID            int64
	GroupID       *int64
	Name          string
	Description   *string
	Price         decimal.Decimal
	OriginalPrice *decimal.Decimal
	ValidityDays  int
	ValidityUnit  string
	Features      *string
	ProductName   *string
	ForSale       bool
	SortOrder     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProviderInstanceLimits represents parsed per-type limits for a provider instance.
type ProviderInstanceLimits struct {
	DailyLimit *decimal.Decimal `json:"daily_limit,omitempty"`
	SingleMin  *decimal.Decimal `json:"single_min,omitempty"`
	SingleMax  *decimal.Decimal `json:"single_max,omitempty"`
}
