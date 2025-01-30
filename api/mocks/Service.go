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

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// GetAlbyOAuthSvc provides a mock function with no fields
func (_m *Service) GetAlbyOAuthSvc() alby.AlbyOAuthService {
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

// GetConfig provides a mock function with no fields
func (_m *Service) GetConfig() config.Config {
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

// GetDB provides a mock function with no fields
func (_m *Service) GetDB() *gorm.DB {
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

// GetEventPublisher provides a mock function with no fields
func (_m *Service) GetEventPublisher() events.EventPublisher {
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

// GetKeys provides a mock function with no fields
func (_m *Service) GetKeys() keys.Keys {
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

// GetLNClient provides a mock function with no fields
func (_m *Service) GetLNClient() lnclient.LNClient {
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

// GetTransactionsService provides a mock function with no fields
func (_m *Service) GetTransactionsService() transactions.TransactionsService {
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

// IsRelayReady provides a mock function with no fields
func (_m *Service) IsRelayReady() bool {
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

// Shutdown provides a mock function with no fields
func (_m *Service) Shutdown() {
	_m.Called()
}

// StartApp provides a mock function with given fields: encryptionKey
func (_m *Service) StartApp(encryptionKey string) error {
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

// StopApp provides a mock function with no fields
func (_m *Service) StopApp() {
	_m.Called()
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
