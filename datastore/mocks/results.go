// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"github.com/ipfs/go-datastore/query"
	mock "github.com/stretchr/testify/mock"
)

// NewResults creates a new instance of Results. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewResults(t interface {
	mock.TestingT
	Cleanup(func())
}) *Results {
	mock := &Results{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Results is an autogenerated mock type for the Results type
type Results struct {
	mock.Mock
}

type Results_Expecter struct {
	mock *mock.Mock
}

func (_m *Results) EXPECT() *Results_Expecter {
	return &Results_Expecter{mock: &_m.Mock}
}

// Close provides a mock function for the type Results
func (_mock *Results) Close() error {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func() error); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// Results_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Results_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Results_Expecter) Close() *Results_Close_Call {
	return &Results_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Results_Close_Call) Run(run func()) *Results_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Results_Close_Call) Return(err error) *Results_Close_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Results_Close_Call) RunAndReturn(run func() error) *Results_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Done provides a mock function for the type Results
func (_mock *Results) Done() <-chan struct{} {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Done")
	}

	var r0 <-chan struct{}
	if returnFunc, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}
	return r0
}

// Results_Done_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Done'
type Results_Done_Call struct {
	*mock.Call
}

// Done is a helper method to define mock.On call
func (_e *Results_Expecter) Done() *Results_Done_Call {
	return &Results_Done_Call{Call: _e.mock.On("Done")}
}

func (_c *Results_Done_Call) Run(run func()) *Results_Done_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Results_Done_Call) Return(valCh <-chan struct{}) *Results_Done_Call {
	_c.Call.Return(valCh)
	return _c
}

func (_c *Results_Done_Call) RunAndReturn(run func() <-chan struct{}) *Results_Done_Call {
	_c.Call.Return(run)
	return _c
}

// Next provides a mock function for the type Results
func (_mock *Results) Next() <-chan query.Result {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 <-chan query.Result
	if returnFunc, ok := ret.Get(0).(func() <-chan query.Result); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan query.Result)
		}
	}
	return r0
}

// Results_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type Results_Next_Call struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
func (_e *Results_Expecter) Next() *Results_Next_Call {
	return &Results_Next_Call{Call: _e.mock.On("Next")}
}

func (_c *Results_Next_Call) Run(run func()) *Results_Next_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Results_Next_Call) Return(resultCh <-chan query.Result) *Results_Next_Call {
	_c.Call.Return(resultCh)
	return _c
}

func (_c *Results_Next_Call) RunAndReturn(run func() <-chan query.Result) *Results_Next_Call {
	_c.Call.Return(run)
	return _c
}

// NextSync provides a mock function for the type Results
func (_mock *Results) NextSync() (query.Result, bool) {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for NextSync")
	}

	var r0 query.Result
	var r1 bool
	if returnFunc, ok := ret.Get(0).(func() (query.Result, bool)); ok {
		return returnFunc()
	}
	if returnFunc, ok := ret.Get(0).(func() query.Result); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(query.Result)
	}
	if returnFunc, ok := ret.Get(1).(func() bool); ok {
		r1 = returnFunc()
	} else {
		r1 = ret.Get(1).(bool)
	}
	return r0, r1
}

// Results_NextSync_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NextSync'
type Results_NextSync_Call struct {
	*mock.Call
}

// NextSync is a helper method to define mock.On call
func (_e *Results_Expecter) NextSync() *Results_NextSync_Call {
	return &Results_NextSync_Call{Call: _e.mock.On("NextSync")}
}

func (_c *Results_NextSync_Call) Run(run func()) *Results_NextSync_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Results_NextSync_Call) Return(result query.Result, b bool) *Results_NextSync_Call {
	_c.Call.Return(result, b)
	return _c
}

func (_c *Results_NextSync_Call) RunAndReturn(run func() (query.Result, bool)) *Results_NextSync_Call {
	_c.Call.Return(run)
	return _c
}

// Query provides a mock function for the type Results
func (_mock *Results) Query() query.Query {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Query")
	}

	var r0 query.Query
	if returnFunc, ok := ret.Get(0).(func() query.Query); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(query.Query)
	}
	return r0
}

// Results_Query_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Query'
type Results_Query_Call struct {
	*mock.Call
}

// Query is a helper method to define mock.On call
func (_e *Results_Expecter) Query() *Results_Query_Call {
	return &Results_Query_Call{Call: _e.mock.On("Query")}
}

func (_c *Results_Query_Call) Run(run func()) *Results_Query_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Results_Query_Call) Return(query1 query.Query) *Results_Query_Call {
	_c.Call.Return(query1)
	return _c
}

func (_c *Results_Query_Call) RunAndReturn(run func() query.Query) *Results_Query_Call {
	_c.Call.Return(run)
	return _c
}

// Rest provides a mock function for the type Results
func (_mock *Results) Rest() ([]query.Entry, error) {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Rest")
	}

	var r0 []query.Entry
	var r1 error
	if returnFunc, ok := ret.Get(0).(func() ([]query.Entry, error)); ok {
		return returnFunc()
	}
	if returnFunc, ok := ret.Get(0).(func() []query.Entry); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]query.Entry)
		}
	}
	if returnFunc, ok := ret.Get(1).(func() error); ok {
		r1 = returnFunc()
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Results_Rest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rest'
type Results_Rest_Call struct {
	*mock.Call
}

// Rest is a helper method to define mock.On call
func (_e *Results_Expecter) Rest() *Results_Rest_Call {
	return &Results_Rest_Call{Call: _e.mock.On("Rest")}
}

func (_c *Results_Rest_Call) Run(run func()) *Results_Rest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Results_Rest_Call) Return(entrys []query.Entry, err error) *Results_Rest_Call {
	_c.Call.Return(entrys, err)
	return _c
}

func (_c *Results_Rest_Call) RunAndReturn(run func() ([]query.Entry, error)) *Results_Rest_Call {
	_c.Call.Return(run)
	return _c
}
