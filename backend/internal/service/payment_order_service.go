package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

// --- Sentinel errors ---

var (
	ErrPaymentOrderNotFound     = infraerrors.NotFound("PAYMENT_ORDER_NOT_FOUND", "payment order not found")
	ErrPaymentChannelNotFound   = infraerrors.NotFound("PAYMENT_CHANNEL_NOT_FOUND", "payment channel not found")
	ErrPaymentProviderNotFound  = infraerrors.NotFound("PAYMENT_PROVIDER_NOT_FOUND", "payment provider instance not found")
	ErrSubscriptionPlanNotFound = infraerrors.NotFound("SUBSCRIPTION_PLAN_NOT_FOUND", "subscription plan not found")
	ErrPaymentTypeDisabled      = infraerrors.BadRequest("PAYMENT_TYPE_DISABLED", "payment type is not enabled")
	ErrPaymentAmountInvalid     = infraerrors.BadRequest("PAYMENT_AMOUNT_INVALID", "payment amount is invalid")
	ErrTooManyPendingOrders     = infraerrors.TooManyRequests("TOO_MANY_PENDING_ORDERS", "too many pending orders")
	ErrDailyRechargeExceeded    = infraerrors.TooManyRequests("DAILY_RECHARGE_EXCEEDED", "daily recharge limit exceeded")
	ErrOrderAlreadyPaid         = infraerrors.Conflict("ORDER_ALREADY_PAID", "order is already paid")
	ErrOrderNotPending          = infraerrors.Conflict("ORDER_NOT_PENDING", "order is not in pending status")
	ErrProviderInstanceUnavail  = infraerrors.ServiceUnavailable("PROVIDER_UNAVAILABLE", "no available provider instance")
	ErrNoProviderForType        = infraerrors.BadRequest("NO_PROVIDER_FOR_TYPE", "no provider configured for this payment type")
	ErrGlobalDailyLimitExceeded = infraerrors.TooManyRequests("GLOBAL_DAILY_LIMIT", "global daily limit exceeded for this method")
	ErrSingleAmountLimitExceed  = infraerrors.BadRequest("SINGLE_AMOUNT_LIMIT", "amount exceeds single transaction limit")
	ErrCancelRateLimited        = infraerrors.TooManyRequests("CANCEL_RATE_LIMITED", "too many order cancellations, please try again later")
)

// --- Repository interfaces ---

// PaymentOrderRepository defines the data access interface for payment orders.
type PaymentOrderRepository interface {
	Create(ctx context.Context, order *PaymentOrder) error
	GetByID(ctx context.Context, id int64) (*PaymentOrder, error)
	GetByRechargeCode(ctx context.Context, code string) (*PaymentOrder, error)
	Update(ctx context.Context, order *PaymentOrder) error
	Delete(ctx context.Context, id int64) error
	// UpdateStatusCAS atomically updates status from fromStatus to toStatus. Returns true if updated.
	UpdateStatusCAS(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error)
	// ConfirmPaidCAS atomically transitions an order from PENDING (or recently-EXPIRED within grace)
	// to PAID, setting trade number and paid time. Returns true if the update was applied.
	ConfirmPaidCAS(ctx context.Context, id int64, tradeNo string, paidAmount decimal.Decimal, paidAt time.Time, graceDeadline time.Time) (bool, error)
	UpdateRechargeCode(ctx context.Context, id int64, code string) error
	UpdatePaymentResult(ctx context.Context, id int64, tradeNo string, payURL, qrCode *string, instanceID *int64) error

	CountPendingByUserID(ctx context.Context, userID int64) (int, error)
	CountActiveByProviderInstanceID(ctx context.Context, instanceID int64) (int, error)
	CountActiveByPlanID(ctx context.Context, planID int64) (int, error)
	SumDailyPaidByUserID(ctx context.Context, userID int64, bizDayStart time.Time) (decimal.Decimal, error)
	SumDailyPaidByPaymentType(ctx context.Context, paymentType string, bizDayStart time.Time) (decimal.Decimal, error)
	SumDailyPaidByInstanceID(ctx context.Context, instanceID int64, bizDayStart time.Time) (decimal.Decimal, error)
	SumDailyPaidByInstanceIDs(ctx context.Context, instanceIDs []int64, bizDayStart time.Time) (map[int64]decimal.Decimal, error)
	ListExpiredPending(ctx context.Context, now time.Time, limit int) ([]PaymentOrder, error)

	List(ctx context.Context, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error)
	ListByUserID(ctx context.Context, userID int64, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error)
	ListFiltered(ctx context.Context, filter PaymentOrderListFilter, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error)
	GetDashboardStats(ctx context.Context, since time.Time) (*RawDashboardData, error)
}

// PaymentAuditLogRepository defines the data access interface for payment audit logs.
type PaymentAuditLogRepository interface {
	Create(ctx context.Context, log *PaymentAuditLog) error
	ListByOrderID(ctx context.Context, orderID int64) ([]PaymentAuditLog, error)
	CountByActionAndOperatorSince(ctx context.Context, action, operator string, since time.Time) (int, error)
}

// PaymentChannelRepository defines the data access interface for payment channels.
type PaymentChannelRepository interface {
	Create(ctx context.Context, channel *PaymentChannel) error
	GetByID(ctx context.Context, id int64) (*PaymentChannel, error)
	GetByGroupID(ctx context.Context, groupID int64) (*PaymentChannel, error)
	Update(ctx context.Context, channel *PaymentChannel) error
	Delete(ctx context.Context, id int64) error
	ListEnabled(ctx context.Context) ([]PaymentChannel, error)
	List(ctx context.Context) ([]PaymentChannel, error)
}

// PaymentProviderInstanceRepository defines the data access interface for provider instances.
type PaymentProviderInstanceRepository interface {
	Create(ctx context.Context, instance *PaymentProviderInstance) error
	GetByID(ctx context.Context, id int64) (*PaymentProviderInstance, error)
	Update(ctx context.Context, instance *PaymentProviderInstance) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]PaymentProviderInstance, error)
	ListEnabled(ctx context.Context) ([]PaymentProviderInstance, error)
	ListEnabledByProviderKey(ctx context.Context, providerKey string) ([]PaymentProviderInstance, error)
}

// SubscriptionPlanRepository defines the data access interface for subscription plans.
type SubscriptionPlanRepository interface {
	Create(ctx context.Context, plan *SubscriptionPlan) error
	GetByID(ctx context.Context, id int64) (*SubscriptionPlan, error)
	Update(ctx context.Context, plan *SubscriptionPlan) error
	Delete(ctx context.Context, id int64) error
	ListForSale(ctx context.Context) ([]SubscriptionPlan, error)
	List(ctx context.Context) ([]SubscriptionPlan, error)
}

// --- PaymentConfigService ---

const paymentConfigCacheTTL = 30 * time.Second

type cachedPaymentConfig struct {
	values    map[string]string
	expiresAt int64 // unix nano
}

// PaymentConfigService provides cached access to payment-related settings.
type PaymentConfigService struct {
	settingService *SettingService
	cache          atomic.Value // *cachedPaymentConfig
	mu             sync.Mutex
	fallbackSecret string // random per-process fallback for status access tokens
}

// NewPaymentConfigService creates a new PaymentConfigService.
func NewPaymentConfigService(settingService *SettingService) *PaymentConfigService {
	// Generate a random fallback secret so deployments without an admin API key
	// don't all share the same hardcoded value.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return &PaymentConfigService{
		settingService: settingService,
		fallbackSecret: hex.EncodeToString(b),
	}
}

