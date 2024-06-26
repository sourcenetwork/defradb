// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	datastore "github.com/ipfs/go-datastore"

	mock "github.com/stretchr/testify/mock"

	query "github.com/ipfs/go-datastore/query"
)

// Rootstore is an autogenerated mock type for the Rootstore type
type Rootstore struct {
	mock.Mock
}

type Rootstore_Expecter struct {
	mock *mock.Mock
}

func (_m *Rootstore) EXPECT() *Rootstore_Expecter {
	return &Rootstore_Expecter{mock: &_m.Mock}
}

// Batch provides a mock function with given fields: ctx
func (_m *Rootstore) Batch(ctx context.Context) (datastore.Batch, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Batch")
	}

	var r0 datastore.Batch
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (datastore.Batch, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) datastore.Batch); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.Batch)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rootstore_Batch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Batch'
type Rootstore_Batch_Call struct {
	*mock.Call
}

// Batch is a helper method to define mock.On call
//   - ctx context.Context
func (_e *Rootstore_Expecter) Batch(ctx interface{}) *Rootstore_Batch_Call {
	return &Rootstore_Batch_Call{Call: _e.mock.On("Batch", ctx)}
}

func (_c *Rootstore_Batch_Call) Run(run func(ctx context.Context)) *Rootstore_Batch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Rootstore_Batch_Call) Return(_a0 datastore.Batch, _a1 error) *Rootstore_Batch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Rootstore_Batch_Call) RunAndReturn(run func(context.Context) (datastore.Batch, error)) *Rootstore_Batch_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with given fields:
func (_m *Rootstore) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rootstore_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Rootstore_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Rootstore_Expecter) Close() *Rootstore_Close_Call {
	return &Rootstore_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Rootstore_Close_Call) Run(run func()) *Rootstore_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Rootstore_Close_Call) Return(_a0 error) *Rootstore_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Rootstore_Close_Call) RunAndReturn(run func() error) *Rootstore_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, key
func (_m *Rootstore) Delete(ctx context.Context, key datastore.Key) error {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rootstore_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Rootstore_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *Rootstore_Expecter) Delete(ctx interface{}, key interface{}) *Rootstore_Delete_Call {
	return &Rootstore_Delete_Call{Call: _e.mock.On("Delete", ctx, key)}
}

func (_c *Rootstore_Delete_Call) Run(run func(ctx context.Context, key datastore.Key)) *Rootstore_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *Rootstore_Delete_Call) Return(_a0 error) *Rootstore_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Rootstore_Delete_Call) RunAndReturn(run func(context.Context, datastore.Key) error) *Rootstore_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *Rootstore) Get(ctx context.Context, key datastore.Key) ([]byte, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) ([]byte, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) []byte); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, datastore.Key) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rootstore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Rootstore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *Rootstore_Expecter) Get(ctx interface{}, key interface{}) *Rootstore_Get_Call {
	return &Rootstore_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *Rootstore_Get_Call) Run(run func(ctx context.Context, key datastore.Key)) *Rootstore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *Rootstore_Get_Call) Return(value []byte, err error) *Rootstore_Get_Call {
	_c.Call.Return(value, err)
	return _c
}

func (_c *Rootstore_Get_Call) RunAndReturn(run func(context.Context, datastore.Key) ([]byte, error)) *Rootstore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetSize provides a mock function with given fields: ctx, key
func (_m *Rootstore) GetSize(ctx context.Context, key datastore.Key) (int, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for GetSize")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) (int, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) int); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, datastore.Key) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rootstore_GetSize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSize'
type Rootstore_GetSize_Call struct {
	*mock.Call
}

// GetSize is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *Rootstore_Expecter) GetSize(ctx interface{}, key interface{}) *Rootstore_GetSize_Call {
	return &Rootstore_GetSize_Call{Call: _e.mock.On("GetSize", ctx, key)}
}

func (_c *Rootstore_GetSize_Call) Run(run func(ctx context.Context, key datastore.Key)) *Rootstore_GetSize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *Rootstore_GetSize_Call) Return(size int, err error) *Rootstore_GetSize_Call {
	_c.Call.Return(size, err)
	return _c
}

func (_c *Rootstore_GetSize_Call) RunAndReturn(run func(context.Context, datastore.Key) (int, error)) *Rootstore_GetSize_Call {
	_c.Call.Return(run)
	return _c
}

// Has provides a mock function with given fields: ctx, key
func (_m *Rootstore) Has(ctx context.Context, key datastore.Key) (bool, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Has")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) (bool, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) bool); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, datastore.Key) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rootstore_Has_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Has'
type Rootstore_Has_Call struct {
	*mock.Call
}

// Has is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *Rootstore_Expecter) Has(ctx interface{}, key interface{}) *Rootstore_Has_Call {
	return &Rootstore_Has_Call{Call: _e.mock.On("Has", ctx, key)}
}

func (_c *Rootstore_Has_Call) Run(run func(ctx context.Context, key datastore.Key)) *Rootstore_Has_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *Rootstore_Has_Call) Return(exists bool, err error) *Rootstore_Has_Call {
	_c.Call.Return(exists, err)
	return _c
}

func (_c *Rootstore_Has_Call) RunAndReturn(run func(context.Context, datastore.Key) (bool, error)) *Rootstore_Has_Call {
	_c.Call.Return(run)
	return _c
}

