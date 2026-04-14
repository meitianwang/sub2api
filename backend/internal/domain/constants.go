package domain

// Status constants
const (
	StatusActive   = "active"
	StatusDisabled = "disabled"
	StatusError    = "error"
	StatusUnused   = "unused"
	StatusUsed     = "used"
	StatusExpired  = "expired"
)

// Role constants
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Account type constants
const (
	AccountTypeOAuth      = "oauth"       // OAuth类型账号（full scope: profile + inference）
	AccountTypeSetupToken = "setup-token" // Setup Token类型账号（inference only scope）
	AccountTypeAPIKey     = "apikey"      // API Key类型账号
	AccountTypeUpstream = "upstream" // 上游透传类型账号（通过 Base URL + API Key 连接上游）
)

// Redeem type constants
const (
	RedeemTypeBalance      = "balance"
	RedeemTypeConcurrency  = "concurrency"
	RedeemTypeSubscription = "subscription"
	RedeemTypeInvitation   = "invitation"
)

// PromoCode status constants
const (
	PromoCodeStatusActive   = "active"
	PromoCodeStatusDisabled = "disabled"
)

// Admin adjustment type constants
const (
	AdjustmentTypeAdminBalance     = "admin_balance"     // 管理员调整余额
	AdjustmentTypeAdminConcurrency = "admin_concurrency" // 管理员调整并发数
)

// Group subscription type constants
const (
	SubscriptionTypeStandard     = "standard"     // 标准计费模式（按余额扣费）
	SubscriptionTypeSubscription = "subscription" // 订阅模式（按限额控制）
)

// Subscription status constants
const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusExpired   = "expired"
	SubscriptionStatusSuspended = "suspended"
)

// Payment order status constants
const (
	PaymentOrderStatusPending           = "pending"
	PaymentOrderStatusPaid              = "paid"
	PaymentOrderStatusRecharging        = "recharging"
	PaymentOrderStatusCompleted         = "completed"
	PaymentOrderStatusExpired           = "expired"
	PaymentOrderStatusCancelled         = "cancelled"
	PaymentOrderStatusFailed            = "failed"
	PaymentOrderStatusRefundRequested   = "refund_requested"
	PaymentOrderStatusRefunding         = "refunding"
	PaymentOrderStatusPartiallyRefunded = "partially_refunded"
	PaymentOrderStatusRefunded          = "refunded"
	PaymentOrderStatusRefundFailed      = "refund_failed"
)

// Payment order type constants
const (
	PaymentOrderTypeBalance      = "balance"
	PaymentOrderTypeSubscription = "subscription"
)

// Payment type constants
const (
	PaymentTypeAlipay       = "alipay"
	PaymentTypeAlipayDirect = "alipay_direct"
	PaymentTypeWxpay        = "wxpay"
	PaymentTypeWxpayDirect  = "wxpay_direct"
	PaymentTypeStripe       = "stripe"
)

// Payment provider key constants
const (
	PaymentProviderEasyPay = "easypay"
	PaymentProviderAlipay  = "alipay"
	PaymentProviderWxpay   = "wxpay"
	PaymentProviderStripe  = "stripe"
)