func (s *PaymentConfigService) getAll(ctx context.Context) map[string]string {
	if cached, ok := s.cache.Load().(*cachedPaymentConfig); ok {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.values
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// double-check
	if cached, ok := s.cache.Load().(*cachedPaymentConfig); ok {
		if time.Now().UnixNano() < cached.expiresAt {
			return cached.values
		}
	}

	// Use GetAll to load all settings so that dynamically-keyed entries
	// (e.g. pay_fee_rate_alipay, pay_fee_rate_provider_easypay) are included.
	all, err := s.settingService.settingRepo.GetAll(ctx)
	if err != nil {
		slog.Warn("failed to load payment config from DB, using empty", "error", err)
		all = make(map[string]string)
	}
	// Keep only pay-related keys to avoid caching unrelated settings.
	values := make(map[string]string, len(all))
	for k, v := range all {
		if strings.HasPrefix(k, "pay_") {
			values[k] = v
		}
	}

	s.cache.Store(&cachedPaymentConfig{
		values:    values,
		expiresAt: time.Now().Add(paymentConfigCacheTTL).UnixNano(),
	})
	return values
}

func (s *PaymentConfigService) getString(ctx context.Context, key, fallback string) string {
	if v, ok := s.getAll(ctx)[key]; ok && v != "" {
		return v
	}
	return fallback
}

func (s *PaymentConfigService) getInt(ctx context.Context, key string, fallback int) int {
	if v, ok := s.getAll(ctx)[key]; ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func (s *PaymentConfigService) getDecimal(ctx context.Context, key string, fallback decimal.Decimal) decimal.Decimal {
	if v, ok := s.getAll(ctx)[key]; ok && v != "" {
		if d, err := decimal.NewFromString(v); err == nil {
			return d
		}
	}
	return fallback
}

func (s *PaymentConfigService) GetOrderTimeoutMinutes(ctx context.Context) int {
	return s.getInt(ctx, SettingKeyPayOrderTimeoutMinutes, 30)
}

func (s *PaymentConfigService) GetMinRechargeAmount(ctx context.Context) decimal.Decimal {
	return s.getDecimal(ctx, SettingKeyPayMinRechargeAmount, decimal.NewFromInt(1))
}

func (s *PaymentConfigService) GetMaxRechargeAmount(ctx context.Context) decimal.Decimal {
	return s.getDecimal(ctx, SettingKeyPayMaxRechargeAmount, decimal.NewFromInt(10000))
}

func (s *PaymentConfigService) GetMaxDailyRechargeAmount(ctx context.Context) decimal.Decimal {
	return s.getDecimal(ctx, SettingKeyPayMaxDailyRechargeAmount, decimal.NewFromInt(100000))
}

func (s *PaymentConfigService) GetProductName(ctx context.Context) string {
	return s.getString(ctx, SettingKeyPayProductName, "Sub2API Balance Recharge")
}

func (s *PaymentConfigService) GetLoadBalanceStrategy(ctx context.Context) string {
	return s.getString(ctx, "pay_load_balance_strategy", "round_robin")
}

func (s *PaymentConfigService) GetGlobalDailyLimit(ctx context.Context, paymentType string) decimal.Decimal {
	var key string
	switch {
	case strings.Contains(paymentType, "alipay"):
		key = SettingKeyPayMaxDailyAmountAlipay
	case strings.Contains(paymentType, "wxpay"):
		key = SettingKeyPayMaxDailyAmountWxpay
	case paymentType == domain.PaymentTypeStripe:
		key = SettingKeyPayMaxDailyAmountStripe
	}
	if key == "" {
		return decimal.Zero // 0 means unlimited
	}
	return s.getDecimal(ctx, key, decimal.Zero)
}

func (s *PaymentConfigService) GetFeeRate(ctx context.Context, paymentType, providerKey string) decimal.Decimal {
	// Priority: FEE_RATE_{TYPE} > FEE_RATE_PROVIDER_{KEY} > 0
	typeKey := "pay_fee_rate_" + paymentType
	if d := s.getDecimal(ctx, typeKey, decimal.NewFromInt(-1)); !d.Equal(decimal.NewFromInt(-1)) {
		return d
	}
	providerKeyConfig := "pay_fee_rate_provider_" + providerKey
	if d := s.getDecimal(ctx, providerKeyConfig, decimal.NewFromInt(-1)); !d.Equal(decimal.NewFromInt(-1)) {
		return d
	}
	return decimal.Zero
}

// GetStatusAccessSecret returns a secret used to sign order status access tokens.
// Uses the admin API key from settings; falls back to a static default if not configured.
func (s *PaymentConfigService) GetStatusAccessSecret(ctx context.Context) string {
	key, err := s.settingService.GetAdminAPIKey(ctx)
	if err == nil && key != "" {
		return key
	}
	return s.fallbackSecret
}

func (s *PaymentConfigService) IsAutoRefundEnabled(ctx context.Context) bool {
	return s.getString(ctx, "pay_auto_refund_enabled", "false") == "true"
}

// GetGracePeriodMinutes returns the number of minutes after expiration during which
// a payment notification is still accepted. Default: 5 minutes.
func (s *PaymentConfigService) GetGracePeriodMinutes(ctx context.Context) int {
	return s.getInt(ctx, SettingKeyPayGracePeriodMinutes, 5)
}

func (s *PaymentConfigService) IsBalancePaymentDisabled(ctx context.Context) bool {
	return s.getString(ctx, "pay_balance_payment_disabled", "false") == "true"
}

func (s *PaymentConfigService) GetProductNamePrefix(ctx context.Context) string {
	return s.getString(ctx, "pay_product_name_prefix", "")
}

func (s *PaymentConfigService) GetProductNameSuffix(ctx context.Context) string {
	return s.getString(ctx, "pay_product_name_suffix", "")
}

func (s *PaymentConfigService) IsCancelRateLimitEnabled(ctx context.Context) bool {
	return s.getString(ctx, SettingKeyPayCancelRateLimitEnabled, "false") == "true"
}

func (s *PaymentConfigService) GetCancelRateLimitWindow(ctx context.Context) int {
	return s.getInt(ctx, SettingKeyPayCancelRateLimitWindow, 1)
}

func (s *PaymentConfigService) GetCancelRateLimitUnit(ctx context.Context) string {
	unit := s.getString(ctx, SettingKeyPayCancelRateLimitUnit, "day")
	switch unit {
	case "minute", "hour", "day":
		return unit
	default:
		return "day"
	}
}

func (s *PaymentConfigService) GetCancelRateLimitMax(ctx context.Context) int {
	return s.getInt(ctx, SettingKeyPayCancelRateLimitMax, 10)
}

func (s *PaymentConfigService) GetCancelRateLimitWindowMode(ctx context.Context) string {
	mode := s.getString(ctx, SettingKeyPayCancelRateLimitWindowMode, "rolling")
	if mode == "fixed" {
		return "fixed"
	}
	return "rolling"
}

func (s *PaymentConfigService) GetMaxPendingOrders(ctx context.Context) int {
	return s.getInt(ctx, "pay_max_pending_orders", 3)
}

// GetEnabledPaymentTypes returns the configured payment types filter.
// If empty, all registered types are enabled.
func (s *PaymentConfigService) GetEnabledPaymentTypes(ctx context.Context) []string {
	raw := s.getString(ctx, "pay_enabled_payment_types", "")
	if raw == "" {
		return nil // nil means all enabled
	}
	var result []string
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

// IsPaymentTypeEnabled checks whether a given payment type is enabled.
// Returns true if no ENABLED_PAYMENT_TYPES filter is configured (all enabled).
func (s *PaymentConfigService) IsPaymentTypeEnabled(ctx context.Context, paymentType string) bool {
	enabled := s.GetEnabledPaymentTypes(ctx)
	if len(enabled) == 0 {
		return true // no filter = all enabled
	}
	for _, t := range enabled {
		if t == paymentType {
			return true
		}
	}
	return false
}

// InvalidateCache forces a cache refresh on next access.
func (s *PaymentConfigService) InvalidateCache() {
	s.cache.Store((*cachedPaymentConfig)(nil))
}

// UpdateSettings writes payment config keys and invalidates the cache.
// Only keys with the "pay_" prefix are accepted.
func (s *PaymentConfigService) UpdateSettings(ctx context.Context, configs map[string]string) error {
	filtered := make(map[string]string, len(configs))
	for k, v := range configs {
		if !strings.HasPrefix(k, "pay_") {
			continue
		}
		filtered[k] = v
	}
	if len(filtered) == 0 {
		return nil
	}
	if err := s.settingService.settingRepo.SetMultiple(ctx, filtered); err != nil {
		return fmt.Errorf("update payment settings: %w", err)
	}
	s.InvalidateCache()
	return nil
}

// GetAllSettings returns all current pay_* settings as a map.
func (s *PaymentConfigService) GetAllSettings(ctx context.Context) map[string]string {
	return s.getAll(ctx)
}

// --- Request / Response types ---

// CreateOrderRequest bundles all input for PaymentOrderService.CreateOrder.
type CreateOrderRequest struct {
	UserID      int64
	Amount      decimal.Decimal
	PaymentType string
	OrderType   string // "balance" or "subscription"
	PlanID      *int64
	ClientIP    string
	SrcHost     string
	SrcURL      string
	ReturnURL   string
	IsMobile    bool
}

// CreateOrderResult holds the result of a successful order creation.
type CreateOrderResult struct {
	Order        *PaymentOrder
	PayURL       string
	QrCode       string
	ClientSecret string
}

// RefundOrderRequest bundles all input for RefundOrder.
type RefundOrderRequest struct {
	OrderID  int64
	Amount   decimal.Decimal
	Reason   string
	Operator string
}

// PaymentOrderListFilter holds filter criteria for admin order listing.
type PaymentOrderListFilter struct {
	UserID      *int64
	Status      *string
	OrderType   *string
	PaymentType *string
	DateFrom    *time.Time
	DateTo      *time.Time
}

// RawDashboardData holds raw aggregation data from the repository.
type RawDashboardData struct {
	TodayAmount     decimal.Decimal
	TodayOrderCount int
	TotalAmount     decimal.Decimal
	TotalOrderCount int
	DailySeries     []DailySeriesPoint
	PaymentMethods  []PaymentMethodStat
	Leaderboard     []LeaderboardEntry
}

// DailySeriesPoint holds daily aggregated amounts and counts.
type DailySeriesPoint struct {
	Date   string          `json:"date"`
	Amount decimal.Decimal `json:"amount"`
	Count  int             `json:"count"`
}

// PaymentMethodStat holds per-payment-type aggregated data.
type PaymentMethodStat struct {
	PaymentType  string          `json:"payment_type"`
	Amount       decimal.Decimal `json:"amount"`
	Count        int             `json:"count"`
	SuccessCount int             `json:"success_count"`
	SuccessRate  float64         `json:"success_rate"`
}

// LeaderboardEntry holds per-user aggregated data for the dashboard leaderboard.
type LeaderboardEntry struct {
	UserID    int64           `json:"user_id"`
	UserEmail *string         `json:"user_email,omitempty"`
	Amount    decimal.Decimal `json:"amount"`
	Count     int             `json:"count"`
}

// --- PaymentOrderService ---

const defaultMaxPendingOrders = 3

// PaymentOrderService handles the core payment order workflow.
type PaymentOrderService struct {
	orderRepo           PaymentOrderRepository
	auditLogRepo        PaymentAuditLogRepository
	instanceRepo        PaymentProviderInstanceRepository
	planRepo            SubscriptionPlanRepository
	groupRepo           GroupRepository
	userRepo            UserRepository
	userService         *UserService
	configService       *PaymentConfigService
	registry            *PaymentProviderRegistry
	loadBalancer        *PaymentLoadBalancer
	encryptor           SecretEncryptor
	entClient           *dbent.Client
	redeemService       *RedeemService
	subscriptionService *SubscriptionService
}

// NewPaymentOrderService creates a new PaymentOrderService.
func NewPaymentOrderService(
	orderRepo PaymentOrderRepository,
	auditLogRepo PaymentAuditLogRepository,
	instanceRepo PaymentProviderInstanceRepository,
	planRepo SubscriptionPlanRepository,
	groupRepo GroupRepository,
	userRepo UserRepository,
	userService *UserService,
	configService *PaymentConfigService,
	registry *PaymentProviderRegistry,
	loadBalancer *PaymentLoadBalancer,
	encryptor SecretEncryptor,
	entClient *dbent.Client,
	redeemService *RedeemService,
	subscriptionService *SubscriptionService,
) *PaymentOrderService {
	return &PaymentOrderService{
		orderRepo:           orderRepo,
		auditLogRepo:        auditLogRepo,
		instanceRepo:        instanceRepo,
		planRepo:            planRepo,
		groupRepo:           groupRepo,
		userRepo:            userRepo,
		userService:         userService,
		configService:       configService,
		registry:            registry,
		loadBalancer:        loadBalancer,
		encryptor:           encryptor,
		entClient:           entClient,
		redeemService:       redeemService,
		subscriptionService: subscriptionService,
	}
}

// CountPendingByUserID returns the number of pending orders for a user.
func (s *PaymentOrderService) CountPendingByUserID(ctx context.Context, userID int64) (int, error) {
	return s.orderRepo.CountPendingByUserID(ctx, userID)
}

// CreateOrder creates a new payment order following the full workflow.
func (s *PaymentOrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*CreateOrderResult, error) {
	// 1. Validate payment type → provider key
	providerKey := PaymentTypeToProviderKey(req.PaymentType)
	if providerKey == "" {
		return nil, ErrNoProviderForType
	}
	if !s.registry.HasProvider(providerKey) {
		return nil, ErrNoProviderForType
	}
	// Check ENABLED_PAYMENT_TYPES filter
	if !s.configService.IsPaymentTypeEnabled(ctx, req.PaymentType) {
		return nil, ErrPaymentTypeDisabled
	}

	// 2. Check balance payment disabled
	if req.OrderType == domain.PaymentOrderTypeBalance && s.configService.IsBalancePaymentDisabled(ctx) {
		return nil, infraerrors.Forbidden("BALANCE_PAYMENT_DISABLED", "balance recharge has been disabled")
	}

	// 2a. Validate subscription plan if orderType = subscription
	var plan *SubscriptionPlan
	if req.OrderType == domain.PaymentOrderTypeSubscription {
		if req.PlanID == nil {
			return nil, infraerrors.BadRequest("PLAN_REQUIRED", "plan_id is required for subscription orders")
		}
		var err error
		plan, err = s.planRepo.GetByID(ctx, *req.PlanID)
		if err != nil {
			return nil, ErrSubscriptionPlanNotFound
		}
		if !plan.ForSale {
			return nil, infraerrors.BadRequest("PLAN_NOT_FOR_SALE", "this plan is not available for purchase")
		}
		// Validate subscription group: plan must be bound to a group, group must exist and be active
		if plan.GroupID == nil {
			return nil, infraerrors.BadRequest("PLAN_NO_GROUP", "subscription plan is not bound to a group")
		}
		group, err := s.groupRepo.GetByID(ctx, *plan.GroupID)
		if err != nil {
			return nil, infraerrors.BadRequest("SUBSCRIPTION_GROUP_GONE", "subscription group no longer exists")
		}
		if group.Status != domain.StatusActive {
			return nil, infraerrors.BadRequest("SUBSCRIPTION_GROUP_INACTIVE", "subscription group is not active")
		}
		if group.SubscriptionType != domain.SubscriptionTypeSubscription {
			return nil, infraerrors.BadRequest("GROUP_NOT_SUBSCRIPTION", "group is not configured for subscriptions")
		}
	}

	// 4. Get user and verify active
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Status != StatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is not active")
	}

	// 4a. Check cancel rate limiting
	if err := s.checkCancelRateLimit(ctx, req.UserID); err != nil {
		return nil, err
	}

	// 3. Determine order amount: for subscriptions use plan price, for balance use request amount
	orderAmount := req.Amount
	if plan != nil {
		orderAmount = plan.Price
	}

	// 3a. Validate amount
	minAmount := s.configService.GetMinRechargeAmount(ctx)
	maxAmount := s.configService.GetMaxRechargeAmount(ctx)
	if orderAmount.LessThan(minAmount) || orderAmount.GreaterThan(maxAmount) {
		return nil, ErrPaymentAmountInvalid
	}

	// 5. Calculate fee
	feeRate := s.configService.GetFeeRate(ctx, req.PaymentType, providerKey)
	_, payAmount := CalculateFee(orderAmount, feeRate)

	// 6. Begin transaction for atomic checks + order creation
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	// 6a. Check global daily limit (inside transaction to prevent race condition)
	globalDailyLimit := s.configService.GetGlobalDailyLimit(ctx, req.PaymentType)
	if globalDailyLimit.GreaterThan(decimal.Zero) {
		bizDayStart := getBizDayStartUTC(time.Now().UTC())
		used, err := s.orderRepo.SumDailyPaidByPaymentType(txCtx, req.PaymentType, bizDayStart)
		if err != nil {
			return nil, fmt.Errorf("check global daily limit: %w", err)
		}
		if used.Add(orderAmount).GreaterThan(globalDailyLimit) {
			return nil, ErrGlobalDailyLimitExceeded
		}
	}

	// 7a. Check pending order count
	pendingCount, err := s.orderRepo.CountPendingByUserID(txCtx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("count pending orders: %w", err)
	}
	maxPending := s.configService.GetMaxPendingOrders(ctx)
	if pendingCount >= maxPending {
		return nil, ErrTooManyPendingOrders
	}

	// 7b. Check daily user recharge limit
	maxDailyRecharge := s.configService.GetMaxDailyRechargeAmount(ctx)
	if maxDailyRecharge.GreaterThan(decimal.Zero) {
		bizDayStart := getBizDayStartUTC(time.Now().UTC())
		dailyUsed, err := s.orderRepo.SumDailyPaidByUserID(txCtx, req.UserID, bizDayStart)
		if err != nil {
			return nil, fmt.Errorf("check daily recharge limit: %w", err)
		}
		if dailyUsed.Add(orderAmount).GreaterThan(maxDailyRecharge) {
			return nil, ErrDailyRechargeExceeded
		}
	}

	// 7c. Build order timeout
	timeoutMinutes := s.configService.GetOrderTimeoutMinutes(ctx)
	expiresAt := time.Now().Add(time.Duration(timeoutMinutes) * time.Minute)

	// 7d. Create order record (rechargeCode will be set after we have the ID)
	order := &PaymentOrder{
		UserID:      req.UserID,
		UserEmail:   stringPtr(user.Email),
		UserName:    stringPtr(user.Username),
		Amount:      orderAmount,
		PayAmount:   &payAmount,
		FeeRate:     &feeRate,
		Status:      domain.PaymentOrderStatusPending,
		PaymentType: req.PaymentType,
		ExpiresAt:   expiresAt,
		ClientIP:    stringPtrIfNotEmpty(req.ClientIP),
		SrcHost:     stringPtrIfNotEmpty(req.SrcHost),
		SrcURL:      stringPtrIfNotEmpty(req.SrcURL),
		OrderType:   req.OrderType,
	}

	if plan != nil {
		order.PlanID = req.PlanID
		order.SubscriptionGroupID = plan.GroupID
		days := computeValidityDays(plan.ValidityDays, plan.ValidityUnit)
		order.SubscriptionDays = &days
	}

	if err := s.orderRepo.Create(txCtx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// 7e. Generate recharge code using order ID (matches TypeScript original)
	rechargeCode := generateRechargeCode(order.ID)
	order.RechargeCode = rechargeCode
	if err := s.orderRepo.UpdateRechargeCode(txCtx, order.ID, rechargeCode); err != nil {
		return nil, fmt.Errorf("update recharge code: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// 8. Select provider instance via load balancer (outside tx)
	selected, err := s.loadBalancer.SelectInstance(ctx, providerKey, req.PaymentType, payAmount)
	if err != nil {
		return nil, fmt.Errorf("select provider instance: %w", err)
	}
	if selected == nil {
		return nil, ErrProviderInstanceUnavail
	}

	// 9. Create provider and call CreatePayment
	provider, err := s.registry.CreateProvider(providerKey, selected.ConfigJSON)
	if err != nil {
		return nil, fmt.Errorf("create payment provider: %w", err)
	}

	subject := s.buildPaymentSubject(ctx, order, plan)
	payResp, err := provider.CreatePayment(ctx, CreatePaymentRequest{
		OrderID:     strconv.FormatInt(order.ID, 10),
		Amount:      payAmount,
		PaymentType: req.PaymentType,
		Subject:     subject,
		NotifyURL:   "", // Will be set by provider from its config
		ReturnURL:   req.ReturnURL,
		ClientIP:    req.ClientIP,
		IsMobile:    req.IsMobile,
	})
	if err != nil {
		// Delete the order on provider failure (clean rollback, matching TypeScript behavior).
		// The order was never sent to the user, so leaving it as FAILED would clutter their order list.
		slog.Warn("payment provider CreatePayment failed, deleting order", "orderID", order.ID, "error", err)
		if delErr := s.orderRepo.Delete(ctx, order.ID); delErr != nil {
			slog.Warn("failed to delete order after provider failure", "orderID", order.ID, "error", delErr)
		}
		return nil, fmt.Errorf("create payment: %w", err)
	}

	// 10. Update order with payment result
	instanceID := &selected.InstanceID
	if err := s.orderRepo.UpdatePaymentResult(ctx, order.ID, payResp.TradeNo, stringPtrIfNotEmpty(payResp.PayURL), stringPtrIfNotEmpty(payResp.QrCode), instanceID); err != nil {
		return nil, fmt.Errorf("update payment result: %w", err)
	}
	order.PaymentTradeNo = stringPtrIfNotEmpty(payResp.TradeNo)
	order.PayURL = stringPtrIfNotEmpty(payResp.PayURL)
	order.QrCode = stringPtrIfNotEmpty(payResp.QrCode)
	order.ProviderInstanceID = instanceID

	// 11. Write audit log
	auditDetail, _ := json.Marshal(map[string]interface{}{
		"amount":      orderAmount.String(),
		"payAmount":   payAmount.String(),
		"paymentType": req.PaymentType,
		"orderType":   req.OrderType,
	})
	s.writeAuditLog(ctx, order.ID, "ORDER_CREATED", string(auditDetail))

	return &CreateOrderResult{
		Order:        order,
		PayURL:       payResp.PayURL,
		QrCode:       payResp.QrCode,
		ClientSecret: payResp.ClientSecret,
	}, nil
}

// CancelOrder cancels a pending order (user-initiated).
func (s *PaymentOrderService) CancelOrder(ctx context.Context, orderID, userID int64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}
	if order.UserID != userID {
		return ErrPaymentOrderNotFound
	}

	if order.Status != domain.PaymentOrderStatusPending {
		if order.Status == domain.PaymentOrderStatusPaid || order.Status == domain.PaymentOrderStatusCompleted {
			return ErrOrderAlreadyPaid
		}
		return ErrOrderNotPending
	}

	operator := fmt.Sprintf("user:%d", userID)
	return s.cancelOrderCore(ctx, order, domain.PaymentOrderStatusCancelled, operator, "cancelled by user")
}

// AdminCancelOrder cancels a pending order (admin-initiated).
func (s *PaymentOrderService) AdminCancelOrder(ctx context.Context, orderID int64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}

	if order.Status != domain.PaymentOrderStatusPending {
		if order.Status == domain.PaymentOrderStatusPaid || order.Status == domain.PaymentOrderStatusCompleted {
			return ErrOrderAlreadyPaid
		}
		return ErrOrderNotPending
	}

	return s.cancelOrderCore(ctx, order, domain.PaymentOrderStatusCancelled, "admin", "cancelled by admin")
}

// cancelOrderCore is the shared cancel workflow. It queries the payment provider before cancelling
// to prevent the race condition where a user pays just as the cancel request arrives.
// Returns ErrOrderAlreadyPaid if the provider reports the order as paid (and triggers fulfillment).
func (s *PaymentOrderService) cancelOrderCore(ctx context.Context, order *PaymentOrder, finalStatus, operator, auditDetail string) error {
	// Check with payment provider before cancelling — prevents money loss if payment
	// went through during the cancel window.
	if order.PaymentTradeNo != nil && *order.PaymentTradeNo != "" && order.ProviderInstanceID != nil {
		paid, notification := s.queryProviderPaymentStatus(ctx, order)
		if paid {
			slog.Info("order was paid during cancel, confirming payment", "orderID", order.ID, "operator", operator)
			if notification != nil {
				_ = s.handleNotificationForPendingOrder(ctx, order, notification)
			}
			return ErrOrderAlreadyPaid
		}
	}

	ok, err := s.orderRepo.UpdateStatusCAS(ctx, order.ID, domain.PaymentOrderStatusPending, finalStatus)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	if !ok {
		return ErrOrderNotPending
	}

	// Try to cancel at provider (best-effort)
	s.tryCancelAtProvider(ctx, order)

	s.writeAuditLogWithOperator(ctx, order.ID, "ORDER_CANCELLED", auditDetail, operator)
	return nil
}

// HandleNotification processes a payment webhook callback.
func (s *PaymentOrderService) HandleNotification(ctx context.Context, providerKey string, rawBody []byte, headers map[string]string) error {
	// Find an enabled instance to get provider config for verification
	instances, err := s.instanceRepo.ListEnabledByProviderKey(ctx, providerKey)
	if err != nil || len(instances) == 0 {
		return fmt.Errorf("no enabled instances for provider %s", providerKey)
	}

	// Try each instance until one verifies the notification.
	// Distinguish between:
	//   (nil, nil)  — signature valid but event type is irrelevant (e.g. Stripe charge.updated)
	//   (nil, err)  — signature verification failed
	//   (notif, nil) — valid payment notification
	var notification *PaymentNotification
	verifiedButIrrelevant := false
	for _, inst := range instances {
		configJSON, err := s.encryptor.Decrypt(inst.Config)
		if err != nil {
			continue
		}
		provider, err := s.registry.CreateProvider(providerKey, configJSON)
		if err != nil {
			continue
		}
		notification, err = provider.VerifyNotification(ctx, rawBody, headers)
		if err == nil {
			if notification != nil {
				break // valid payment notification
			}
			// Signature verified but event is irrelevant — acknowledge to stop retries
			verifiedButIrrelevant = true
			break
		}
	}
	if notification == nil {
		if verifiedButIrrelevant {
			return nil // unrecognized event type — return success so platform stops retrying
		}
		return fmt.Errorf("notification verification failed for provider %s", providerKey)
	}

	if notification.Status != "success" {
		slog.Warn("payment notification with non-success status", "tradeNo", notification.TradeNo, "status", notification.Status)
		return nil
	}

	// Find order by trade number
	orderID, err := strconv.ParseInt(notification.OrderID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID in notification: %s", notification.OrderID)
	}

	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order %d: %w", orderID, err)
	}

	// Dispatch based on current order status
	switch order.Status {
	case domain.PaymentOrderStatusPending, domain.PaymentOrderStatusExpired:
		return s.handleNotificationForPendingOrder(ctx, order, notification)

	case domain.PaymentOrderStatusFailed:
		// Retry fulfillment using the payment platform's retry notification
		return s.handleNotificationRetryFulfillment(ctx, order)

	case domain.PaymentOrderStatusPaid, domain.PaymentOrderStatusRecharging:
		// In progress — tell payment platform to retry later
		return fmt.Errorf("order %d is still being processed (status=%s)", orderID, order.Status)

	case domain.PaymentOrderStatusCompleted, domain.PaymentOrderStatusRefunded:
		// Already done — acknowledge to stop retries
		return nil

	default:
		// Unexpected status (cancelled etc.) — acknowledge to stop retries, but audit-log for reconciliation.
		slog.Warn("notification for order in unexpected status", "orderID", orderID, "status", order.Status)
		s.writeAuditLog(ctx, orderID, "NOTIFICATION_IGNORED",
			fmt.Sprintf("received payment notification while order status=%s, tradeNo=%s", order.Status, notification.TradeNo))
		return nil
	}
}

// handleNotificationForPendingOrder validates the amount, atomically confirms payment, and fulfills the order.
func (s *PaymentOrderService) handleNotificationForPendingOrder(ctx context.Context, order *PaymentOrder, notification *PaymentNotification) error {
	orderID := order.ID

	// Validate paid amount against expected amount (0.01 CNY tolerance for rounding)
	expectedAmount := order.Amount
	if order.PayAmount != nil {
		expectedAmount = *order.PayAmount
	}
	diff := notification.Amount.Sub(expectedAmount).Abs()
	amountTolerance := decimal.NewFromFloat(0.01)
	if diff.GreaterThan(amountTolerance) {
		s.writeAuditLog(ctx, orderID, "PAYMENT_AMOUNT_MISMATCH",
			fmt.Sprintf("expected=%s, paid=%s, diff=%s, tradeNo=%s",
				expectedAmount, notification.Amount, diff, notification.TradeNo))
		return fmt.Errorf("amount mismatch: expected %s, got %s (diff %s)", expectedAmount, notification.Amount, diff)
	}

	// Atomic CAS: PENDING (or EXPIRED within grace period) → PAID
	now := time.Now()
	graceMins := s.configService.GetGracePeriodMinutes(ctx)
	graceDeadline := now.Add(-time.Duration(graceMins) * time.Minute)
	ok, err := s.orderRepo.ConfirmPaidCAS(ctx, orderID, notification.TradeNo, notification.Amount, now, graceDeadline)
	if err != nil {
		return fmt.Errorf("confirm paid CAS: %w", err)
	}
	if !ok {
		// Re-fetch to check current status
		current, refetchErr := s.orderRepo.GetByID(ctx, orderID)
		if refetchErr != nil {
			return fmt.Errorf("failed to re-fetch order %d after CAS failure: %w", orderID, refetchErr)
		}
		if current != nil {
			switch current.Status {
			case domain.PaymentOrderStatusCompleted, domain.PaymentOrderStatusRefunded:
				return nil // already done
			case domain.PaymentOrderStatusPaid, domain.PaymentOrderStatusRecharging:
				return fmt.Errorf("order %d is being processed concurrently", orderID)
			}
		}
		return nil // stop retries for other statuses
	}

	s.writeAuditLog(ctx, orderID, "ORDER_PAID",
		fmt.Sprintf("tradeNo=%s, expected=%s, paid=%s", notification.TradeNo, expectedAmount, notification.Amount))

	// Transition PAID → RECHARGING → fulfillment
	return s.executeFulfillment(ctx, orderID, order)
}

// handleNotificationRetryFulfillment retries fulfillment for a previously failed order.
func (s *PaymentOrderService) handleNotificationRetryFulfillment(ctx context.Context, order *PaymentOrder) error {
	slog.Info("retrying fulfillment via notification", "orderID", order.ID)
	return s.executeFulfillment(ctx, order.ID, order)
}

// executeFulfillment transitions an order through RECHARGING → COMPLETED (or FAILED).
func (s *PaymentOrderService) executeFulfillment(ctx context.Context, orderID int64, order *PaymentOrder) error {
	// Try all valid source statuses for RECHARGING transition to avoid stale-state issues.
	// The order object may be stale by the time we reach here, so we attempt CAS from
	// every status that is eligible for fulfillment (PAID, FAILED).
	var ok bool
	for _, fromStatus := range []string{domain.PaymentOrderStatusPaid, domain.PaymentOrderStatusFailed} {
		var casErr error
		ok, casErr = s.orderRepo.UpdateStatusCAS(ctx, orderID, fromStatus, domain.PaymentOrderStatusRecharging)
		if casErr != nil {
			slog.Error("CRITICAL: CAS failed during fulfillment — paid order may be lost",
				"orderID", orderID, "fromStatus", fromStatus, "error", casErr)
			return fmt.Errorf("fulfillment CAS failed: %w", casErr)
		}
		if ok {
			break
		}
	}
	if !ok {
		return nil // another goroutine is handling it, or order is in a terminal state
	}

	// Re-fetch order to get current state after CAS succeeded
	fresh, err := s.orderRepo.GetByID(ctx, orderID)
	if err == nil {
		order = fresh
	}

	// Fulfill the order
	if err := s.fulfillOrder(ctx, order); err != nil {
		slog.Error("failed to fulfill order", "orderID", orderID, "error", err)
		failedNow := time.Now()
		order.Status = domain.PaymentOrderStatusFailed
		order.FailedAt = &failedNow
		failedReason := err.Error()
		order.FailedReason = &failedReason
		if updateErr := s.orderRepo.Update(ctx, order); updateErr != nil {
			slog.Error("CRITICAL: failed to persist FAILED status after fulfillment error",
				"orderID", orderID, "fulfillmentErr", err, "updateErr", updateErr)
		}
		s.writeAuditLog(ctx, orderID, "ORDER_FAILED", fmt.Sprintf("fulfillment error: %v", err))
		return nil // Don't return error to webhook caller — let platform retry
	}

	// Update to completed
	completedNow := time.Now()
	order.Status = domain.PaymentOrderStatusCompleted
	order.CompletedAt = &completedNow
	if updateErr := s.orderRepo.Update(ctx, order); updateErr != nil {
		slog.Error("CRITICAL: fulfillment succeeded but failed to persist COMPLETED status — may cause duplicate fulfillment on retry",
			"orderID", orderID, "error", updateErr)
	}
	s.writeAuditLog(ctx, orderID, "ORDER_COMPLETED", "")

	return nil
}

// RefundOrder initiates a refund for a paid order using two-phase commit:
// Phase 1: Deduct balance / revoke subscription (safe side — can be restored if gateway fails)
// Phase 2: Call payment gateway refund
// If gateway fails, restore the deduction. If restore also fails, mark REFUND_FAILED for manual intervention.
func (s *PaymentOrderService) RefundOrder(ctx context.Context, req RefundOrderRequest) error {
	order, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}

	// Must be in a refundable state
	switch order.Status {
	case domain.PaymentOrderStatusCompleted, domain.PaymentOrderStatusPaid,
		domain.PaymentOrderStatusRefundRequested, domain.PaymentOrderStatusRefundFailed,
		domain.PaymentOrderStatusPartiallyRefunded:
		// OK — proceed with refund
	default:
		return infraerrors.Conflict("ORDER_NOT_REFUNDABLE", "order is not in a refundable status")
	}
	if order.PaymentTradeNo == nil || *order.PaymentTradeNo == "" {
		return infraerrors.BadRequest("NO_TRADE_NO", "order has no payment trade number")
	}

	// Get provider instance
	if order.ProviderInstanceID == nil {
		return infraerrors.BadRequest("NO_PROVIDER_INSTANCE", "order has no provider instance")
	}
	inst, err := s.instanceRepo.GetByID(ctx, *order.ProviderInstanceID)
	if err != nil {
		return ErrPaymentProviderNotFound
	}
	if !inst.RefundEnabled {
		return infraerrors.BadRequest("REFUND_NOT_ENABLED", "refund is not enabled for this provider instance")
	}

	providerKey := PaymentTypeToProviderKey(order.PaymentType)
	configJSON, err := s.encryptor.Decrypt(inst.Config)
	if err != nil {
		return fmt.Errorf("decrypt instance config: %w", err)
	}
	provider, err := s.registry.CreateProvider(providerKey, configJSON)
	if err != nil {
		return fmt.Errorf("create provider: %w", err)
	}

	// CAS: current status → REFUNDING
	previousStatus := order.Status
	ok, err := s.orderRepo.UpdateStatusCAS(ctx, req.OrderID, previousStatus, domain.PaymentOrderStatusRefunding)
	if err != nil || !ok {
		return infraerrors.Conflict("ORDER_STATUS_CHANGED", "order status changed during refund")
	}

	// Phase 1: Deduct balance or revoke subscription (safe side — easier to restore than to claw back)
	var revokedSubID int64
	deductErr := s.deductForRefund(ctx, order, req.Amount, &revokedSubID)
	if deductErr != nil {
		// Can't deduct — roll back status
		_, _ = s.orderRepo.UpdateStatusCAS(ctx, req.OrderID, domain.PaymentOrderStatusRefunding, previousStatus)
		s.writeAuditLog(ctx, req.OrderID, "REFUND_DEDUCT_FAILED", fmt.Sprintf("error: %v", deductErr))
		return fmt.Errorf("refund deduction failed: %w", deductErr)
	}

	// Phase 2: Call payment gateway refund
	// TotalAmount must be the actual amount charged at the gateway (including fee),
	// not the recharge amount. WeChat Pay validates this against the original transaction.
	totalAmount := order.Amount
	if order.PayAmount != nil {
		totalAmount = *order.PayAmount
	}

	_, refundErr := provider.Refund(ctx, RefundRequest{
		TradeNo:     *order.PaymentTradeNo,
		OrderID:     strconv.FormatInt(order.ID, 10),
		Amount:      req.Amount,
		TotalAmount: totalAmount,
		Reason:      req.Reason,
	})

	if refundErr != nil {
		// Gateway refund failed — try to restore what we deducted
		restoreErr := s.restoreAfterRefundFailure(ctx, order, req.Amount, revokedSubID)
		if restoreErr != nil {
			// Both gateway and restore failed — mark as REFUND_FAILED for manual intervention
			slog.Error("CRITICAL: refund gateway failed AND restore failed",
				"orderID", req.OrderID, "refundErr", refundErr, "restoreErr", restoreErr)
			now := time.Now()
			order.Status = domain.PaymentOrderStatusRefundFailed
			order.FailedAt = &now
			reason := fmt.Sprintf("gateway: %v; restore: %v", refundErr, restoreErr)
			order.FailedReason = &reason
			if updateErr := s.orderRepo.Update(ctx, order); updateErr != nil {
				slog.Error("CRITICAL: failed to persist REFUND_FAILED status",
					"orderID", req.OrderID, "error", updateErr)
			}
			s.writeAuditLog(ctx, req.OrderID, "REFUND_FAILED",
				fmt.Sprintf("gateway error: %v; restore error: %v", refundErr, restoreErr))
			return fmt.Errorf("refund failed and could not restore: %w", refundErr)
		}

		// Restore succeeded — roll back to previous status
		_, _ = s.orderRepo.UpdateStatusCAS(ctx, req.OrderID, domain.PaymentOrderStatusRefunding, previousStatus)
		s.writeAuditLog(ctx, req.OrderID, "REFUND_FAILED", fmt.Sprintf("gateway error: %v (deduction restored)", refundErr))
		return fmt.Errorf("refund failed: %w", refundErr)
	}

	// Both phases succeeded — determine if partial or full refund
	now := time.Now()
	// Accumulate refund amount for partial refunds
	totalRefunded := req.Amount
	if order.RefundAmount != nil {
		totalRefunded = order.RefundAmount.Add(req.Amount)
	}
	if totalRefunded.GreaterThanOrEqual(order.Amount) {
		order.Status = domain.PaymentOrderStatusRefunded
	} else {
		order.Status = domain.PaymentOrderStatusPartiallyRefunded
	}
	order.RefundAmount = &totalRefunded
	order.RefundReason = stringPtrIfNotEmpty(req.Reason)
	order.RefundAt = &now
	if updateErr := s.orderRepo.Update(ctx, order); updateErr != nil {
		slog.Error("CRITICAL: refund succeeded at gateway but failed to persist order status",
			"orderID", req.OrderID, "error", updateErr)
	}

	s.writeAuditLog(ctx, req.OrderID, "ORDER_REFUNDED",
		fmt.Sprintf("amount=%s, reason=%s, operator=%s, partial=%v", req.Amount, req.Reason, req.Operator, req.Amount.LessThan(order.Amount)))
	return nil
}

// RequestRefund is a user-initiated refund request. It transitions the order to REFUND_REQUESTED
// without actually processing the refund — an admin must later call RefundOrder to execute it.
func (s *PaymentOrderService) RequestRefund(ctx context.Context, orderID, userID int64, amount decimal.Decimal, reason string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}
	if order.UserID != userID {
		return ErrPaymentOrderNotFound
	}

	// Only completed orders can be refund-requested by users
	if order.Status != domain.PaymentOrderStatusCompleted {
		return infraerrors.Conflict("ORDER_NOT_REFUNDABLE", "order is not in a refundable status")
	}

	// Validate amount
	if amount.LessThanOrEqual(decimal.Zero) || amount.GreaterThan(order.Amount) {
		return ErrPaymentAmountInvalid
	}

	ok, err := s.orderRepo.UpdateStatusCAS(ctx, orderID, domain.PaymentOrderStatusCompleted, domain.PaymentOrderStatusRefundRequested)
	if err != nil || !ok {
		return infraerrors.Conflict("ORDER_STATUS_CHANGED", "order status changed during refund request")
	}

	// Store refund request details
	order.RefundAmount = &amount
	order.RefundReason = stringPtrIfNotEmpty(reason)
	if updateErr := s.orderRepo.Update(ctx, order); updateErr != nil {
		slog.Error("failed to persist refund request details", "orderID", orderID, "error", updateErr)
	}

	operator := fmt.Sprintf("user:%d", userID)
	s.writeAuditLogWithOperator(ctx, orderID, "REFUND_REQUESTED",
		fmt.Sprintf("amount=%s, reason=%s", amount, reason), operator)

	// Auto-process refund if enabled
	if s.configService.IsAutoRefundEnabled(ctx) {
		if err := s.RefundOrder(ctx, RefundOrderRequest{
			OrderID:  orderID,
			Amount:   amount,
			Reason:   reason,
			Operator: "auto:" + operator,
		}); err != nil {
			slog.Warn("auto-refund failed, awaiting admin action", "orderID", orderID, "error", err)
			// Don't return error — order is already in REFUND_REQUESTED, admin can handle manually
		}
	}

	return nil
}

