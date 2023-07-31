// Code generated by mockery v2.30.1. DO NOT EDIT.

package mocks

import (
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"

	context "context"

	mock "github.com/stretchr/testify/mock"
)

// DAGStore is an autogenerated mock type for the DAGStore type
type DAGStore struct {
	mock.Mock
}

type DAGStore_Expecter struct {
	mock *mock.Mock
}

func (_m *DAGStore) EXPECT() *DAGStore_Expecter {
	return &DAGStore_Expecter{mock: &_m.Mock}
}

// AllKeysChan provides a mock function with given fields: ctx
func (_m *DAGStore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	ret := _m.Called(ctx)

	var r0 <-chan cid.Cid
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (<-chan cid.Cid, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) <-chan cid.Cid); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan cid.Cid)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DAGStore_AllKeysChan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AllKeysChan'
type DAGStore_AllKeysChan_Call struct {
	*mock.Call
}

// AllKeysChan is a helper method to define mock.On call
//   - ctx context.Context
func (_e *DAGStore_Expecter) AllKeysChan(ctx interface{}) *DAGStore_AllKeysChan_Call {
	return &DAGStore_AllKeysChan_Call{Call: _e.mock.On("AllKeysChan", ctx)}
}

func (_c *DAGStore_AllKeysChan_Call) Run(run func(ctx context.Context)) *DAGStore_AllKeysChan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *DAGStore_AllKeysChan_Call) Return(_a0 <-chan cid.Cid, _a1 error) *DAGStore_AllKeysChan_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DAGStore_AllKeysChan_Call) RunAndReturn(run func(context.Context) (<-chan cid.Cid, error)) *DAGStore_AllKeysChan_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteBlock provides a mock function with given fields: _a0, _a1
func (_m *DAGStore) DeleteBlock(_a0 context.Context, _a1 cid.Cid) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DAGStore_DeleteBlock_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteBlock'
type DAGStore_DeleteBlock_Call struct {
	*mock.Call
}

// DeleteBlock is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 cid.Cid
func (_e *DAGStore_Expecter) DeleteBlock(_a0 interface{}, _a1 interface{}) *DAGStore_DeleteBlock_Call {
	return &DAGStore_DeleteBlock_Call{Call: _e.mock.On("DeleteBlock", _a0, _a1)}
}

func (_c *DAGStore_DeleteBlock_Call) Run(run func(_a0 context.Context, _a1 cid.Cid)) *DAGStore_DeleteBlock_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(cid.Cid))
	})
	return _c
}

func (_c *DAGStore_DeleteBlock_Call) Return(_a0 error) *DAGStore_DeleteBlock_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DAGStore_DeleteBlock_Call) RunAndReturn(run func(context.Context, cid.Cid) error) *DAGStore_DeleteBlock_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: _a0, _a1
func (_m *DAGStore) Get(_a0 context.Context, _a1 cid.Cid) (blocks.Block, error) {
	ret := _m.Called(_a0, _a1)

	var r0 blocks.Block
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) (blocks.Block, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) blocks.Block); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(blocks.Block)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, cid.Cid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DAGStore_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type DAGStore_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 cid.Cid
func (_e *DAGStore_Expecter) Get(_a0 interface{}, _a1 interface{}) *DAGStore_Get_Call {
	return &DAGStore_Get_Call{Call: _e.mock.On("Get", _a0, _a1)}
}

func (_c *DAGStore_Get_Call) Run(run func(_a0 context.Context, _a1 cid.Cid)) *DAGStore_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(cid.Cid))
	})
	return _c
}

func (_c *DAGStore_Get_Call) Return(_a0 blocks.Block, _a1 error) *DAGStore_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DAGStore_Get_Call) RunAndReturn(run func(context.Context, cid.Cid) (blocks.Block, error)) *DAGStore_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetSize provides a mock function with given fields: _a0, _a1
func (_m *DAGStore) GetSize(_a0 context.Context, _a1 cid.Cid) (int, error) {
	ret := _m.Called(_a0, _a1)

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) (int, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) int); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, cid.Cid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DAGStore_GetSize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSize'
type DAGStore_GetSize_Call struct {
	*mock.Call
}

// GetSize is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 cid.Cid
func (_e *DAGStore_Expecter) GetSize(_a0 interface{}, _a1 interface{}) *DAGStore_GetSize_Call {
	return &DAGStore_GetSize_Call{Call: _e.mock.On("GetSize", _a0, _a1)}
}

func (_c *DAGStore_GetSize_Call) Run(run func(_a0 context.Context, _a1 cid.Cid)) *DAGStore_GetSize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(cid.Cid))
	})
	return _c
}

