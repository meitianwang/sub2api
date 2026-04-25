package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnnouncementTargeting_Matches_EmptyMatchesAll(t *testing.T) {
	var targeting AnnouncementTargeting
	require.True(t, targeting.Matches(0))
	require.True(t, targeting.Matches(123.45))
}

func TestAnnouncementTargeting_NormalizeAndValidate_RejectsEmptyGroup(t *testing.T) {
	targeting := AnnouncementTargeting{
		AnyOf: []AnnouncementConditionGroup{
			{AllOf: nil},
		},
	}
	_, err := targeting.NormalizeAndValidate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrAnnouncementInvalidTarget)
}

func TestAnnouncementTargeting_NormalizeAndValidate_RejectsInvalidCondition(t *testing.T) {
	targeting := AnnouncementTargeting{
		AnyOf: []AnnouncementConditionGroup{
			{
				AllOf: []AnnouncementCondition{
					{Type: "balance", Operator: "between", Value: 10},
				},
			},
		},
	}
	_, err := targeting.NormalizeAndValidate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrAnnouncementInvalidTarget)
}

func TestAnnouncementTargeting_Matches_AndOrSemantics(t *testing.T) {
	targeting := AnnouncementTargeting{
		AnyOf: []AnnouncementConditionGroup{
			{
				AllOf: []AnnouncementCondition{
					{Type: AnnouncementConditionTypeBalance, Operator: AnnouncementOperatorGTE, Value: 100},
				},
			},
			{
				AllOf: []AnnouncementCondition{
					{Type: AnnouncementConditionTypeBalance, Operator: AnnouncementOperatorLT, Value: 5},
				},
			},
		},
	}

	// 命中第 2 组（balance < 5）
	require.True(t, targeting.Matches(4.99))
	require.False(t, targeting.Matches(5))

	// 命中第 1 组（balance >= 100）
	require.True(t, targeting.Matches(100))
	require.False(t, targeting.Matches(99.9))
}