// RetryRecharge retries fulfillment for a failed order (admin-initiated).
func (s *PaymentOrderService) RetryRecharge(ctx context.Context, orderID int64) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrPaymentOrderNotFound
	}

	// Only FAILED or PAID orders can be retried
	if order.Status != domain.PaymentOrderStatusFailed && order.Status != domain.PaymentOrderStatusPaid {
		return infraerrors.Conflict("ORDER_NOT_RETRYABLE", "order is not in a retryable status")
	}

	s.writeAuditLog(ctx, orderID, "RECHARGE_RETRY", fmt.Sprintf("previous_status=%s", order.Status))

	return s.executeFulfillment(ctx, orderID, order)
}

// deductForRefund subtracts user balance (for balance orders) or revokes subscription (for subscription orders).
func (s *PaymentOrderService) deductForRefund(ctx context.Context, order *PaymentOrder, refundAmount decimal.Decimal, revokedSubID *int64) error {
	switch order.OrderType {
	case domain.PaymentOrderTypeBalance:
		// Deduct the refund amount from user balance
		// Note: UpdateBalance uses float64; we use InexactFloat64 with 2-decimal rounding
		// to minimize precision loss at the boundary.
		return s.userService.UpdateBalance(ctx, order.UserID, -refundAmount.Round(2).InexactFloat64())

	case domain.PaymentOrderTypeSubscription:
		if order.SubscriptionGroupID == nil {
			return fmt.Errorf("missing subscription group ID on order")
		}
		// Find and revoke the active subscription in this group
		sub, err := s.subscriptionService.GetActiveSubscription(ctx, order.UserID, *order.SubscriptionGroupID)
		if err != nil {
			return fmt.Errorf("get active subscription: %w", err)
		}
		if sub != nil {
			*revokedSubID = sub.ID
			return s.subscriptionService.RevokeSubscription(ctx, sub.ID)
		}
		return nil // no active subscription to revoke

	default:
		return nil
	}
}

