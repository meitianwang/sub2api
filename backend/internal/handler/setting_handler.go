package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler 公开设置处理器（无需认证）
type SettingHandler struct {
	settingService *service.SettingService
	groupRepo      service.GroupRepository
	version        string
}

// NewSettingHandler 创建公开设置处理器
func NewSettingHandler(settingService *service.SettingService, version string, groupRepo service.GroupRepository) *SettingHandler {
	return &SettingHandler{
		settingService: settingService,
		version:        version,
		groupRepo:      groupRepo,
	}
}

// GetPublicSettings 获取公开设置
// GET /api/v1/settings/public
func (h *SettingHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settingService.GetPublicSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.PublicSettings{
		RegistrationEnabled:              settings.RegistrationEnabled,
		EmailVerifyEnabled:               settings.EmailVerifyEnabled,
		RegistrationEmailSuffixWhitelist: settings.RegistrationEmailSuffixWhitelist,
		PromoCodeEnabled:                 settings.PromoCodeEnabled,
		PasswordResetEnabled:             settings.PasswordResetEnabled,
		InvitationCodeEnabled:            settings.InvitationCodeEnabled,
		TotpEnabled:                      settings.TotpEnabled,
		TurnstileEnabled:                 settings.TurnstileEnabled,
		TurnstileSiteKey:                 settings.TurnstileSiteKey,
		SiteName:                         settings.SiteName,
		SiteLogo:                         settings.SiteLogo,
		SiteSubtitle:                     settings.SiteSubtitle,
		APIBaseURL:                       settings.APIBaseURL,
		ContactInfo:                      settings.ContactInfo,
		DocURL:                           settings.DocURL,
		HomeContent:                      settings.HomeContent,
		HideCcsImportButton:              settings.HideCcsImportButton,
		PurchaseSubscriptionEnabled:      settings.PurchaseSubscriptionEnabled,
		PurchaseSubscriptionURL:          settings.PurchaseSubscriptionURL,
		CustomMenuItems:                  dto.ParseUserVisibleMenuItems(settings.CustomMenuItems),
		CustomEndpoints:                  dto.ParseCustomEndpoints(settings.CustomEndpoints),
		LinuxDoOAuthEnabled:              settings.LinuxDoOAuthEnabled,
		SoraClientEnabled:                settings.SoraClientEnabled,
		BackendModeEnabled:               settings.BackendModeEnabled,
		Version:                          h.version,
	})
}

// publicModelPricing 公开模型价格信息
type publicModelPricing struct {
	InputPrice  *float64 `json:"input_price,omitempty"`  // USD/M tokens
	OutputPrice *float64 `json:"output_price,omitempty"` // USD/M tokens
}

// publicGroup 公开分组信息
type publicGroup struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Platform       string  `json:"platform"`
	RateMultiplier float64 `json:"rate_multiplier"`
}

// publicModelsResponse 公开模型列表响应
type publicModelsResponse struct {
	Models []publicModelItem            `json:"models"`
	Groups []publicGroupWithPricing     `json:"groups"`
}

// publicGroupWithPricing 公开分组信息（含该分组的模型定价）
type publicGroupWithPricing struct {
	ID             int64                         `json:"id"`
	Name           string                        `json:"name"`
	Platform       string                        `json:"platform"`
	RateMultiplier float64                       `json:"rate_multiplier"`
	ModelPricing   map[string]publicModelPricing  `json:"model_pricing,omitempty"` // modelID -> pricing
}

// publicModelItem 公开模型项
type publicModelItem struct {
	ID          string              `json:"id"`
	DisplayName string              `json:"display_name"`
	Provider    string              `json:"provider"`
	CreatedAt   string              `json:"created_at,omitempty"`
	Pricing     *publicModelPricing `json:"pricing,omitempty"` // 默认定价（官方价）
	GroupIDs    []int64             `json:"group_ids,omitempty"`
}

// GetPublicModels 获取公开模型列表（含分组和价格）
// GET /api/v1/settings/models
func (h *SettingHandler) GetPublicModels(c *gin.Context) {
	// 收集所有模型
	type modelDef struct {
		id, displayName, provider, createdAt string
	}
	var allModels []modelDef
	for _, m := range antigravity.GetPublicClaudeModels() {
		allModels = append(allModels, modelDef{m.ID, m.DisplayName, m.Provider, m.CreatedAt})
	}
	for _, m := range antigravity.GetPublicGeminiModels() {
		allModels = append(allModels, modelDef{m.ID, m.DisplayName, m.Provider, m.CreatedAt})
	}
	for _, m := range openai.DefaultModels {
		allModels = append(allModels, modelDef{m.ID, m.DisplayName, "openai", ""})
	}

	// 查询活跃分组，收集每个分组的模型定价和模型所属分组
	var groups []publicGroupWithPricing
	modelGroupIDs := make(map[string][]int64) // modelID(lower) -> groupIDs

	if h.groupRepo != nil {
		activeGroups, err := h.groupRepo.ListActive(c.Request.Context())
		if err == nil {
			for _, g := range activeGroups {
				if g.IsExclusive {
					continue
				}
				pg := publicGroupWithPricing{
					ID:             g.ID,
					Name:           g.Name,
					Platform:       g.Platform,
					RateMultiplier: g.RateMultiplier,
				}
				// 导出该分组的 per-model 卖价（仅 sell price）
				if g.ModelPricing != nil {
					for _, m := range allModels {
						entry := g.ModelPricing.GetPricing(m.id)
						if entry == nil {
							continue
						}
						if pg.ModelPricing == nil {
							pg.ModelPricing = make(map[string]publicModelPricing)
						}
						pg.ModelPricing[strings.ToLower(m.id)] = publicModelPricing{
							InputPrice:  &entry.SellInputPrice,
							OutputPrice: &entry.SellOutputPrice,
						}
						modelGroupIDs[strings.ToLower(m.id)] = append(modelGroupIDs[strings.ToLower(m.id)], g.ID)
					}
				}
				groups = append(groups, pg)
			}
		}
	}

	// 构建响应：每个模型默认显示官方定价，分组自定义价格在 groups 里
	result := make([]publicModelItem, 0, len(allModels))
	for _, m := range allModels {
		item := publicModelItem{
			ID:          m.id,
			DisplayName: m.displayName,
			Provider:    m.provider,
			CreatedAt:   m.createdAt,
			GroupIDs:    modelGroupIDs[strings.ToLower(m.id)],
		}
		// 默认定价（官方价）
		if dp := service.GetDefaultPublicPricing(m.id); dp != nil {
			item.Pricing = &publicModelPricing{
				InputPrice:  &dp.InputPrice,
				OutputPrice: &dp.OutputPrice,
			}
		}
		result = append(result, item)
	}

	response.Success(c, publicModelsResponse{
		Models: result,
		Groups: groups,
	})
}
