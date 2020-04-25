// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import flow "github.com/dapperlabs/flow-go/model/flow"
import mock "github.com/stretchr/testify/mock"

// TransactionErrors is an autogenerated mock type for the TransactionErrors type
type TransactionErrors struct {
	mock.Mock
}

// ByBlockIDTransactionID provides a mock function with given fields: blockID, transactionID
func (_m *TransactionErrors) ByBlockIDTransactionID(blockID flow.Identifier, transactionID flow.Identifier) (*flow.TransactionError, error) {
	ret := _m.Called(blockID, transactionID)

	var r0 *flow.TransactionError
	if rf, ok := ret.Get(0).(func(flow.Identifier, flow.Identifier) *flow.TransactionError); ok {
		r0 = rf(blockID, transactionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.TransactionError)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier, flow.Identifier) error); ok {
		r1 = rf(blockID, transactionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Store provides a mock function with given fields: blockID, transactionError
func (_m *TransactionErrors) Store(blockID flow.Identifier, transactionError *flow.TransactionError) error {
	ret := _m.Called(blockID, transactionError)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Identifier, *flow.TransactionError) error); ok {
		r0 = rf(blockID, transactionError)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}