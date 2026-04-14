package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"
)

// SelectedInstance holds the result of instance selection.
type SelectedInstance struct {
	InstanceID int64
	ConfigJSON string // decrypted provider config
}

// PaymentLoadBalancer selects a provider instance using round-robin or least-amount strategy.
type PaymentLoadBalancer struct {
	instanceRepo  PaymentProviderInstanceRepository
	orderRepo     PaymentOrderRepository
	configService *PaymentConfigService
	encryptor     SecretEncryptor
	rrCounter     atomic.Uint64
}

// NewPaymentLoadBalancer creates a new load balancer.
func NewPaymentLoadBalancer(
	instanceRepo PaymentProviderInstanceRepository,
	orderRepo PaymentOrderRepository,
	configService *PaymentConfigService,
	encryptor SecretEncryptor,
) *PaymentLoadBalancer {
	return &PaymentLoadBalancer{
		instanceRepo:  instanceRepo,
		orderRepo:     orderRepo,
		configService: configService,
		encryptor:     encryptor,
	}
}

// SelectInstance picks an eligible provider instance for the given provider key, payment type, and amount.
// Returns nil if no instances are eligible.
func (lb *PaymentLoadBalancer) SelectInstance(
	ctx context.Context,
	providerKey string,
	paymentType string,
	amount decimal.Decimal,
) (*SelectedInstance, error) {
	instances, err := lb.instanceRepo.ListEnabledByProviderKey(ctx, providerKey)
	if err != nil {
		return nil, fmt.Errorf("list provider instances: %w", err)
	}
	if len(instances) == 0 {
		return nil, nil
	}

	eligible, usageMap, err := lb.filterEligible(ctx, instances, paymentType, amount)
	if err != nil {
		return nil, err
	}
	if len(eligible) == 0 {
		return nil, nil
	}

	var selected *PaymentProviderInstance
	strategy := lb.configService.GetLoadBalanceStrategy(ctx)

	switch strategy {
	case "least_amount", "least-amount":
		selected = lb.selectLeastAmount(eligible, usageMap)
	default: // round-robin
		idx := lb.rrCounter.Add(1) - 1
		selected = &eligible[idx%uint64(len(eligible))]
	}

	configJSON, err := lb.encryptor.Decrypt(selected.Config)
	if err != nil {
		return nil, fmt.Errorf("decrypt instance config: %w", err)
	}

	return &SelectedInstance{
		InstanceID: selected.ID,
		ConfigJSON: configJSON,
	}, nil
}