// restoreAfterRefundFailure reverses the deduction done in Phase 1.
func (s *PaymentOrderService) restoreAfterRefundFailure(ctx context.Context, order *PaymentOrder, refundAmount decimal.Decimal, revokedSubID int64) error {
	switch order.OrderType {
	case domain.PaymentOrderTypeBalance:
		return s.userService.UpdateBalance(ctx, order.UserID, refundAmount.Round(2).InexactFloat64())

	case domain.PaymentOrderTypeSubscription:
		if revokedSubID > 0 && order.SubscriptionGroupID != nil && order.SubscriptionDays != nil {
			_, _, err := s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
				UserID:       order.UserID,
				GroupID:      *order.SubscriptionGroupID,
				ValidityDays: *order.SubscriptionDays,
				AssignedBy:   0,
				Notes:        fmt.Sprintf("restore after refund failure, order #%d", order.ID),
			})
			return err
		}
		return nil

	default:
		return nil
	}
}

// HasActiveOrdersForProviderInstance checks if there are PENDING/PAID/RECHARGING orders
// using the given provider instance. Used to prevent unsafe credential changes or deletion.
func (s *PaymentOrderService) HasActiveOrdersForProviderInstance(ctx context.Context, instanceID int64) (bool, error) {
	count, err := s.orderRepo.CountActiveByProviderInstanceID(ctx, instanceID)
	return count > 0, err
}

