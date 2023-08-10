// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	datastore "github.com/sourcenetwork/defradb/datastore"
	mock "github.com/stretchr/testify/mock"
)

// Txn is an autogenerated mock type for the Txn type
type Txn struct {
	mock.Mock
}

type Txn_Expecter struct {
	mock *mock.Mock
}

func (_m *Txn) EXPECT() *Txn_Expecter {
	return &Txn_Expecter{mock: &_m.Mock}
}

// Commit provides a mock function with given fields: ctx
func (_m *Txn) Commit(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Txn_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type Txn_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Txn_Expecter) Commit(ctx interface{}) *Txn_Commit_Call {
	return &Txn_Commit_Call{Call: _e.mock.On("Commit", ctx)}
}

func (_c *Txn_Commit_Call) Run(run func(ctx context.Context)) *Txn_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Txn_Commit_Call) Return(_a0 error) *Txn_Commit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_Commit_Call) RunAndReturn(run func(context.Context) error) *Txn_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// DAGstore provides a mock function with given fields:
func (_m *Txn) DAGstore() datastore.DAGStore {
	ret := _m.Called()

	var r0 datastore.DAGStore
	if rf, ok := ret.Get(0).(func() datastore.DAGStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.DAGStore)
		}
	}

	return r0
}

// Txn_DAGstore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DAGstore'
type Txn_DAGstore_Call struct {
	*mock.Call
}

// DAGstore is a helper method to define mock.On call
func (_e *Txn_Expecter) DAGstore() *Txn_DAGstore_Call {
	return &Txn_DAGstore_Call{Call: _e.mock.On("DAGstore")}
}

func (_c *Txn_DAGstore_Call) Run(run func()) *Txn_DAGstore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txn_DAGstore_Call) Return(_a0 datastore.DAGStore) *Txn_DAGstore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_DAGstore_Call) RunAndReturn(run func() datastore.DAGStore) *Txn_DAGstore_Call {
	_c.Call.Return(run)
	return _c
}

// Datastore provides a mock function with given fields:
func (_m *Txn) Datastore() datastore.DSReaderWriter {
	ret := _m.Called()

	var r0 datastore.DSReaderWriter
	if rf, ok := ret.Get(0).(func() datastore.DSReaderWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.DSReaderWriter)
		}
	}

	return r0
}

// Txn_Datastore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Datastore'
type Txn_Datastore_Call struct {
	*mock.Call
}

// Datastore is a helper method to define mock.On call
func (_e *Txn_Expecter) Datastore() *Txn_Datastore_Call {
	return &Txn_Datastore_Call{Call: _e.mock.On("Datastore")}
}

func (_c *Txn_Datastore_Call) Run(run func()) *Txn_Datastore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txn_Datastore_Call) Return(_a0 datastore.DSReaderWriter) *Txn_Datastore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_Datastore_Call) RunAndReturn(run func() datastore.DSReaderWriter) *Txn_Datastore_Call {
	_c.Call.Return(run)
	return _c
}

// Discard provides a mock function with given fields: ctx
func (_m *Txn) Discard(ctx context.Context) {
	_m.Called(ctx)
}

// Txn_Discard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Discard'
type Txn_Discard_Call struct {
	*mock.Call
}

// Discard is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Txn_Expecter) Discard(ctx interface{}) *Txn_Discard_Call {
	return &Txn_Discard_Call{Call: _e.mock.On("Discard", ctx)}
}

func (_c *Txn_Discard_Call) Run(run func(ctx context.Context)) *Txn_Discard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Txn_Discard_Call) Return() *Txn_Discard_Call {
	_c.Call.Return()
	return _c
}

func (_c *Txn_Discard_Call) RunAndReturn(run func(context.Context)) *Txn_Discard_Call {
	_c.Call.Return(run)
	return _c
}

// Headstore provides a mock function with given fields:
func (_m *Txn) Headstore() datastore.DSReaderWriter {
	ret := _m.Called()

	var r0 datastore.DSReaderWriter
	if rf, ok := ret.Get(0).(func() datastore.DSReaderWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.DSReaderWriter)
		}
	}

	return r0
}

// Txn_Headstore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Headstore'
type Txn_Headstore_Call struct {
	*mock.Call
}

// Headstore is a helper method to define mock.On call
func (_e *Txn_Expecter) Headstore() *Txn_Headstore_Call {
	return &Txn_Headstore_Call{Call: _e.mock.On("Headstore")}
}

func (_c *Txn_Headstore_Call) Run(run func()) *Txn_Headstore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txn_Headstore_Call) Return(_a0 datastore.DSReaderWriter) *Txn_Headstore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_Headstore_Call) RunAndReturn(run func() datastore.DSReaderWriter) *Txn_Headstore_Call {
	_c.Call.Return(run)
	return _c
}

// ID provides a mock function with given fields:
func (_m *Txn) ID() uint64 {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// Txn_ID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ID'
type Txn_ID_Call struct {
	*mock.Call
}

// ID is a helper method to define mock.On call
func (_e *Txn_Expecter) ID() *Txn_ID_Call {
	return &Txn_ID_Call{Call: _e.mock.On("ID")}
}

func (_c *Txn_ID_Call) Run(run func()) *Txn_ID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txn_ID_Call) Return(_a0 uint64) *Txn_ID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_ID_Call) RunAndReturn(run func() uint64) *Txn_ID_Call {
	_c.Call.Return(run)
	return _c
}

