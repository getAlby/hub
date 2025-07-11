// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"github.com/getAlby/hub/config"
	mock "github.com/stretchr/testify/mock"
)

// NewMockConfig creates a new instance of MockConfig. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockConfig(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConfig {
	mock := &MockConfig{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockConfig is an autogenerated mock type for the Config type
type MockConfig struct {
	mock.Mock
}

type MockConfig_Expecter struct {
	mock *mock.Mock
}

func (_m *MockConfig) EXPECT() *MockConfig_Expecter {
	return &MockConfig_Expecter{mock: &_m.Mock}
}

// ChangeUnlockPassword provides a mock function for the type MockConfig
func (_mock *MockConfig) ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error {
	ret := _mock.Called(currentUnlockPassword, newUnlockPassword)

	if len(ret) == 0 {
		panic("no return value specified for ChangeUnlockPassword")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = returnFunc(currentUnlockPassword, newUnlockPassword)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockConfig_ChangeUnlockPassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChangeUnlockPassword'
type MockConfig_ChangeUnlockPassword_Call struct {
	*mock.Call
}

// ChangeUnlockPassword is a helper method to define mock.On call
//   - currentUnlockPassword string
//   - newUnlockPassword string
func (_e *MockConfig_Expecter) ChangeUnlockPassword(currentUnlockPassword interface{}, newUnlockPassword interface{}) *MockConfig_ChangeUnlockPassword_Call {
	return &MockConfig_ChangeUnlockPassword_Call{Call: _e.mock.On("ChangeUnlockPassword", currentUnlockPassword, newUnlockPassword)}
}

func (_c *MockConfig_ChangeUnlockPassword_Call) Run(run func(currentUnlockPassword string, newUnlockPassword string)) *MockConfig_ChangeUnlockPassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *MockConfig_ChangeUnlockPassword_Call) Return(err error) *MockConfig_ChangeUnlockPassword_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockConfig_ChangeUnlockPassword_Call) RunAndReturn(run func(currentUnlockPassword string, newUnlockPassword string) error) *MockConfig_ChangeUnlockPassword_Call {
	_c.Call.Return(run)
	return _c
}

// CheckUnlockPassword provides a mock function for the type MockConfig
func (_mock *MockConfig) CheckUnlockPassword(password string) bool {
	ret := _mock.Called(password)

	if len(ret) == 0 {
		panic("no return value specified for CheckUnlockPassword")
	}

	var r0 bool
	if returnFunc, ok := ret.Get(0).(func(string) bool); ok {
		r0 = returnFunc(password)
	} else {
		r0 = ret.Get(0).(bool)
	}
	return r0
}

// MockConfig_CheckUnlockPassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckUnlockPassword'
type MockConfig_CheckUnlockPassword_Call struct {
	*mock.Call
}

// CheckUnlockPassword is a helper method to define mock.On call
//   - password string
func (_e *MockConfig_Expecter) CheckUnlockPassword(password interface{}) *MockConfig_CheckUnlockPassword_Call {
	return &MockConfig_CheckUnlockPassword_Call{Call: _e.mock.On("CheckUnlockPassword", password)}
}

func (_c *MockConfig_CheckUnlockPassword_Call) Run(run func(password string)) *MockConfig_CheckUnlockPassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *MockConfig_CheckUnlockPassword_Call) Return(b bool) *MockConfig_CheckUnlockPassword_Call {
	_c.Call.Return(b)
	return _c
}

func (_c *MockConfig_CheckUnlockPassword_Call) RunAndReturn(run func(password string) bool) *MockConfig_CheckUnlockPassword_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function for the type MockConfig
func (_mock *MockConfig) Get(key string, encryptionKey string) (string, error) {
	ret := _mock.Called(key, encryptionKey)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return returnFunc(key, encryptionKey)
	}
	if returnFunc, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = returnFunc(key, encryptionKey)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = returnFunc(key, encryptionKey)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockConfig_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockConfig_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - key string
//   - encryptionKey string
func (_e *MockConfig_Expecter) Get(key interface{}, encryptionKey interface{}) *MockConfig_Get_Call {
	return &MockConfig_Get_Call{Call: _e.mock.On("Get", key, encryptionKey)}
}

func (_c *MockConfig_Get_Call) Run(run func(key string, encryptionKey string)) *MockConfig_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *MockConfig_Get_Call) Return(s string, err error) *MockConfig_Get_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockConfig_Get_Call) RunAndReturn(run func(key string, encryptionKey string) (string, error)) *MockConfig_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetCurrency provides a mock function for the type MockConfig
func (_mock *MockConfig) GetCurrency() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCurrency")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// MockConfig_GetCurrency_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCurrency'
type MockConfig_GetCurrency_Call struct {
	*mock.Call
}

