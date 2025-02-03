// Code generated by mockery v2.51.1. DO NOT EDIT.

package mocks

import (
	alby "github.com/getAlby/hub/alby"
	config "github.com/getAlby/hub/config"

	events "github.com/getAlby/hub/events"

	gorm "gorm.io/gorm"

	keys "github.com/getAlby/hub/service/keys"

	lnclient "github.com/getAlby/hub/lnclient"

	mock "github.com/stretchr/testify/mock"

	transactions "github.com/getAlby/hub/transactions"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

type MockService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockService) EXPECT() *MockService_Expecter {
	return &MockService_Expecter{mock: &_m.Mock}
}

// GetAlbyOAuthSvc provides a mock function with no fields
func (_m *MockService) GetAlbyOAuthSvc() alby.AlbyOAuthService {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAlbyOAuthSvc")
	}

	var r0 alby.AlbyOAuthService
	if rf, ok := ret.Get(0).(func() alby.AlbyOAuthService); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(alby.AlbyOAuthService)
		}
	}

	return r0
}

// MockService_GetAlbyOAuthSvc_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAlbyOAuthSvc'
type MockService_GetAlbyOAuthSvc_Call struct {
	*mock.Call
}

// GetAlbyOAuthSvc is a helper method to define mock.On call
func (_e *MockService_Expecter) GetAlbyOAuthSvc() *MockService_GetAlbyOAuthSvc_Call {
	return &MockService_GetAlbyOAuthSvc_Call{Call: _e.mock.On("GetAlbyOAuthSvc")}
}

func (_c *MockService_GetAlbyOAuthSvc_Call) Run(run func()) *MockService_GetAlbyOAuthSvc_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetAlbyOAuthSvc_Call) Return(_a0 alby.AlbyOAuthService) *MockService_GetAlbyOAuthSvc_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetAlbyOAuthSvc_Call) RunAndReturn(run func() alby.AlbyOAuthService) *MockService_GetAlbyOAuthSvc_Call {
	_c.Call.Return(run)
	return _c
}

// GetConfig provides a mock function with no fields
func (_m *MockService) GetConfig() config.Config {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetConfig")
	}

	var r0 config.Config
	if rf, ok := ret.Get(0).(func() config.Config); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(config.Config)
		}
	}

	return r0
}

// MockService_GetConfig_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConfig'
type MockService_GetConfig_Call struct {
	*mock.Call
}

// GetConfig is a helper method to define mock.On call
func (_e *MockService_Expecter) GetConfig() *MockService_GetConfig_Call {
	return &MockService_GetConfig_Call{Call: _e.mock.On("GetConfig")}
}

func (_c *MockService_GetConfig_Call) Run(run func()) *MockService_GetConfig_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetConfig_Call) Return(_a0 config.Config) *MockService_GetConfig_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetConfig_Call) RunAndReturn(run func() config.Config) *MockService_GetConfig_Call {
	_c.Call.Return(run)
	return _c
}

// GetDB provides a mock function with no fields
func (_m *MockService) GetDB() *gorm.DB {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetDB")
	}

	var r0 *gorm.DB
	if rf, ok := ret.Get(0).(func() *gorm.DB); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gorm.DB)
		}
	}

	return r0
}

// MockService_GetDB_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDB'
type MockService_GetDB_Call struct {
	*mock.Call
}

// GetDB is a helper method to define mock.On call
func (_e *MockService_Expecter) GetDB() *MockService_GetDB_Call {
	return &MockService_GetDB_Call{Call: _e.mock.On("GetDB")}
}

func (_c *MockService_GetDB_Call) Run(run func()) *MockService_GetDB_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetDB_Call) Return(_a0 *gorm.DB) *MockService_GetDB_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetDB_Call) RunAndReturn(run func() *gorm.DB) *MockService_GetDB_Call {
	_c.Call.Return(run)
	return _c
}

// GetEventPublisher provides a mock function with no fields
func (_m *MockService) GetEventPublisher() events.EventPublisher {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetEventPublisher")
	}

	var r0 events.EventPublisher
	if rf, ok := ret.Get(0).(func() events.EventPublisher); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(events.EventPublisher)
		}
	}

	return r0
}

// MockService_GetEventPublisher_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEventPublisher'
type MockService_GetEventPublisher_Call struct {
	*mock.Call
}

// GetEventPublisher is a helper method to define mock.On call
func (_e *MockService_Expecter) GetEventPublisher() *MockService_GetEventPublisher_Call {
	return &MockService_GetEventPublisher_Call{Call: _e.mock.On("GetEventPublisher")}
}

func (_c *MockService_GetEventPublisher_Call) Run(run func()) *MockService_GetEventPublisher_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetEventPublisher_Call) Return(_a0 events.EventPublisher) *MockService_GetEventPublisher_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetEventPublisher_Call) RunAndReturn(run func() events.EventPublisher) *MockService_GetEventPublisher_Call {
	_c.Call.Return(run)
	return _c
}

// GetKeys provides a mock function with no fields
func (_m *MockService) GetKeys() keys.Keys {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetKeys")
	}

	var r0 keys.Keys
	if rf, ok := ret.Get(0).(func() keys.Keys); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(keys.Keys)
		}
	}

	return r0
}

// MockService_GetKeys_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetKeys'
type MockService_GetKeys_Call struct {
	*mock.Call
}

// GetKeys is a helper method to define mock.On call
func (_e *MockService_Expecter) GetKeys() *MockService_GetKeys_Call {
	return &MockService_GetKeys_Call{Call: _e.mock.On("GetKeys")}
}

