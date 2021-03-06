// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	flow "github.com/onflow/flow-go/model/flow"

	mock "github.com/stretchr/testify/mock"
)

// AddressState is an autogenerated mock type for the AddressState type
type AddressState struct {
	mock.Mock
}

// Bytes provides a mock function with given fields:
func (_m *AddressState) Bytes() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// CurrentAddress provides a mock function with given fields:
func (_m *AddressState) CurrentAddress() flow.Address {
	ret := _m.Called()

	var r0 flow.Address
	if rf, ok := ret.Get(0).(func() flow.Address); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.Address)
		}
	}

	return r0
}

// NextAddress provides a mock function with given fields:
func (_m *AddressState) NextAddress() (flow.Address, error) {
	ret := _m.Called()

	var r0 flow.Address
	if rf, ok := ret.Get(0).(func() flow.Address); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(flow.Address)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