// HasActiveOrdersForPlan checks if there are PENDING/PAID/RECHARGING orders referencing
// the given subscription plan. Used to prevent deletion of plans with in-flight orders.
func (s *PaymentOrderService) HasActiveOrdersForPlan(ctx context.Context, planID int64) (bool, error) {
	count, err := s.orderRepo.CountActiveByPlanID(ctx, planID)
	return count > 0, err
}

// GetOrder gets an order by ID with optional user ownership check.
func (s *PaymentOrderService) GetOrder(ctx context.Context, orderID int64) (*PaymentOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrPaymentOrderNotFound
	}
	return order, nil
}

// ListUserOrders lists orders for a user with pagination.
func (s *PaymentOrderService) ListUserOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListByUserID(ctx, userID, params)
}

// ListOrders lists orders with admin-level filters and pagination.
func (s *PaymentOrderService) ListOrders(ctx context.Context, filter PaymentOrderListFilter, params pagination.PaginationParams) ([]PaymentOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListFiltered(ctx, filter, params)
}

// GetOrderWithAuditLogs retrieves an order and its audit logs.
func (s *PaymentOrderService) GetOrderWithAuditLogs(ctx context.Context, orderID int64) (*PaymentOrder, []PaymentAuditLog, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, nil, ErrPaymentOrderNotFound
	}
	logs, err := s.auditLogRepo.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, nil, fmt.Errorf("list audit logs: %w", err)
	}
	return order, logs, nil
}