// NewTransaction provides a mock function with given fields: ctx, readOnly
func (_m *Rootstore) NewTransaction(ctx context.Context, readOnly bool) (datastore.Txn, error) {
	ret := _m.Called(ctx, readOnly)

	if len(ret) == 0 {
		panic("no return value specified for NewTransaction")
	}

	var r0 datastore.Txn
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, bool) (datastore.Txn, error)); ok {
		return rf(ctx, readOnly)
	}
	if rf, ok := ret.Get(0).(func(context.Context, bool) datastore.Txn); ok {
		r0 = rf(ctx, readOnly)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(datastore.Txn)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, bool) error); ok {
		r1 = rf(ctx, readOnly)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rootstore_NewTransaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewTransaction'
type Rootstore_NewTransaction_Call struct {
	*mock.Call
}

// NewTransaction is a helper method to define mock.On call
//   - ctx context.Context
//   - readOnly bool
func (_e *Rootstore_Expecter) NewTransaction(ctx interface{}, readOnly interface{}) *Rootstore_NewTransaction_Call {
	return &Rootstore_NewTransaction_Call{Call: _e.mock.On("NewTransaction", ctx, readOnly)}
}

func (_c *Rootstore_NewTransaction_Call) Run(run func(ctx context.Context, readOnly bool)) *Rootstore_NewTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(bool))
	})
	return _c
}

func (_c *Rootstore_NewTransaction_Call) Return(_a0 datastore.Txn, _a1 error) *Rootstore_NewTransaction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Rootstore_NewTransaction_Call) RunAndReturn(run func(context.Context, bool) (datastore.Txn, error)) *Rootstore_NewTransaction_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, key, value
func (_m *Rootstore) Put(ctx context.Context, key datastore.Key, value []byte) error {
	ret := _m.Called(ctx, key, value)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key, []byte) error); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rootstore_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type Rootstore_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
//   - value []byte
func (_e *Rootstore_Expecter) Put(ctx interface{}, key interface{}, value interface{}) *Rootstore_Put_Call {
	return &Rootstore_Put_Call{Call: _e.mock.On("Put", ctx, key, value)}
}

func (_c *Rootstore_Put_Call) Run(run func(ctx context.Context, key datastore.Key, value []byte)) *Rootstore_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key), args[2].([]byte))
	})
	return _c
}

func (_c *Rootstore_Put_Call) Return(_a0 error) *Rootstore_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Rootstore_Put_Call) RunAndReturn(run func(context.Context, datastore.Key, []byte) error) *Rootstore_Put_Call {
	_c.Call.Return(run)
	return _c
}

// Query provides a mock function with given fields: ctx, q
func (_m *Rootstore) Query(ctx context.Context, q query.Query) (query.Results, error) {
	ret := _m.Called(ctx, q)

	if len(ret) == 0 {
		panic("no return value specified for Query")
	}

	var r0 query.Results
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, query.Query) (query.Results, error)); ok {
		return rf(ctx, q)
	}
	if rf, ok := ret.Get(0).(func(context.Context, query.Query) query.Results); ok {
		r0 = rf(ctx, q)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(query.Results)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, query.Query) error); ok {
		r1 = rf(ctx, q)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rootstore_Query_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Query'
type Rootstore_Query_Call struct {
	*mock.Call
}

// Query is a helper method to define mock.On call
//   - ctx context.Context
//   - q query.Query
func (_e *Rootstore_Expecter) Query(ctx interface{}, q interface{}) *Rootstore_Query_Call {
	return &Rootstore_Query_Call{Call: _e.mock.On("Query", ctx, q)}
}

func (_c *Rootstore_Query_Call) Run(run func(ctx context.Context, q query.Query)) *Rootstore_Query_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(query.Query))
	})
	return _c
}

func (_c *Rootstore_Query_Call) Return(_a0 query.Results, _a1 error) *Rootstore_Query_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Rootstore_Query_Call) RunAndReturn(run func(context.Context, query.Query) (query.Results, error)) *Rootstore_Query_Call {
	_c.Call.Return(run)
	return _c
}

// Sync provides a mock function with given fields: ctx, prefix
func (_m *Rootstore) Sync(ctx context.Context, prefix datastore.Key) error {
	ret := _m.Called(ctx, prefix)

	if len(ret) == 0 {
		panic("no return value specified for Sync")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) error); ok {
		r0 = rf(ctx, prefix)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Rootstore_Sync_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Sync'
type Rootstore_Sync_Call struct {
	*mock.Call
}

// Sync is a helper method to define mock.On call
//   - ctx context.Context
//   - prefix datastore.Key
func (_e *Rootstore_Expecter) Sync(ctx interface{}, prefix interface{}) *Rootstore_Sync_Call {
	return &Rootstore_Sync_Call{Call: _e.mock.On("Sync", ctx, prefix)}
}

func (_c *Rootstore_Sync_Call) Run(run func(ctx context.Context, prefix datastore.Key)) *Rootstore_Sync_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *Rootstore_Sync_Call) Return(_a0 error) *Rootstore_Sync_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Rootstore_Sync_Call) RunAndReturn(run func(context.Context, datastore.Key) error) *Rootstore_Sync_Call {
	_c.Call.Return(run)
	return _c
}

// NewRootstore creates a new instance of Rootstore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRootstore(t interface {
	mock.TestingT
	Cleanup(func())
}) *Rootstore {
	mock := &Rootstore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
