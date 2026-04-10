//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type userRepoStubForListUsers struct {
	userRepoStub
	users []User
	err   error
}

func (s *userRepoStubForListUsers) ListWithFilters(_ context.Context, params pagination.PaginationParams, _ UserListFilters) ([]User, *pagination.PaginationResult, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	out := make([]User, len(s.users))
	copy(out, s.users)
	return out, &pagination.PaginationResult{
		Total:    int64(len(out)),
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

func TestAdminService_ListUsers_ReturnsUsers(t *testing.T) {
	userRepo := &userRepoStubForListUsers{
		users: []User{
			{ID: 101, Username: "u1"},
			{ID: 202, Username: "u2"},
		},
	}
	svc := &adminServiceImpl{
		userRepo: userRepo,
	}

	users, total, err := svc.ListUsers(context.Background(), 1, 20, UserListFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, users, 2)
}
