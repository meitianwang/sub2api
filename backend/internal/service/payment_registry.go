package service

import (
	"fmt"
	"sync"

	"github.com/Wei-Shaw/sub2api/internal/domain"
)

// PaymentProviderFactory creates a PaymentProvider from decrypted config JSON.
type PaymentProviderFactory func(configJSON string) (PaymentProvider, error)

// PaymentProviderRegistry maps provider keys to their factory functions.
type PaymentProviderRegistry struct {
	mu        sync.RWMutex
	factories map[string]PaymentProviderFactory
}

// NewPaymentProviderRegistry creates a new empty registry.
func NewPaymentProviderRegistry() *PaymentProviderRegistry {
	return &PaymentProviderRegistry{
		factories: make(map[string]PaymentProviderFactory),
	}
}

// Register adds a provider factory for the given provider key.
func (r *PaymentProviderRegistry) Register(providerKey string, factory PaymentProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[providerKey] = factory
}

// CreateProvider creates a PaymentProvider instance from decrypted config JSON.
func (r *PaymentProviderRegistry) CreateProvider(providerKey, configJSON string) (PaymentProvider, error) {
	r.mu.RLock()
	factory, ok := r.factories[providerKey]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no factory registered for provider key: %s", providerKey)
	}
	return factory(configJSON)
}

// HasProvider reports whether a factory is registered for the given key.
func (r *PaymentProviderRegistry) HasProvider(providerKey string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[providerKey]
	return ok
}

// PaymentTypeToProviderKey maps a PaymentType constant to its provider key.
// "alipay"/"wxpay" (via EasyPay aggregator) → "easypay"
// "alipay_direct" → "alipay"
// "wxpay_direct" → "wxpay"
// "stripe" → "stripe"
func PaymentTypeToProviderKey(paymentType string) string {
	switch paymentType {
	case domain.PaymentTypeAlipay, domain.PaymentTypeWxpay:
		return domain.PaymentProviderEasyPay
	case domain.PaymentTypeAlipayDirect:
		return domain.PaymentProviderAlipay
	case domain.PaymentTypeWxpayDirect:
		return domain.PaymentProviderWxpay
	case domain.PaymentTypeStripe:
		return domain.PaymentProviderStripe
	default:
		return ""
	}
}