// GetCurrency is a helper method to define mock.On call
func (_e *MockConfig_Expecter) GetCurrency() *MockConfig_GetCurrency_Call {
	return &MockConfig_GetCurrency_Call{Call: _e.mock.On("GetCurrency")}
}

func (_c *MockConfig_GetCurrency_Call) Run(run func()) *MockConfig_GetCurrency_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_GetCurrency_Call) Return(s string) *MockConfig_GetCurrency_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *MockConfig_GetCurrency_Call) RunAndReturn(run func() string) *MockConfig_GetCurrency_Call {
	_c.Call.Return(run)
	return _c
}

// GetEnv provides a mock function for the type MockConfig
func (_mock *MockConfig) GetEnv() *config.AppConfig {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetEnv")
	}

	var r0 *config.AppConfig
	if returnFunc, ok := ret.Get(0).(func() *config.AppConfig); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*config.AppConfig)
		}
	}
	return r0
}

// MockConfig_GetEnv_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEnv'
type MockConfig_GetEnv_Call struct {
	*mock.Call
}

// GetEnv is a helper method to define mock.On call
func (_e *MockConfig_Expecter) GetEnv() *MockConfig_GetEnv_Call {
	return &MockConfig_GetEnv_Call{Call: _e.mock.On("GetEnv")}
}

func (_c *MockConfig_GetEnv_Call) Run(run func()) *MockConfig_GetEnv_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_GetEnv_Call) Return(appConfig *config.AppConfig) *MockConfig_GetEnv_Call {
	_c.Call.Return(appConfig)
	return _c
}

func (_c *MockConfig_GetEnv_Call) RunAndReturn(run func() *config.AppConfig) *MockConfig_GetEnv_Call {
	_c.Call.Return(run)
	return _c
}

// GetJWTSecret provides a mock function for the type MockConfig
func (_mock *MockConfig) GetJWTSecret() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetJWTSecret")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// MockConfig_GetJWTSecret_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJWTSecret'
type MockConfig_GetJWTSecret_Call struct {
	*mock.Call
}

// GetJWTSecret is a helper method to define mock.On call
func (_e *MockConfig_Expecter) GetJWTSecret() *MockConfig_GetJWTSecret_Call {
	return &MockConfig_GetJWTSecret_Call{Call: _e.mock.On("GetJWTSecret")}
}

func (_c *MockConfig_GetJWTSecret_Call) Run(run func()) *MockConfig_GetJWTSecret_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_GetJWTSecret_Call) Return(s string) *MockConfig_GetJWTSecret_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *MockConfig_GetJWTSecret_Call) RunAndReturn(run func() string) *MockConfig_GetJWTSecret_Call {
	_c.Call.Return(run)
	return _c
}

// GetMempoolUrl provides a mock function for the type MockConfig
func (_mock *MockConfig) GetMempoolUrl() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMempoolUrl")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// MockConfig_GetMempoolUrl_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMempoolUrl'
type MockConfig_GetMempoolUrl_Call struct {
	*mock.Call
}

// GetMempoolUrl is a helper method to define mock.On call
func (_e *MockConfig_Expecter) GetMempoolUrl() *MockConfig_GetMempoolUrl_Call {
	return &MockConfig_GetMempoolUrl_Call{Call: _e.mock.On("GetMempoolUrl")}
}

func (_c *MockConfig_GetMempoolUrl_Call) Run(run func()) *MockConfig_GetMempoolUrl_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_GetMempoolUrl_Call) Return(s string) *MockConfig_GetMempoolUrl_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *MockConfig_GetMempoolUrl_Call) RunAndReturn(run func() string) *MockConfig_GetMempoolUrl_Call {
	_c.Call.Return(run)
	return _c
}

// GetNetwork provides a mock function for the type MockConfig
func (_mock *MockConfig) GetNetwork() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetNetwork")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// MockConfig_GetNetwork_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNetwork'
type MockConfig_GetNetwork_Call struct {
	*mock.Call
}

// GetNetwork is a helper method to define mock.On call
func (_e *MockConfig_Expecter) GetNetwork() *MockConfig_GetNetwork_Call {
	return &MockConfig_GetNetwork_Call{Call: _e.mock.On("GetNetwork")}
}

func (_c *MockConfig_GetNetwork_Call) Run(run func()) *MockConfig_GetNetwork_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_GetNetwork_Call) Return(s string) *MockConfig_GetNetwork_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *MockConfig_GetNetwork_Call) RunAndReturn(run func() string) *MockConfig_GetNetwork_Call {
	_c.Call.Return(run)
	return _c
}

