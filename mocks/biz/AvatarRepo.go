// Code generated by mockery. DO NOT EDIT.

package biz

import (
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// AvatarRepo is an autogenerated mock type for the AvatarRepo type
type AvatarRepo struct {
	mock.Mock
}

type AvatarRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *AvatarRepo) EXPECT() *AvatarRepo_Expecter {
	return &AvatarRepo_Expecter{mock: &_m.Mock}
}

// GetByType provides a mock function with given fields: avatarType, option
func (_m *AvatarRepo) GetByType(avatarType string, option ...string) ([]byte, time.Time, error) {
	_va := make([]interface{}, len(option))
	for _i := range option {
		_va[_i] = option[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, avatarType)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetByType")
	}

	var r0 []byte
	var r1 time.Time
	var r2 error
	if rf, ok := ret.Get(0).(func(string, ...string) ([]byte, time.Time, error)); ok {
		return rf(avatarType, option...)
	}
	if rf, ok := ret.Get(0).(func(string, ...string) []byte); ok {
		r0 = rf(avatarType, option...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string, ...string) time.Time); ok {
		r1 = rf(avatarType, option...)
	} else {
		r1 = ret.Get(1).(time.Time)
	}

	if rf, ok := ret.Get(2).(func(string, ...string) error); ok {
		r2 = rf(avatarType, option...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// AvatarRepo_GetByType_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetByType'
type AvatarRepo_GetByType_Call struct {
	*mock.Call
}

// GetByType is a helper method to define mock.On call
//   - avatarType string
//   - option ...string
func (_e *AvatarRepo_Expecter) GetByType(avatarType interface{}, option ...interface{}) *AvatarRepo_GetByType_Call {
	return &AvatarRepo_GetByType_Call{Call: _e.mock.On("GetByType",
		append([]interface{}{avatarType}, option...)...)}
}

func (_c *AvatarRepo_GetByType_Call) Run(run func(avatarType string, option ...string)) *AvatarRepo_GetByType_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *AvatarRepo_GetByType_Call) Return(_a0 []byte, _a1 time.Time, _a2 error) *AvatarRepo_GetByType_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *AvatarRepo_GetByType_Call) RunAndReturn(run func(string, ...string) ([]byte, time.Time, error)) *AvatarRepo_GetByType_Call {
	_c.Call.Return(run)
	return _c
}

// GetGravatarByHash provides a mock function with given fields: hash
func (_m *AvatarRepo) GetGravatarByHash(hash string) ([]byte, time.Time, error) {
	ret := _m.Called(hash)

	if len(ret) == 0 {
		panic("no return value specified for GetGravatarByHash")
	}

	var r0 []byte
	var r1 time.Time
	var r2 error
	if rf, ok := ret.Get(0).(func(string) ([]byte, time.Time, error)); ok {
		return rf(hash)
	}
	if rf, ok := ret.Get(0).(func(string) []byte); ok {
		r0 = rf(hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string) time.Time); ok {
		r1 = rf(hash)
	} else {
		r1 = ret.Get(1).(time.Time)
	}

	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(hash)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// AvatarRepo_GetGravatarByHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetGravatarByHash'
type AvatarRepo_GetGravatarByHash_Call struct {
	*mock.Call
}

// GetGravatarByHash is a helper method to define mock.On call
//   - hash string
func (_e *AvatarRepo_Expecter) GetGravatarByHash(hash interface{}) *AvatarRepo_GetGravatarByHash_Call {
	return &AvatarRepo_GetGravatarByHash_Call{Call: _e.mock.On("GetGravatarByHash", hash)}
}

func (_c *AvatarRepo_GetGravatarByHash_Call) Run(run func(hash string)) *AvatarRepo_GetGravatarByHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *AvatarRepo_GetGravatarByHash_Call) Return(_a0 []byte, _a1 time.Time, _a2 error) *AvatarRepo_GetGravatarByHash_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *AvatarRepo_GetGravatarByHash_Call) RunAndReturn(run func(string) ([]byte, time.Time, error)) *AvatarRepo_GetGravatarByHash_Call {
	_c.Call.Return(run)
	return _c
}

