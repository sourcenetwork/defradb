// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Iterator is an autogenerated mock type for the Iterator type
type Iterator struct {
	mock.Mock
}

type Iterator_Expecter struct {
	mock *mock.Mock
}

func (_m *Iterator) EXPECT() *Iterator_Expecter {
	return &Iterator_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields: ctx
func (_m *Iterator) Close(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Iterator_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Iterator_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Iterator_Expecter) Close(ctx interface{}) *Iterator_Close_Call {
	return &Iterator_Close_Call{Call: _e.mock.On("Close", ctx)}
}

func (_c *Iterator_Close_Call) Run(run func(ctx context.Context)) *Iterator_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Iterator_Close_Call) Return(_a0 error) *Iterator_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Iterator_Close_Call) RunAndReturn(run func(context.Context) error) *Iterator_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Domain provides a mock function with given fields:
func (_m *Iterator) Domain() ([]byte, []byte) {
	ret := _m.Called()

	var r0 []byte
	var r1 []byte
	if rf, ok := ret.Get(0).(func() ([]byte, []byte)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func() []byte); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	return r0, r1
}

// Iterator_Domain_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Domain'
type Iterator_Domain_Call struct {
	*mock.Call
}

// Domain is a helper method to define mock.On call
func (_e *Iterator_Expecter) Domain() *Iterator_Domain_Call {
	return &Iterator_Domain_Call{Call: _e.mock.On("Domain")}
}

func (_c *Iterator_Domain_Call) Run(run func()) *Iterator_Domain_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Iterator_Domain_Call) Return(start []byte, end []byte) *Iterator_Domain_Call {
	_c.Call.Return(start, end)
	return _c
}

func (_c *Iterator_Domain_Call) RunAndReturn(run func() ([]byte, []byte)) *Iterator_Domain_Call {
	_c.Call.Return(run)
	return _c
}

// Key provides a mock function with given fields:
func (_m *Iterator) Key() []byte {
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

// Iterator_Key_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Key'
type Iterator_Key_Call struct {
	*mock.Call
}

// Key is a helper method to define mock.On call
func (_e *Iterator_Expecter) Key() *Iterator_Key_Call {
	return &Iterator_Key_Call{Call: _e.mock.On("Key")}
}

func (_c *Iterator_Key_Call) Run(run func()) *Iterator_Key_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Iterator_Key_Call) Return(_a0 []byte) *Iterator_Key_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Iterator_Key_Call) RunAndReturn(run func() []byte) *Iterator_Key_Call {
	_c.Call.Return(run)
	return _c
}

// Next provides a mock function with given fields:
func (_m *Iterator) Next() {
	_m.Called()
}

// Iterator_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type Iterator_Next_Call struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
func (_e *Iterator_Expecter) Next() *Iterator_Next_Call {
	return &Iterator_Next_Call{Call: _e.mock.On("Next")}
}

func (_c *Iterator_Next_Call) Run(run func()) *Iterator_Next_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Iterator_Next_Call) Return() *Iterator_Next_Call {
	_c.Call.Return()
	return _c
}

func (_c *Iterator_Next_Call) RunAndReturn(run func()) *Iterator_Next_Call {
	_c.Call.Return(run)
	return _c
}

// Seek provides a mock function with given fields: _a0
func (_m *Iterator) Seek(_a0 []byte) {
	_m.Called(_a0)
}

// Iterator_Seek_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Seek'
type Iterator_Seek_Call struct {
	*mock.Call
}

// Seek is a helper method to define mock.On call
//   - _a0 []byte
func (_e *Iterator_Expecter) Seek(_a0 interface{}) *Iterator_Seek_Call {
	return &Iterator_Seek_Call{Call: _e.mock.On("Seek", _a0)}
}

func (_c *Iterator_Seek_Call) Run(run func(_a0 []byte)) *Iterator_Seek_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte))
	})
	return _c
}

func (_c *Iterator_Seek_Call) Return() *Iterator_Seek_Call {
	_c.Call.Return()
	return _c
}

func (_c *Iterator_Seek_Call) RunAndReturn(run func([]byte)) *Iterator_Seek_Call {
	_c.Call.Return(run)
	return _c
}

// Valid provides a mock function with given fields:
func (_m *Iterator) Valid() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Iterator_Valid_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Valid'
type Iterator_Valid_Call struct {
	*mock.Call
}

// Valid is a helper method to define mock.On call
func (_e *Iterator_Expecter) Valid() *Iterator_Valid_Call {
	return &Iterator_Valid_Call{Call: _e.mock.On("Valid")}
}

func (_c *Iterator_Valid_Call) Run(run func()) *Iterator_Valid_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Iterator_Valid_Call) Return(_a0 bool) *Iterator_Valid_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Iterator_Valid_Call) RunAndReturn(run func() bool) *Iterator_Valid_Call {
	_c.Call.Return(run)
	return _c
}

// Value provides a mock function with given fields:
func (_m *Iterator) Value() []byte {
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

// Iterator_Value_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Value'
type Iterator_Value_Call struct {
	*mock.Call
}

// Value is a helper method to define mock.On call
func (_e *Iterator_Expecter) Value() *Iterator_Value_Call {
	return &Iterator_Value_Call{Call: _e.mock.On("Value")}
}

func (_c *Iterator_Value_Call) Run(run func()) *Iterator_Value_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Iterator_Value_Call) Return(_a0 []byte) *Iterator_Value_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Iterator_Value_Call) RunAndReturn(run func() []byte) *Iterator_Value_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewIterator interface {
	mock.TestingT
	Cleanup(func())
}

// NewIterator creates a new instance of Iterator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIterator(t mockConstructorTestingTNewIterator) *Iterator {
	mock := &Iterator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
