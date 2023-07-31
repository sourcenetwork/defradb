// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	datastore "github.com/ipfs/go-datastore"

	iterable "github.com/sourcenetwork/defradb/datastore/iterable"

	mock "github.com/stretchr/testify/mock"

	query "github.com/ipfs/go-datastore/query"
)

// DSReaderWriter is an autogenerated mock type for the DSReaderWriter type
type DSReaderWriter struct {
	mock.Mock
}

type DSReaderWriter_Expecter struct {
	mock *mock.Mock
}

func (_m *DSReaderWriter) EXPECT() *DSReaderWriter_Expecter {
	return &DSReaderWriter_Expecter{mock: &_m.Mock}
}

// Delete provides a mock function with given fields: ctx, key
func (_m *DSReaderWriter) Delete(ctx context.Context, key datastore.Key) error {
	ret := _m.Called(ctx, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DSReaderWriter_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type DSReaderWriter_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *DSReaderWriter_Expecter) Delete(ctx interface{}, key interface{}) *DSReaderWriter_Delete_Call {
	return &DSReaderWriter_Delete_Call{Call: _e.mock.On("Delete", ctx, key)}
}

func (_c *DSReaderWriter_Delete_Call) Run(run func(ctx context.Context, key datastore.Key)) *DSReaderWriter_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *DSReaderWriter_Delete_Call) Return(_a0 error) *DSReaderWriter_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DSReaderWriter_Delete_Call) RunAndReturn(run func(context.Context, datastore.Key) error) *DSReaderWriter_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, key
func (_m *DSReaderWriter) Get(ctx context.Context, key datastore.Key) ([]byte, error) {
	ret := _m.Called(ctx, key)

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

// DSReaderWriter_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type DSReaderWriter_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *DSReaderWriter_Expecter) Get(ctx interface{}, key interface{}) *DSReaderWriter_Get_Call {
	return &DSReaderWriter_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *DSReaderWriter_Get_Call) Run(run func(ctx context.Context, key datastore.Key)) *DSReaderWriter_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *DSReaderWriter_Get_Call) Return(value []byte, err error) *DSReaderWriter_Get_Call {
	_c.Call.Return(value, err)
	return _c
}

func (_c *DSReaderWriter_Get_Call) RunAndReturn(run func(context.Context, datastore.Key) ([]byte, error)) *DSReaderWriter_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetIterator provides a mock function with given fields: q
func (_m *DSReaderWriter) GetIterator(q query.Query) (iterable.Iterator, error) {
	ret := _m.Called(q)

	var r0 iterable.Iterator
	var r1 error
	if rf, ok := ret.Get(0).(func(query.Query) (iterable.Iterator, error)); ok {
		return rf(q)
	}
	if rf, ok := ret.Get(0).(func(query.Query) iterable.Iterator); ok {
		r0 = rf(q)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(iterable.Iterator)
		}
	}

	if rf, ok := ret.Get(1).(func(query.Query) error); ok {
		r1 = rf(q)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DSReaderWriter_GetIterator_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIterator'
type DSReaderWriter_GetIterator_Call struct {
	*mock.Call
}

// GetIterator is a helper method to define mock.On call
//   - q query.Query
func (_e *DSReaderWriter_Expecter) GetIterator(q interface{}) *DSReaderWriter_GetIterator_Call {
	return &DSReaderWriter_GetIterator_Call{Call: _e.mock.On("GetIterator", q)}
}

func (_c *DSReaderWriter_GetIterator_Call) Run(run func(q query.Query)) *DSReaderWriter_GetIterator_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(query.Query))
	})
	return _c
}

func (_c *DSReaderWriter_GetIterator_Call) Return(_a0 iterable.Iterator, _a1 error) *DSReaderWriter_GetIterator_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DSReaderWriter_GetIterator_Call) RunAndReturn(run func(query.Query) (iterable.Iterator, error)) *DSReaderWriter_GetIterator_Call {
	_c.Call.Return(run)
	return _c
}

// GetSize provides a mock function with given fields: ctx, key
func (_m *DSReaderWriter) GetSize(ctx context.Context, key datastore.Key) (int, error) {
	ret := _m.Called(ctx, key)

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

// DSReaderWriter_GetSize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSize'
type DSReaderWriter_GetSize_Call struct {
	*mock.Call
}

// GetSize is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *DSReaderWriter_Expecter) GetSize(ctx interface{}, key interface{}) *DSReaderWriter_GetSize_Call {
	return &DSReaderWriter_GetSize_Call{Call: _e.mock.On("GetSize", ctx, key)}
}