// GetQqByHash provides a mock function with given fields: hash
func (_m *AvatarRepo) GetQqByHash(hash string) (string, []byte, time.Time, error) {
	ret := _m.Called(hash)

	if len(ret) == 0 {
		panic("no return value specified for GetQqByHash")
	}

	var r0 string
	var r1 []byte
	var r2 time.Time
	var r3 error
	if rf, ok := ret.Get(0).(func(string) (string, []byte, time.Time, error)); ok {
		return rf(hash)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(hash)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) []byte); ok {
		r1 = rf(hash)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	if rf, ok := ret.Get(2).(func(string) time.Time); ok {
		r2 = rf(hash)
	} else {
		r2 = ret.Get(2).(time.Time)
	}

	if rf, ok := ret.Get(3).(func(string) error); ok {
		r3 = rf(hash)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// AvatarRepo_GetQqByHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetQqByHash'
type AvatarRepo_GetQqByHash_Call struct {
	*mock.Call
}

// GetQqByHash is a helper method to define mock.On call
//   - hash string
func (_e *AvatarRepo_Expecter) GetQqByHash(hash interface{}) *AvatarRepo_GetQqByHash_Call {
	return &AvatarRepo_GetQqByHash_Call{Call: _e.mock.On("GetQqByHash", hash)}
}

func (_c *AvatarRepo_GetQqByHash_Call) Run(run func(hash string)) *AvatarRepo_GetQqByHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *AvatarRepo_GetQqByHash_Call) Return(_a0 string, _a1 []byte, _a2 time.Time, _a3 error) *AvatarRepo_GetQqByHash_Call {
	_c.Call.Return(_a0, _a1, _a2, _a3)
	return _c
}

func (_c *AvatarRepo_GetQqByHash_Call) RunAndReturn(run func(string) (string, []byte, time.Time, error)) *AvatarRepo_GetQqByHash_Call {
	_c.Call.Return(run)
	return _c
}

// GetWeAvatar provides a mock function with given fields: hash, appID
func (_m *AvatarRepo) GetWeAvatar(hash string, appID string) ([]byte, time.Time, error) {
	ret := _m.Called(hash, appID)

	if len(ret) == 0 {
		panic("no return value specified for GetWeAvatar")
	}

	var r0 []byte
	var r1 time.Time
	var r2 error
	if rf, ok := ret.Get(0).(func(string, string) ([]byte, time.Time, error)); ok {
		return rf(hash, appID)
	}
	if rf, ok := ret.Get(0).(func(string, string) []byte); ok {
		r0 = rf(hash, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) time.Time); ok {
		r1 = rf(hash, appID)
	} else {
		r1 = ret.Get(1).(time.Time)
	}

	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(hash, appID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// AvatarRepo_GetWeAvatar_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetWeAvatar'
type AvatarRepo_GetWeAvatar_Call struct {
	*mock.Call
}

// GetWeAvatar is a helper method to define mock.On call
//   - hash string
//   - appID string
func (_e *AvatarRepo_Expecter) GetWeAvatar(hash interface{}, appID interface{}) *AvatarRepo_GetWeAvatar_Call {
	return &AvatarRepo_GetWeAvatar_Call{Call: _e.mock.On("GetWeAvatar", hash, appID)}
}

func (_c *AvatarRepo_GetWeAvatar_Call) Run(run func(hash string, appID string)) *AvatarRepo_GetWeAvatar_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *AvatarRepo_GetWeAvatar_Call) Return(_a0 []byte, _a1 time.Time, _a2 error) *AvatarRepo_GetWeAvatar_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *AvatarRepo_GetWeAvatar_Call) RunAndReturn(run func(string, string) ([]byte, time.Time, error)) *AvatarRepo_GetWeAvatar_Call {
	_c.Call.Return(run)
	return _c
}

// IsBanned provides a mock function with given fields: img
func (_m *AvatarRepo) IsBanned(img []byte) (bool, error) {
	ret := _m.Called(img)

	if len(ret) == 0 {
		panic("no return value specified for IsBanned")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (bool, error)); ok {
		return rf(img)
	}
	if rf, ok := ret.Get(0).(func([]byte) bool); ok {
		r0 = rf(img)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(img)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AvatarRepo_IsBanned_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsBanned'
type AvatarRepo_IsBanned_Call struct {
	*mock.Call
}

// IsBanned is a helper method to define mock.On call
//   - img []byte
func (_e *AvatarRepo_Expecter) IsBanned(img interface{}) *AvatarRepo_IsBanned_Call {
	return &AvatarRepo_IsBanned_Call{Call: _e.mock.On("IsBanned", img)}
}

func (_c *AvatarRepo_IsBanned_Call) Run(run func(img []byte)) *AvatarRepo_IsBanned_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte))
	})
	return _c
}

func (_c *AvatarRepo_IsBanned_Call) Return(_a0 bool, _a1 error) *AvatarRepo_IsBanned_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *AvatarRepo_IsBanned_Call) RunAndReturn(run func([]byte) (bool, error)) *AvatarRepo_IsBanned_Call {
	_c.Call.Return(run)
	return _c
}

// NewAvatarRepo creates a new instance of AvatarRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAvatarRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *AvatarRepo {
	mock := &AvatarRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