// GetRelayUrl provides a mock function for the type MockConfig
func (_mock *MockConfig) GetRelayUrl() string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetRelayUrl")
	}

	var r0 string
	if returnFunc, ok := ret.Get(0).(func() string); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(string)
	}
	return r0
}

// MockConfig_GetRelayUrl_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetRelayUrl'
type MockConfig_GetRelayUrl_Call struct {
	*mock.Call
}

// GetRelayUrl is a helper method to define mock.On call
func (_e *MockConfig_Expecter) GetRelayUrl() *MockConfig_GetRelayUrl_Call {
	return &MockConfig_GetRelayUrl_Call{Call: _e.mock.On("GetRelayUrl")}
}

func (_c *MockConfig_GetRelayUrl_Call) Run(run func()) *MockConfig_GetRelayUrl_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_GetRelayUrl_Call) Return(s string) *MockConfig_GetRelayUrl_Call {
	_c.Call.Return(s)
	return _c
}

func (_c *MockConfig_GetRelayUrl_Call) RunAndReturn(run func() string) *MockConfig_GetRelayUrl_Call {
	_c.Call.Return(run)
	return _c
}

// SaveUnlockPasswordCheck provides a mock function for the type MockConfig
func (_mock *MockConfig) SaveUnlockPasswordCheck(encryptionKey string) error {
	ret := _mock.Called(encryptionKey)

	if len(ret) == 0 {
		panic("no return value specified for SaveUnlockPasswordCheck")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string) error); ok {
		r0 = returnFunc(encryptionKey)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockConfig_SaveUnlockPasswordCheck_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SaveUnlockPasswordCheck'
type MockConfig_SaveUnlockPasswordCheck_Call struct {
	*mock.Call
}

// SaveUnlockPasswordCheck is a helper method to define mock.On call
//   - encryptionKey string
func (_e *MockConfig_Expecter) SaveUnlockPasswordCheck(encryptionKey interface{}) *MockConfig_SaveUnlockPasswordCheck_Call {
	return &MockConfig_SaveUnlockPasswordCheck_Call{Call: _e.mock.On("SaveUnlockPasswordCheck", encryptionKey)}
}

func (_c *MockConfig_SaveUnlockPasswordCheck_Call) Run(run func(encryptionKey string)) *MockConfig_SaveUnlockPasswordCheck_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *MockConfig_SaveUnlockPasswordCheck_Call) Return(err error) *MockConfig_SaveUnlockPasswordCheck_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockConfig_SaveUnlockPasswordCheck_Call) RunAndReturn(run func(encryptionKey string) error) *MockConfig_SaveUnlockPasswordCheck_Call {
	_c.Call.Return(run)
	return _c
}

// SetAutoUnlockPassword provides a mock function for the type MockConfig
func (_mock *MockConfig) SetAutoUnlockPassword(unlockPassword string) error {
	ret := _mock.Called(unlockPassword)

	if len(ret) == 0 {
		panic("no return value specified for SetAutoUnlockPassword")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string) error); ok {
		r0 = returnFunc(unlockPassword)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockConfig_SetAutoUnlockPassword_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetAutoUnlockPassword'
type MockConfig_SetAutoUnlockPassword_Call struct {
	*mock.Call
}

// SetAutoUnlockPassword is a helper method to define mock.On call
//   - unlockPassword string
func (_e *MockConfig_Expecter) SetAutoUnlockPassword(unlockPassword interface{}) *MockConfig_SetAutoUnlockPassword_Call {
	return &MockConfig_SetAutoUnlockPassword_Call{Call: _e.mock.On("SetAutoUnlockPassword", unlockPassword)}
}

func (_c *MockConfig_SetAutoUnlockPassword_Call) Run(run func(unlockPassword string)) *MockConfig_SetAutoUnlockPassword_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *MockConfig_SetAutoUnlockPassword_Call) Return(err error) *MockConfig_SetAutoUnlockPassword_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockConfig_SetAutoUnlockPassword_Call) RunAndReturn(run func(unlockPassword string) error) *MockConfig_SetAutoUnlockPassword_Call {
	_c.Call.Return(run)
	return _c
}

