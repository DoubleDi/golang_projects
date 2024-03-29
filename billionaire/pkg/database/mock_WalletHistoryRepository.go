// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package database

import (
	context "context"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// MockWalletHistoryRepository is an autogenerated mock type for the WalletHistoryRepository type
type MockWalletHistoryRepository struct {
	mock.Mock
}

// AddBalance provides a mock function with given fields: ctx, b
func (_m *MockWalletHistoryRepository) AddBalance(ctx context.Context, b *Balance) error {
	ret := _m.Called(ctx, b)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *Balance) error); ok {
		r0 = rf(ctx, b)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBalances provides a mock function with given fields: ctx, from, to
func (_m *MockWalletHistoryRepository) GetBalances(ctx context.Context, from time.Time, to time.Time) ([]Balance, error) {
	ret := _m.Called(ctx, from, to)

	var r0 []Balance
	if rf, ok := ret.Get(0).(func(context.Context, time.Time, time.Time) []Balance); ok {
		r0 = rf(ctx, from, to)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Balance)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, time.Time, time.Time) error); ok {
		r1 = rf(ctx, from, to)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
