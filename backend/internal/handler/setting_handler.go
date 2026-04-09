package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
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

// publicModelEntry 一条模型记录 = 模型 + 分组 + 价格
type publicModelEntry struct {
	ModelID     string  `json:"model_id"`
	DisplayName string  `json:"display_name"`
	Provider    string  `json:"provider"`
	GroupID     int64   `json:"group_id"`
	GroupName   string  `json:"group_name"`
	InputPrice  float64 `json:"input_price"`  // USD/M tokens
	OutputPrice float64 `json:"output_price"` // USD/M tokens
}

// GetPublicModels 获取公开模型列表
// GET /api/v1/settings/models
func (h *SettingHandler) GetPublicModels(c *gin.Context) {
	var items []publicModelEntry

	if h.groupRepo != nil {
		activeGroups, err := h.groupRepo.ListActive(c.Request.Context())
		if err == nil {
			for _, g := range activeGroups {
				if g.IsExclusive || g.ModelPricing == nil {
					continue
				}
				for modelID, entry := range g.ModelPricing {
					provider := guessProvider(modelID)
					items = append(items, publicModelEntry{
						ModelID:     modelID,
						DisplayName: modelID, // 先用 ID，下面替换
						Provider:    provider,
						GroupID:     g.ID,
						GroupName:   g.Name,
						InputPrice:  entry.SellInputPrice,
						OutputPrice: entry.SellOutputPrice,
					})
				}
			}
		}
	}

	// 用已知模型列表补充 display_name
	displayNames := buildDisplayNameMap()
	for i := range items {
		if name, ok := displayNames[items[i].ModelID]; ok {
			items[i].DisplayName = name
		}
	}

	if items == nil {
		items = []publicModelEntry{}
	}

	response.Success(c, items)
}

func guessProvider(modelID string) string {
	if len(modelID) >= 6 && modelID[:6] == "claude" {
		return "claude"
	}
	if len(modelID) >= 3 && modelID[:3] == "gpt" {
		return "openai"
	}
	if len(modelID) >= 5 && modelID[:5] == "codex" {
		return "openai"
	}
	if len(modelID) >= 6 && modelID[:6] == "gemini" {
		return "gemini"
	}
	return "other"
}

func buildDisplayNameMap() map[string]string {
	m := make(map[string]string)
	for _, mod := range openai.DefaultModels {
		m[mod.ID] = mod.DisplayName
	}
	return m
}