// GetDashboardStats returns payment dashboard statistics.
func (s *PaymentOrderService) GetDashboardStats(ctx context.Context, days int) (*RawDashboardData, error) {
	since := time.Now().AddDate(0, 0, -days)
	return s.orderRepo.GetDashboardStats(ctx, since)
}

// --- Internal helpers ---

// fulfillOrder dispatches balance recharge or subscription assignment.
func (s *PaymentOrderService) fulfillOrder(ctx context.Context, order *PaymentOrder) error {
	switch order.OrderType {
	case domain.PaymentOrderTypeBalance:
		return s.fulfillBalanceOrder(ctx, order)
	case domain.PaymentOrderTypeSubscription:
		return s.fulfillSubscriptionOrder(ctx, order)
	default:
		return fmt.Errorf("unknown order type: %s", order.OrderType)
	}
}

func (s *PaymentOrderService) fulfillBalanceOrder(ctx context.Context, order *PaymentOrder) error {
	// Create and redeem a balance recharge code
	code := &RedeemCode{
		Code:  order.RechargeCode,
		Type:  RedeemTypeBalance,
		Value: order.Amount.Round(2).InexactFloat64(),
		Notes: fmt.Sprintf("payment order #%d", order.ID),
	}
	if err := s.redeemService.CreateCode(ctx, code); err != nil {
		return fmt.Errorf("create redeem code: %w", err)
	}

	_, err := s.redeemService.Redeem(ctx, order.UserID, order.RechargeCode)
	if err != nil {
		return fmt.Errorf("redeem code: %w", err)
	}
	return nil
}