func (_c *DAGStore_GetSize_Call) Return(_a0 int, _a1 error) *DAGStore_GetSize_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DAGStore_GetSize_Call) RunAndReturn(run func(context.Context, cid.Cid) (int, error)) *DAGStore_GetSize_Call {
	_c.Call.Return(run)
	return _c
}

// Has provides a mock function with given fields: _a0, _a1
func (_m *DAGStore) Has(_a0 context.Context, _a1 cid.Cid) (bool, error) {
	ret := _m.Called(_a0, _a1)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) (bool, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, cid.Cid) bool); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, cid.Cid) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DAGStore_Has_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Has'
type DAGStore_Has_Call struct {
	*mock.Call
}

// Has is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 cid.Cid
func (_e *DAGStore_Expecter) Has(_a0 interface{}, _a1 interface{}) *DAGStore_Has_Call {
	return &DAGStore_Has_Call{Call: _e.mock.On("Has", _a0, _a1)}
}

func (_c *DAGStore_Has_Call) Run(run func(_a0 context.Context, _a1 cid.Cid)) *DAGStore_Has_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(cid.Cid))
	})
	return _c
}

func (_c *DAGStore_Has_Call) Return(_a0 bool, _a1 error) *DAGStore_Has_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *DAGStore_Has_Call) RunAndReturn(run func(context.Context, cid.Cid) (bool, error)) *DAGStore_Has_Call {
	_c.Call.Return(run)
	return _c
}

// HashOnRead provides a mock function with given fields: enabled
func (_m *DAGStore) HashOnRead(enabled bool) {
	_m.Called(enabled)
}

// DAGStore_HashOnRead_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HashOnRead'
type DAGStore_HashOnRead_Call struct {
	*mock.Call
}

// HashOnRead is a helper method to define mock.On call
//   - enabled bool
func (_e *DAGStore_Expecter) HashOnRead(enabled interface{}) *DAGStore_HashOnRead_Call {
	return &DAGStore_HashOnRead_Call{Call: _e.mock.On("HashOnRead", enabled)}
}

func (_c *DAGStore_HashOnRead_Call) Run(run func(enabled bool)) *DAGStore_HashOnRead_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bool))
	})
	return _c
}

func (_c *DAGStore_HashOnRead_Call) Return() *DAGStore_HashOnRead_Call {
	_c.Call.Return()
	return _c
}

func (_c *DAGStore_HashOnRead_Call) RunAndReturn(run func(bool)) *DAGStore_HashOnRead_Call {
	_c.Call.Return(run)
	return _c
}

// Put provides a mock function with given fields: _a0, _a1
func (_m *DAGStore) Put(_a0 context.Context, _a1 blocks.Block) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, blocks.Block) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DAGStore_Put_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Put'
type DAGStore_Put_Call struct {
	*mock.Call
}

// Put is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 blocks.Block
func (_e *DAGStore_Expecter) Put(_a0 interface{}, _a1 interface{}) *DAGStore_Put_Call {
	return &DAGStore_Put_Call{Call: _e.mock.On("Put", _a0, _a1)}
}

func (_c *DAGStore_Put_Call) Run(run func(_a0 context.Context, _a1 blocks.Block)) *DAGStore_Put_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(blocks.Block))
	})
	return _c
}

func (_c *DAGStore_Put_Call) Return(_a0 error) *DAGStore_Put_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DAGStore_Put_Call) RunAndReturn(run func(context.Context, blocks.Block) error) *DAGStore_Put_Call {
	_c.Call.Return(run)
	return _c
}

// PutMany provides a mock function with given fields: _a0, _a1
func (_m *DAGStore) PutMany(_a0 context.Context, _a1 []blocks.Block) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []blocks.Block) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DAGStore_PutMany_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PutMany'
type DAGStore_PutMany_Call struct {
	*mock.Call
}

// PutMany is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 []blocks.Block
func (_e *DAGStore_Expecter) PutMany(_a0 interface{}, _a1 interface{}) *DAGStore_PutMany_Call {
	return &DAGStore_PutMany_Call{Call: _e.mock.On("PutMany", _a0, _a1)}
}

func (_c *DAGStore_PutMany_Call) Run(run func(_a0 context.Context, _a1 []blocks.Block)) *DAGStore_PutMany_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]blocks.Block))
	})
	return _c
}

func (_c *DAGStore_PutMany_Call) Return(_a0 error) *DAGStore_PutMany_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *DAGStore_PutMany_Call) RunAndReturn(run func(context.Context, []blocks.Block) error) *DAGStore_PutMany_Call {
	_c.Call.Return(run)
	return _c
}

// NewDAGStore creates a new instance of DAGStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDAGStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *DAGStore {
	mock := &DAGStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