func (_c *MockService_GetKeys_Call) Run(run func()) *MockService_GetKeys_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetKeys_Call) Return(_a0 keys.Keys) *MockService_GetKeys_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetKeys_Call) RunAndReturn(run func() keys.Keys) *MockService_GetKeys_Call {
	_c.Call.Return(run)
	return _c
}

// GetLNClient provides a mock function with no fields
func (_m *MockService) GetLNClient() lnclient.LNClient {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetLNClient")
	}

	var r0 lnclient.LNClient
	if rf, ok := ret.Get(0).(func() lnclient.LNClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(lnclient.LNClient)
		}
	}

	return r0
}

// MockService_GetLNClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetLNClient'
type MockService_GetLNClient_Call struct {
	*mock.Call
}

// GetLNClient is a helper method to define mock.On call
func (_e *MockService_Expecter) GetLNClient() *MockService_GetLNClient_Call {
	return &MockService_GetLNClient_Call{Call: _e.mock.On("GetLNClient")}
}

func (_c *MockService_GetLNClient_Call) Run(run func()) *MockService_GetLNClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetLNClient_Call) Return(_a0 lnclient.LNClient) *MockService_GetLNClient_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetLNClient_Call) RunAndReturn(run func() lnclient.LNClient) *MockService_GetLNClient_Call {
	_c.Call.Return(run)
	return _c
}

// GetTransactionsService provides a mock function with no fields
func (_m *MockService) GetTransactionsService() transactions.TransactionsService {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetTransactionsService")
	}

	var r0 transactions.TransactionsService
	if rf, ok := ret.Get(0).(func() transactions.TransactionsService); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(transactions.TransactionsService)
		}
	}

	return r0
}

// MockService_GetTransactionsService_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTransactionsService'
type MockService_GetTransactionsService_Call struct {
	*mock.Call
}

// GetTransactionsService is a helper method to define mock.On call
func (_e *MockService_Expecter) GetTransactionsService() *MockService_GetTransactionsService_Call {
	return &MockService_GetTransactionsService_Call{Call: _e.mock.On("GetTransactionsService")}
}

func (_c *MockService_GetTransactionsService_Call) Run(run func()) *MockService_GetTransactionsService_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_GetTransactionsService_Call) Return(_a0 transactions.TransactionsService) *MockService_GetTransactionsService_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_GetTransactionsService_Call) RunAndReturn(run func() transactions.TransactionsService) *MockService_GetTransactionsService_Call {
	_c.Call.Return(run)
	return _c
}

// IsRelayReady provides a mock function with no fields
func (_m *MockService) IsRelayReady() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsRelayReady")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockService_IsRelayReady_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsRelayReady'
type MockService_IsRelayReady_Call struct {
	*mock.Call
}

// IsRelayReady is a helper method to define mock.On call
func (_e *MockService_Expecter) IsRelayReady() *MockService_IsRelayReady_Call {
	return &MockService_IsRelayReady_Call{Call: _e.mock.On("IsRelayReady")}
}

func (_c *MockService_IsRelayReady_Call) Run(run func()) *MockService_IsRelayReady_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_IsRelayReady_Call) Return(_a0 bool) *MockService_IsRelayReady_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_IsRelayReady_Call) RunAndReturn(run func() bool) *MockService_IsRelayReady_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function with no fields
func (_m *MockService) Shutdown() {
	_m.Called()
}

// MockService_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type MockService_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
func (_e *MockService_Expecter) Shutdown() *MockService_Shutdown_Call {
	return &MockService_Shutdown_Call{Call: _e.mock.On("Shutdown")}
}

func (_c *MockService_Shutdown_Call) Run(run func()) *MockService_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_Shutdown_Call) Return() *MockService_Shutdown_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockService_Shutdown_Call) RunAndReturn(run func()) *MockService_Shutdown_Call {
	_c.Run(run)
	return _c
}

// StartApp provides a mock function with given fields: encryptionKey
func (_m *MockService) StartApp(encryptionKey string) error {
	ret := _m.Called(encryptionKey)

	if len(ret) == 0 {
		panic("no return value specified for StartApp")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(encryptionKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockService_StartApp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartApp'
type MockService_StartApp_Call struct {
	*mock.Call
}

// StartApp is a helper method to define mock.On call
//   - encryptionKey string
func (_e *MockService_Expecter) StartApp(encryptionKey interface{}) *MockService_StartApp_Call {
	return &MockService_StartApp_Call{Call: _e.mock.On("StartApp", encryptionKey)}
}

func (_c *MockService_StartApp_Call) Run(run func(encryptionKey string)) *MockService_StartApp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockService_StartApp_Call) Return(_a0 error) *MockService_StartApp_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockService_StartApp_Call) RunAndReturn(run func(string) error) *MockService_StartApp_Call {
	_c.Call.Return(run)
	return _c
}

// StopApp provides a mock function with no fields
func (_m *MockService) StopApp() {
	_m.Called()
}

// MockService_StopApp_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StopApp'
type MockService_StopApp_Call struct {
	*mock.Call
}

// StopApp is a helper method to define mock.On call
func (_e *MockService_Expecter) StopApp() *MockService_StopApp_Call {
	return &MockService_StopApp_Call{Call: _e.mock.On("StopApp")}
}

func (_c *MockService_StopApp_Call) Run(run func()) *MockService_StopApp_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockService_StopApp_Call) Return() *MockService_StopApp_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockService_StopApp_Call) RunAndReturn(run func()) *MockService_StopApp_Call {
	_c.Run(run)
	return _c
}

// NewMockService creates a new instance of MockService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockService {
	mock := &MockService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
