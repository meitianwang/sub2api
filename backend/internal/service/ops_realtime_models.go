package service

import "time"

// PlatformAvailability aggregates account availability by platform.
type PlatformAvailability struct {
	Platform       string `json:"platform"`
	TotalAccounts  int64  `json:"total_accounts"`
	AvailableCount int64  `json:"available_count"`
	RateLimitCount int64  `json:"rate_limit_count"`
	ErrorCount     int64  `json:"error_count"`
}

// GroupAvailability aggregates account availability by group.
type GroupAvailability struct {
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	TotalAccounts  int64  `json:"total_accounts"`
	AvailableCount int64  `json:"available_count"`
	RateLimitCount int64  `json:"rate_limit_count"`
	ErrorCount     int64  `json:"error_count"`
}

// AccountAvailability represents current availability for a single account.
type AccountAvailability struct {
	AccountID   int64  `json:"account_id"`
	AccountName string `json:"account_name"`
	Platform    string `json:"platform"`
	GroupID     int64  `json:"group_id"`
	GroupName   string `json:"group_name"`

	Status string `json:"status"`

	IsAvailable   bool `json:"is_available"`
	IsRateLimited bool `json:"is_rate_limited"`
	IsOverloaded  bool `json:"is_overloaded"`
	HasError      bool `json:"has_error"`

	RateLimitResetAt       *time.Time `json:"rate_limit_reset_at"`
	RateLimitRemainingSec  *int64     `json:"rate_limit_remaining_sec"`
	OverloadUntil          *time.Time `json:"overload_until"`
	OverloadRemainingSec   *int64     `json:"overload_remaining_sec"`
	ErrorMessage           string     `json:"error_message"`
	TempUnschedulableUntil *time.Time `json:"temp_unschedulable_until,omitempty"`
}