// filterEligible applies supportedTypes, daily limits, and single amount limits.
// It also returns a pre-fetched usage map for selectLeastAmount to reuse.
func (lb *PaymentLoadBalancer) filterEligible(
	ctx context.Context,
	instances []PaymentProviderInstance,
	paymentType string,
	amount decimal.Decimal,
) ([]PaymentProviderInstance, map[int64]decimal.Decimal, error) {
	// Pre-filter by supported types and single amount limits (no DB needed)
	var candidates []PaymentProviderInstance
	var needDailyCheck []int64
	for i := range instances {
		inst := &instances[i]

		// Filter by supported types
		if inst.SupportedTypes != "" {
			types := strings.Split(inst.SupportedTypes, ",")
			found := false
			for _, t := range types {
				if strings.TrimSpace(t) == paymentType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by single amount limits (no DB needed)
		limits := lb.parseInstanceLimits(inst.Limits, paymentType)
		if limits != nil {
			if limits.SingleMin != nil && amount.LessThan(*limits.SingleMin) {
				continue
			}
			if limits.SingleMax != nil && amount.GreaterThan(*limits.SingleMax) {
				continue
			}
		}

		candidates = append(candidates, *inst)
		needDailyCheck = append(needDailyCheck, inst.ID)
	}

	if len(candidates) == 0 {
		return nil, nil, nil
	}

	// Batch query daily usage for all candidates in one DB call
	bizDayStart := getBizDayStartUTC(time.Now().UTC())
	usageMap, err := lb.orderRepo.SumDailyPaidByInstanceIDs(ctx, needDailyCheck, bizDayStart)
	if err != nil {
		return nil, nil, fmt.Errorf("batch sum daily paid: %w", err)
	}

	// Filter by daily limits using the batched result
	var result []PaymentProviderInstance
	for i := range candidates {
		inst := &candidates[i]
		limits := lb.parseInstanceLimits(inst.Limits, paymentType)
		if limits != nil && limits.DailyLimit != nil && limits.DailyLimit.GreaterThan(decimal.Zero) {
			used := usageMap[inst.ID] // zero value if not in map
			if used.Add(amount).GreaterThan(*limits.DailyLimit) {
				continue
			}
		}
		result = append(result, *inst)
	}

	return result, usageMap, nil
}

// selectLeastAmount picks the instance with the lowest daily usage, reusing a pre-fetched usage map.
func (lb *PaymentLoadBalancer) selectLeastAmount(
	instances []PaymentProviderInstance,
	usageMap map[int64]decimal.Decimal,
) *PaymentProviderInstance {
	var best *PaymentProviderInstance
	bestUsed := decimal.New(1, 18) // large sentinel

	for i := range instances {
		used := usageMap[instances[i].ID]
		if used.LessThan(bestUsed) {
			bestUsed = used
			best = &instances[i]
		}
	}

	return best
}

// parseInstanceLimits extracts the limits for a given payment type from the JSON limits field.
func (lb *PaymentLoadBalancer) parseInstanceLimits(limitsJSON *string, paymentType string) *ProviderInstanceLimits {
	if limitsJSON == nil || *limitsJSON == "" {
		return nil
	}
	var allLimits map[string]*ProviderInstanceLimits
	if err := json.Unmarshal([]byte(*limitsJSON), &allLimits); err != nil {
		return nil
	}
	return allLimits[paymentType]
}

// MethodLimitStatus describes the current availability and limits for a single payment type.
type MethodLimitStatus struct {
	PaymentType string           `json:"paymentType"`
	Available   bool             `json:"available"`
	DailyLimit  decimal.Decimal  `json:"dailyLimit"`  // 0 = unlimited
	DailyUsed   decimal.Decimal  `json:"dailyUsed"`
	Remaining   *decimal.Decimal `json:"remaining"`   // nil = unlimited
	SingleMin   decimal.Decimal  `json:"singleMin"`
	SingleMax   decimal.Decimal  `json:"singleMax"`   // 0 = unlimited
	FeeRate     decimal.Decimal  `json:"feeRate"`
}

// QueryMethodLimits returns current availability and limit status for each given payment type.
func (lb *PaymentLoadBalancer) QueryMethodLimits(ctx context.Context, paymentTypes []string) ([]MethodLimitStatus, error) {
	bizDayStart := getBizDayStartUTC(time.Now().UTC())
	results := make([]MethodLimitStatus, 0, len(paymentTypes))

	for _, pt := range paymentTypes {
		providerKey := PaymentTypeToProviderKey(pt)
		if providerKey == "" {
			results = append(results, MethodLimitStatus{PaymentType: pt, Available: false})
			continue
		}

		// Get global limits from config
		globalDailyLimit := lb.configService.GetGlobalDailyLimit(ctx, pt)
		feeRate := lb.configService.GetFeeRate(ctx, pt, providerKey)

		// Get all enabled instances for this provider
		instances, err := lb.instanceRepo.ListEnabledByProviderKey(ctx, providerKey)
		if err != nil || len(instances) == 0 {
			results = append(results, MethodLimitStatus{PaymentType: pt, Available: false, FeeRate: feeRate})
			continue
		}

		// Filter instances that support this payment type and aggregate limits
		var (
			widestSingleMax    = decimal.Zero
			narrowestSingleMin = decimal.Zero
			anyAvailable       bool
			unlimitedDaily     bool
		)

		for _, inst := range instances {
			limits := lb.parseInstanceLimits(inst.Limits, pt)

			// Check if this instance supports this payment type
			if inst.SupportedTypes != "" {
				types := strings.Split(inst.SupportedTypes, ",")
				found := false
				for _, t := range types {
					if strings.TrimSpace(t) == pt {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			anyAvailable = true

			if limits != nil {
				if limits.SingleMax != nil && limits.SingleMax.IsPositive() && (widestSingleMax.IsZero() || limits.SingleMax.GreaterThan(widestSingleMax)) {
					widestSingleMax = *limits.SingleMax
				}
				if limits.SingleMin != nil && limits.SingleMin.IsPositive() && (narrowestSingleMin.IsZero() || limits.SingleMin.LessThan(narrowestSingleMin)) {
					narrowestSingleMin = *limits.SingleMin
				}
				if limits.DailyLimit == nil || limits.DailyLimit.IsZero() {
					unlimitedDaily = true
				}
			} else {
				unlimitedDaily = true
			}
		}

		// Query global daily usage
		globalUsed := decimal.Zero
		if globalDailyLimit.IsPositive() {
			globalUsed, _ = lb.orderRepo.SumDailyPaidByPaymentType(ctx, pt, bizDayStart)
		}

		status := MethodLimitStatus{
			PaymentType: pt,
			Available:   anyAvailable,
			DailyLimit:  globalDailyLimit,
			DailyUsed:   globalUsed,
			SingleMin:   narrowestSingleMin,
			SingleMax:   widestSingleMax,
			FeeRate:     feeRate,
		}

		if globalDailyLimit.IsPositive() && !unlimitedDaily {
			remaining := globalDailyLimit.Sub(globalUsed)
			if remaining.IsNegative() {
				remaining = decimal.Zero
			}
			status.Remaining = &remaining
			if remaining.IsZero() {
				status.Available = false
			}
		}

		results = append(results, status)
	}

	return results, nil
}

// shanghaiLoc is cached to avoid repeated LoadLocation calls and panic risk in minimal Docker images.
var shanghaiLoc *time.Location

func init() {
	var err error
	shanghaiLoc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// Fallback to fixed +08:00 offset if tzdata is not available
		shanghaiLoc = time.FixedZone("CST", 8*60*60)
	}
}

// getBizDayStartUTC returns the start of the current business day (00:00 Asia/Shanghai) as UTC.
func getBizDayStartUTC(now time.Time) time.Time {
	shanghai := now.In(shanghaiLoc)
	return time.Date(shanghai.Year(), shanghai.Month(), shanghai.Day(), 0, 0, 0, 0, shanghaiLoc).UTC()
}