func (_c *DSReaderWriter_GetSize_Call) Run(run func(ctx context.Context, key datastore.Key)) *DSReaderWriter_GetSize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *DSReaderWriter_GetSize_Call) Return(size int, err error) *DSReaderWriter_GetSize_Call {
	_c.Call.Return(size, err)
	return _c
}

func (_c *DSReaderWriter_GetSize_Call) RunAndReturn(run func(context.Context, datastore.Key) (int, error)) *DSReaderWriter_GetSize_Call {
	_c.Call.Return(run)
	return _c
}

// Has provides a mock function with given fields: ctx, key
func (_m *DSReaderWriter) Has(ctx context.Context, key datastore.Key) (bool, error) {
	ret := _m.Called(ctx, key)

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

// DSReaderWriter_Has_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Has'
type DSReaderWriter_Has_Call struct {
	*mock.Call
}

// Has is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
func (_e *DSReaderWriter_Expecter) Has(ctx interface{}, key interface{}) *DSReaderWriter_Has_Call {
	return &DSReaderWriter_Has_Call{Call: _e.mock.On("Has", ctx, key)}
}

func (_c *DSReaderWriter_Has_Call) Run(run func(ctx context.Context, key datastore.Key)) *DSReaderWriter_Has_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key))
	})
	return _c
}

func (_c *DSReaderWriter_Has_Call) Return(exists bool, err error) *DSReaderWriter_Has_Call {
	_c.Call.Return(exists, err)
	return _c
}

func (_c *DSReaderWriter_Has_Call) RunAndReturn(run func(context.Context, datastore.Key) (bool, error)) *DSReaderWriter_Has_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: ctx, key, value
func (_m *DSReaderWriter) Put(ctx context.Context, key datastore.Key, value []byte) error {
	ret := _m.Called(ctx, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, datastore.Key, []byte) error); ok {
		r0 = rf(ctx, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DSReaderWriter_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type DSReaderWriter_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - ctx context.Context
//   - key datastore.Key
//   - value []byte
func (_e *DSReaderWriter_Expecter) Put(ctx interface{}, key interface{}, value interface{}) *DSReaderWriter_Put_Call {
	return &DSReaderWriter_Put_Call{Call: _e.mock.On("Put", ctx, key, value)}
}

func (_c *DSReaderWriter_Put_Call) Run(run func(ctx context.Context, key datastore.Key, value []byte)) *DSReaderWriter_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(datastore.Key), args[2].([]byte))
	})
	return _c
}

func (_c *DSReaderWriter_Put_Call) Return(_a0 error) *DSReaderWriter_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DSReaderWriter_Put_Call) RunAndReturn(run func(context.Context, datastore.Key, []byte) error) *DSReaderWriter_Put_Call {
	_c.Call.Return(run)
	return _c
}

// Query provides a mock function with given fields: ctx, q
func (_m *DSReaderWriter) Query(ctx context.Context, q query.Query) (query.Results, error) {
	ret := _m.Called(ctx, q)

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

// DSReaderWriter_Query_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Query'
type DSReaderWriter_Query_Call struct {
	*mock.Call
}

// Query is a helper method to define mock.On call
//   - ctx context.Context
//   - q query.Query
func (_e *DSReaderWriter_Expecter) Query(ctx interface{}, q interface{}) *DSReaderWriter_Query_Call {
	return &DSReaderWriter_Query_Call{Call: _e.mock.On("Query", ctx, q)}
}

func (_c *DSReaderWriter_Query_Call) Run(run func(ctx context.Context, q query.Query)) *DSReaderWriter_Query_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(query.Query))
	})
	return _c
}

func (_c *DSReaderWriter_Query_Call) Return(_a0 query.Results, _a1 error) *DSReaderWriter_Query_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DSReaderWriter_Query_Call) RunAndReturn(run func(context.Context, query.Query) (query.Results, error)) *DSReaderWriter_Query_Call {
	_c.Call.Return(run)
	return _c
}

// NewDSReaderWriter creates a new instance of DSReaderWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDSReaderWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *DSReaderWriter {
	mock := &DSReaderWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