// SetCurrency provides a mock function for the type MockConfig
func (_mock *MockConfig) SetCurrency(value string) error {
	ret := _mock.Called(value)

	if len(ret) == 0 {
		panic("no return value specified for SetCurrency")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string) error); ok {
		r0 = returnFunc(value)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockConfig_SetCurrency_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetCurrency'
type MockConfig_SetCurrency_Call struct {
	*mock.Call
}

// SetCurrency is a helper method to define mock.On call
//   - value string
func (_e *MockConfig_Expecter) SetCurrency(value interface{}) *MockConfig_SetCurrency_Call {
	return &MockConfig_SetCurrency_Call{Call: _e.mock.On("SetCurrency", value)}
}

func (_c *MockConfig_SetCurrency_Call) Run(run func(value string)) *MockConfig_SetCurrency_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *MockConfig_SetCurrency_Call) Return(err error) *MockConfig_SetCurrency_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockConfig_SetCurrency_Call) RunAndReturn(run func(value string) error) *MockConfig_SetCurrency_Call {
	_c.Call.Return(run)
	return _c
}

// SetIgnore provides a mock function for the type MockConfig
func (_mock *MockConfig) SetIgnore(key string, value string, encryptionKey string) error {
	ret := _mock.Called(key, value, encryptionKey)

	if len(ret) == 0 {
		panic("no return value specified for SetIgnore")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = returnFunc(key, value, encryptionKey)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockConfig_SetIgnore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetIgnore'
type MockConfig_SetIgnore_Call struct {
	*mock.Call
}

// SetIgnore is a helper method to define mock.On call
//   - key string
//   - value string
//   - encryptionKey string
func (_e *MockConfig_Expecter) SetIgnore(key interface{}, value interface{}, encryptionKey interface{}) *MockConfig_SetIgnore_Call {
	return &MockConfig_SetIgnore_Call{Call: _e.mock.On("SetIgnore", key, value, encryptionKey)}
}

func (_c *MockConfig_SetIgnore_Call) Run(run func(key string, value string, encryptionKey string)) *MockConfig_SetIgnore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 string
		if args[2] != nil {
			arg2 = args[2].(string)
		}
		run(
			arg0,
			arg1,
			arg2,
		)
	})
	return _c
}

func (_c *MockConfig_SetIgnore_Call) Return(err error) *MockConfig_SetIgnore_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockConfig_SetIgnore_Call) RunAndReturn(run func(key string, value string, encryptionKey string) error) *MockConfig_SetIgnore_Call {
	_c.Call.Return(run)
	return _c
}

// SetUpdate provides a mock function for the type MockConfig
func (_mock *MockConfig) SetUpdate(key string, value string, encryptionKey string) error {
	ret := _mock.Called(key, value, encryptionKey)

	if len(ret) == 0 {
		panic("no return value specified for SetUpdate")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = returnFunc(key, value, encryptionKey)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockConfig_SetUpdate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetUpdate'
type MockConfig_SetUpdate_Call struct {
	*mock.Call
}

// SetUpdate is a helper method to define mock.On call
//   - key string
//   - value string
//   - encryptionKey string
func (_e *MockConfig_Expecter) SetUpdate(key interface{}, value interface{}, encryptionKey interface{}) *MockConfig_SetUpdate_Call {
	return &MockConfig_SetUpdate_Call{Call: _e.mock.On("SetUpdate", key, value, encryptionKey)}
}

func (_c *MockConfig_SetUpdate_Call) Run(run func(key string, value string, encryptionKey string)) *MockConfig_SetUpdate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 string
		if args[2] != nil {
			arg2 = args[2].(string)
		}
		run(
			arg0,
			arg1,
			arg2,
		)
	})
	return _c
}

func (_c *MockConfig_SetUpdate_Call) Return(err error) *MockConfig_SetUpdate_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockConfig_SetUpdate_Call) RunAndReturn(run func(key string, value string, encryptionKey string) error) *MockConfig_SetUpdate_Call {
	_c.Call.Return(run)
	return _c
}

// SetupCompleted provides a mock function for the type MockConfig
func (_mock *MockConfig) SetupCompleted() bool {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for SetupCompleted")
	}

	var r0 bool
	if returnFunc, ok := ret.Get(0).(func() bool); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(bool)
	}
	return r0
}

// MockConfig_SetupCompleted_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetupCompleted'
type MockConfig_SetupCompleted_Call struct {
	*mock.Call
}

// SetupCompleted is a helper method to define mock.On call
func (_e *MockConfig_Expecter) SetupCompleted() *MockConfig_SetupCompleted_Call {
	return &MockConfig_SetupCompleted_Call{Call: _e.mock.On("SetupCompleted")}
}

func (_c *MockConfig_SetupCompleted_Call) Run(run func()) *MockConfig_SetupCompleted_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockConfig_SetupCompleted_Call) Return(b bool) *MockConfig_SetupCompleted_Call {
	_c.Call.Return(b)
	return _c
}

func (_c *MockConfig_SetupCompleted_Call) RunAndReturn(run func() bool) *MockConfig_SetupCompleted_Call {
	_c.Call.Return(run)
	return _c
}
