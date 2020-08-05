// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	epoch "github.com/dapperlabs/flow-go/model/epoch"
	badger "github.com/dgraph-io/badger/v2"

	mock "github.com/stretchr/testify/mock"
)

// EpochCommits is an autogenerated mock type for the EpochCommits type
type EpochCommits struct {
	mock.Mock
}

// ByCounter provides a mock function with given fields: counter
func (_m *EpochCommits) ByCounter(counter uint64) (*epoch.Commit, error) {
	ret := _m.Called(counter)

	var r0 *epoch.Commit
	if rf, ok := ret.Get(0).(func(uint64) *epoch.Commit); ok {
		r0 = rf(counter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*epoch.Commit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(counter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StoreTx provides a mock function with given fields: commit
func (_m *EpochCommits) StoreTx(commit *epoch.Commit) func(*badger.Txn) error {
	ret := _m.Called(commit)

	var r0 func(*badger.Txn) error
	if rf, ok := ret.Get(0).(func(*epoch.Commit) func(*badger.Txn) error); ok {
		r0 = rf(commit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func(*badger.Txn) error)
		}
	}

	return r0
}