func (s *PaymentOrderService) fulfillSubscriptionOrder(ctx context.Context, order *PaymentOrder) error {
	if order.SubscriptionGroupID == nil || order.SubscriptionDays == nil {
		return fmt.Errorf("missing subscription group or days")
	}

	validityDays := *order.SubscriptionDays

	// For renewals, recalculate validity days from the active subscription's expiry date
	// (matching TypeScript behavior where month-based plans compute calendar days from expiry).
	if order.PlanID != nil {
		activeSub, err := s.subscriptionService.GetActiveSubscription(ctx, order.UserID, *order.SubscriptionGroupID)
		if err == nil && activeSub != nil {
			plan, err := s.planRepo.GetByID(ctx, *order.PlanID)
			if err == nil && plan != nil {
				validityDays = computeValidityDays(plan.ValidityDays, plan.ValidityUnit, activeSub.ExpiresAt)
			}
		}
	}

	_, _, err := s.subscriptionService.AssignOrExtendSubscription(ctx, &AssignSubscriptionInput{
		UserID:       order.UserID,
		GroupID:      *order.SubscriptionGroupID,
		ValidityDays: validityDays,
		AssignedBy:   0,
		Notes:        fmt.Sprintf("payment order #%d", order.ID),
	})
	return err
}

// queryProviderPaymentStatus checks with the payment provider whether an order was actually paid.
// Returns (true, notification) if the provider reports the order as paid.
// Returns (false, nil) if not paid or on any error (fail-open).
func (s *PaymentOrderService) queryProviderPaymentStatus(ctx context.Context, order *PaymentOrder) (bool, *PaymentNotification) {
	if order.ProviderInstanceID == nil {
		return false, nil
	}

	// Bound the provider query so a slow upstream doesn't block cancel indefinitely.
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	providerKey := PaymentTypeToProviderKey(order.PaymentType)
	inst, err := s.instanceRepo.GetByID(queryCtx, *order.ProviderInstanceID)
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
	tradeNo := strconv.FormatInt(order.ID, 10)
	if providerKey == domain.PaymentProviderStripe && order.PaymentTradeNo != nil && *order.PaymentTradeNo != "" {
		tradeNo = *order.PaymentTradeNo
	}

	result, err := provider.QueryOrder(queryCtx, tradeNo)
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

func (s *PaymentOrderService) tryCancelAtProvider(ctx context.Context, order *PaymentOrder) {
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

func (s *PaymentOrderService) buildPaymentSubject(ctx context.Context, order *PaymentOrder, plan *SubscriptionPlan) string {
	if plan != nil {
		if plan.ProductName != nil && *plan.ProductName != "" {
			return *plan.ProductName
		}
		// Subscription fallback: "Sub2API 订阅 {groupName}"
		if plan.GroupID != nil {
			if group, err := s.groupRepo.GetByID(ctx, *plan.GroupID); err == nil {
				return "Sub2API 订阅 " + group.Name
			}
		}
		return "Sub2API 订阅 " + plan.Name
	}

	// Balance order: include the actual pay amount in the subject, matching TypeScript:
	//   `${prefix || ''} ${payAmountStr} ${suffix || ''}`.trim()
	payAmountStr := order.Amount.String()
	if order.PayAmount != nil {
		payAmountStr = order.PayAmount.String()
	}

	prefix := strings.TrimSpace(s.configService.GetProductNamePrefix(ctx))
	suffix := strings.TrimSpace(s.configService.GetProductNameSuffix(ctx))

	if prefix != "" || suffix != "" {
		parts := make([]string, 0, 3)
		if prefix != "" {
			parts = append(parts, prefix)
		}
		parts = append(parts, payAmountStr)
		if suffix != "" {
			parts = append(parts, suffix)
		}
		return strings.Join(parts, " ")
	}
	return "Sub2API " + payAmountStr + " CNY"
}

func (s *PaymentOrderService) writeAuditLog(ctx context.Context, orderID int64, action, detail string) {
	s.writeAuditLogWithOperator(ctx, orderID, action, detail, "")
}

func (s *PaymentOrderService) writeAuditLogWithOperator(ctx context.Context, orderID int64, action, detail, operator string) {
	log := &PaymentAuditLog{
		OrderID:  orderID,
		Action:   action,
		Detail:   stringPtrIfNotEmpty(detail),
		Operator: stringPtrIfNotEmpty(operator),
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		slog.Warn("failed to write payment audit log", "orderID", orderID, "action", action, "error", err)
	}
}

func generateRechargeCode(orderID int64) string {
	idStr := strconv.FormatInt(orderID, 10)
	const prefix = "s2p_"
	const randomLen = 16 // 8 bytes → 16 hex chars
	maxIDLen := 32 - len(prefix) - randomLen
	if len(idStr) > maxIDLen {
		idStr = idStr[:maxIDLen]
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + idStr + hex.EncodeToString(b)
}

// computeValidityDays converts a validity value + unit to actual days.
// Matches the TypeScript computeValidityDays: day=value, week=value*7, month=calendar diff.
// The optional refDate controls the starting point for month calculations (defaults to now).
func computeValidityDays(value int, unit string, refDate ...time.Time) int {
	switch unit {
	case "week":
		return value * 7
	case "month":
		base := time.Now()
		if len(refDate) > 0 && !refDate[0].IsZero() {
			base = refDate[0]
		}
		target := base.AddDate(0, value, 0)
		return int(target.Sub(base).Hours()/24 + 0.5)
	default: // "day" or unrecognized
		return value
	}
}

// checkCancelRateLimit checks whether the user has exceeded the cancellation rate limit.
// Matches the TypeScript cancel rate limiting logic: counts ORDER_CANCELLED audit logs
// by user within a configurable time window.
func (s *PaymentOrderService) checkCancelRateLimit(ctx context.Context, userID int64) error {
	if !s.configService.IsCancelRateLimitEnabled(ctx) {
		return nil
	}

	maxCount := s.configService.GetCancelRateLimitMax(ctx)
	if maxCount <= 0 {
		return nil
	}

	windowSize := s.configService.GetCancelRateLimitWindow(ctx)
	unit := s.configService.GetCancelRateLimitUnit(ctx)
	mode := s.configService.GetCancelRateLimitWindowMode(ctx)

	windowStart := cancelRateLimitWindowStart(time.Now(), windowSize, unit, mode)
	operator := fmt.Sprintf("user:%d", userID)

	count, err := s.auditLogRepo.CountByActionAndOperatorSince(ctx, "ORDER_CANCELLED", operator, windowStart)
	if err != nil {
		slog.Warn("cancel rate limit check failed, allowing through", "userID", userID, "error", err)
		return nil // fail-open
	}

	if count >= maxCount {
		return ErrCancelRateLimited
	}
	return nil
}

// cancelRateLimitWindowStart computes the window start time for cancel rate limiting.
func cancelRateLimitWindowStart(now time.Time, windowSize int, unit, mode string) time.Time {
	nowLocal := now.In(shanghaiLoc)

	if mode == "fixed" {
		switch unit {
		case "minute":
			start := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), nowLocal.Hour(), nowLocal.Minute(), 0, 0, shanghaiLoc)
			return start.Add(-time.Duration(windowSize-1) * time.Minute).UTC()
		case "hour":
			start := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), nowLocal.Hour(), 0, 0, 0, shanghaiLoc)
			return start.Add(-time.Duration(windowSize-1) * time.Hour).UTC()
		default: // day
			start := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, shanghaiLoc)
			return start.AddDate(0, 0, -(windowSize - 1)).UTC()
		}
	}

	// Rolling mode
	switch unit {
	case "minute":
		return now.Add(-time.Duration(windowSize) * time.Minute)
	case "hour":
		return now.Add(-time.Duration(windowSize) * time.Hour)
	default: // day
		return now.Add(-time.Duration(windowSize) * 24 * time.Hour)
	}
}

func stringPtr(s string) *string {
	return &s
}

func stringPtrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