// OnDiscard provides a mock function with given fields: fn
func (_m *Txn) OnDiscard(fn func()) {
	_m.Called(fn)
}

// Txn_OnDiscard_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnDiscard'
type Txn_OnDiscard_Call struct {
	*mock.Call
}

// OnDiscard is a helper method to define mock.On call
//   - fn func()
func (_e *Txn_Expecter) OnDiscard(fn interface{}) *Txn_OnDiscard_Call {
	return &Txn_OnDiscard_Call{Call: _e.mock.On("OnDiscard", fn)}
}

func (_c *Txn_OnDiscard_Call) Run(run func(fn func())) *Txn_OnDiscard_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(func()))
	})
	return _c
}

func (_c *Txn_OnDiscard_Call) Return() *Txn_OnDiscard_Call {
	_c.Call.Return()
	return _c
}

func (_c *Txn_OnDiscard_Call) RunAndReturn(run func(func())) *Txn_OnDiscard_Call {
	_c.Call.Return(run)
	return _c
}

// OnError provides a mock function with given fields: fn
func (_m *Txn) OnError(fn func()) {
	_m.Called(fn)
}

// Txn_OnError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnError'
type Txn_OnError_Call struct {
	*mock.Call
}

// OnError is a helper method to define mock.On call
//   - fn func()
func (_e *Txn_Expecter) OnError(fn interface{}) *Txn_OnError_Call {
	return &Txn_OnError_Call{Call: _e.mock.On("OnError", fn)}
}

func (_c *Txn_OnError_Call) Run(run func(fn func())) *Txn_OnError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(func()))
	})
	return _c
}

func (_c *Txn_OnError_Call) Return() *Txn_OnError_Call {
	_c.Call.Return()
	return _c
}

func (_c *Txn_OnError_Call) RunAndReturn(run func(func())) *Txn_OnError_Call {
	_c.Call.Return(run)
	return _c
}

// OnSuccess provides a mock function with given fields: fn
func (_m *Txn) OnSuccess(fn func()) {
	_m.Called(fn)
}

// Txn_OnSuccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OnSuccess'
type Txn_OnSuccess_Call struct {
	*mock.Call
}

// OnSuccess is a helper method to define mock.On call
//   - fn func()
func (_e *Txn_Expecter) OnSuccess(fn interface{}) *Txn_OnSuccess_Call {
	return &Txn_OnSuccess_Call{Call: _e.mock.On("OnSuccess", fn)}
}

func (_c *Txn_OnSuccess_Call) Run(run func(fn func())) *Txn_OnSuccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(func()))
	})
	return _c
}

func (_c *Txn_OnSuccess_Call) Return() *Txn_OnSuccess_Call {
	_c.Call.Return()
	return _c
}

func (_c *Txn_OnSuccess_Call) RunAndReturn(run func(func())) *Txn_OnSuccess_Call {
	_c.Call.Return(run)
	return _c
}

// Rootstore provides a mock function with given fields:
func (_m *Txn) Rootstore() datastore.DSReaderWriter {
	ret := _m.Called()

	var r0 datastore.DSReaderWriter
	if rf, ok := ret.Get(0).(func() datastore.DSReaderWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.DSReaderWriter)
		}
	}

	return r0
}

// Txn_Rootstore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rootstore'
type Txn_Rootstore_Call struct {
	*mock.Call
}

// Rootstore is a helper method to define mock.On call
func (_e *Txn_Expecter) Rootstore() *Txn_Rootstore_Call {
	return &Txn_Rootstore_Call{Call: _e.mock.On("Rootstore")}
}

func (_c *Txn_Rootstore_Call) Run(run func()) *Txn_Rootstore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txn_Rootstore_Call) Return(_a0 datastore.DSReaderWriter) *Txn_Rootstore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_Rootstore_Call) RunAndReturn(run func() datastore.DSReaderWriter) *Txn_Rootstore_Call {
	_c.Call.Return(run)
	return _c
}

// Systemstore provides a mock function with given fields:
func (_m *Txn) Systemstore() datastore.DSReaderWriter {
	ret := _m.Called()

	var r0 datastore.DSReaderWriter
	if rf, ok := ret.Get(0).(func() datastore.DSReaderWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.DSReaderWriter)
		}
	}

	return r0
}

// Txn_Systemstore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Systemstore'
type Txn_Systemstore_Call struct {
	*mock.Call
}

// Systemstore is a helper method to define mock.On call
func (_e *Txn_Expecter) Systemstore() *Txn_Systemstore_Call {
	return &Txn_Systemstore_Call{Call: _e.mock.On("Systemstore")}
}

func (_c *Txn_Systemstore_Call) Run(run func()) *Txn_Systemstore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Txn_Systemstore_Call) Return(_a0 datastore.DSReaderWriter) *Txn_Systemstore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Txn_Systemstore_Call) RunAndReturn(run func() datastore.DSReaderWriter) *Txn_Systemstore_Call {
	_c.Call.Return(run)
	return _c
}

// NewTxn creates a new instance of Txn. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTxn(t interface {
	mock.TestingT
	Cleanup(func())
}) *Txn {
	mock := &Txn{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
