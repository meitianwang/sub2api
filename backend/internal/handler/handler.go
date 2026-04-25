package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
)

// AdminHandlers contains all admin-related HTTP handlers
type AdminHandlers struct {
	Dashboard             *admin.DashboardHandler
	User                  *admin.UserHandler
	Group                 *admin.GroupHandler
	Account               *admin.AccountHandler
	Announcement          *admin.AnnouncementHandler
	DataManagement        *admin.DataManagementHandler
	Backup                *admin.BackupHandler
	OAuth                 *admin.OAuthHandler
	Proxy                 *admin.ProxyHandler
	Redeem                *admin.RedeemHandler
	Promo                 *admin.PromoHandler
	Setting               *admin.SettingHandler
	Ops                   *admin.OpsHandler
	System                *admin.SystemHandler
	Usage                 *admin.UsageHandler
	UserAttribute         *admin.UserAttributeHandler
	ErrorPassthrough      *admin.ErrorPassthroughHandler
	TLSFingerprintProfile *admin.TLSFingerprintProfileHandler
	APIKey                *admin.AdminAPIKeyHandler
	ScheduledTest         *admin.ScheduledTestHandler
	PaymentOrder          *admin.PaymentOrderHandler
	PaymentRefund         *admin.PaymentRefundHandler
	PaymentConfig         *admin.PaymentConfigHandler
	PaymentProviderInstance *admin.PaymentProviderInstanceHandler
	PaymentChannel        *admin.PaymentChannelHandler
	PaymentDashboard      *admin.PaymentDashboardHandler
}

// Handlers contains all HTTP handlers
type Handlers struct {
	Auth            *AuthHandler
	User            *UserHandler
	APIKey          *APIKeyHandler
	Usage           *UsageHandler
	Redeem          *RedeemHandler
	Announcement    *AnnouncementHandler
	Admin           *AdminHandlers
	Gateway         *GatewayHandler
	Setting         *SettingHandler
	Totp            *TotpHandler
	Payment         *PaymentHandler
	PaymentWebhook  *PaymentWebhookHandler
}

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string
	BuildType string // "source" for manual builds, "release" for CI builds
}
