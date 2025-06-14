// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	"internProject/todo_app_service/pkg/constants"
	"internProject/todo_app_service/pkg/models"

	mock "github.com/stretchr/testify/mock"
)

// IService is an autogenerated mock type for the IService type
type IService struct {
	mock.Mock
}

type IService_Expecter struct {
	mock *mock.Mock
}

func (_m *IService) EXPECT() *IService_Expecter {
	return &IService_Expecter{mock: &_m.Mock}
}

// CreateTodoRecord provides a mock function with given fields: listId, name, description, status
func (_m *IService) CreateTodoRecord(listId string, name string, description string, status constants.TodoStatus) (*models.Todo, error) {
	ret := _m.Called(listId, name, description, status)

	if len(ret) == 0 {
		panic("no return value specified for CreateTodoRecord")
	}

	var r0 *models.Todo
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string, constants.TodoStatus) (*models.Todo, error)); ok {
		return rf(listId, name, description, status)
	}
	if rf, ok := ret.Get(0).(func(string, string, string, constants.TodoStatus) *models.Todo); ok {
		r0 = rf(listId, name, description, status)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Todo)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string, constants.TodoStatus) error); ok {
		r1 = rf(listId, name, description, status)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IService_CreateTodoRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTodoRecord'
type IService_CreateTodoRecord_Call struct {
	*mock.Call
}

// CreateTodoRecord is a helper method to define mock.On call
//   - listId string
//   - name string
//   - description string
//   - status constants.TodoStatus
func (_e *IService_Expecter) CreateTodoRecord(listId interface{}, name interface{}, description interface{}, status interface{}) *IService_CreateTodoRecord_Call {
	return &IService_CreateTodoRecord_Call{Call: _e.mock.On("CreateTodoRecord", listId, name, description, status)}
}

func (_c *IService_CreateTodoRecord_Call) Run(run func(listId string, name string, description string, status constants.TodoStatus)) *IService_CreateTodoRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(constants.TodoStatus))
	})
	return _c
}

func (_c *IService_CreateTodoRecord_Call) Return(_a0 *models.Todo, _a1 error) *IService_CreateTodoRecord_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IService_CreateTodoRecord_Call) RunAndReturn(run func(string, string, string, constants.TodoStatus) (*models.Todo, error)) *IService_CreateTodoRecord_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteTodoRecord provides a mock function with given fields: listId, todoId
func (_m *IService) DeleteTodoRecord(listId string, todoId string) error {
	ret := _m.Called(listId, todoId)

	if len(ret) == 0 {
		panic("no return value specified for DeleteTodoRecord")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(listId, todoId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IService_DeleteTodoRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteTodoRecord'
type IService_DeleteTodoRecord_Call struct {
	*mock.Call
}

// DeleteTodoRecord is a helper method to define mock.On call
//   - listId string
//   - todoId string
func (_e *IService_Expecter) DeleteTodoRecord(listId interface{}, todoId interface{}) *IService_DeleteTodoRecord_Call {
	return &IService_DeleteTodoRecord_Call{Call: _e.mock.On("DeleteTodoRecord", listId, todoId)}
}

func (_c *IService_DeleteTodoRecord_Call) Run(run func(listId string, todoId string)) *IService_DeleteTodoRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *IService_DeleteTodoRecord_Call) Return(_a0 error) *IService_DeleteTodoRecord_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IService_DeleteTodoRecord_Call) RunAndReturn(run func(string, string) error) *IService_DeleteTodoRecord_Call {
	_c.Call.Return(run)
	return _c
}

// GetTodoRecord provides a mock function with given fields: listId, todoId
func (_m *IService) GetTodoRecord(listId string, todoId string) (*models.Todo, error) {
	ret := _m.Called(listId, todoId)

	if len(ret) == 0 {
		panic("no return value specified for GetTodoRecord")
	}

	var r0 *models.Todo
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*models.Todo, error)); ok {
		return rf(listId, todoId)
	}
	if rf, ok := ret.Get(0).(func(string, string) *models.Todo); ok {
		r0 = rf(listId, todoId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Todo)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(listId, todoId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IService_GetTodoRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTodoRecord'
type IService_GetTodoRecord_Call struct {
	*mock.Call
}

// GetTodoRecord is a helper method to define mock.On call
//   - listId string
//   - todoId string
func (_e *IService_Expecter) GetTodoRecord(listId interface{}, todoId interface{}) *IService_GetTodoRecord_Call {
	return &IService_GetTodoRecord_Call{Call: _e.mock.On("GetTodoRecord", listId, todoId)}
}

func (_c *IService_GetTodoRecord_Call) Run(run func(listId string, todoId string)) *IService_GetTodoRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *IService_GetTodoRecord_Call) Return(_a0 *models.Todo, _a1 error) *IService_GetTodoRecord_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IService_GetTodoRecord_Call) RunAndReturn(run func(string, string) (*models.Todo, error)) *IService_GetTodoRecord_Call {
	_c.Call.Return(run)
	return _c
}

// GetTodoRecords provides a mock function with given fields: listId
func (_m *IService) GetTodoRecords(listId string) ([]*models.Todo, error) {
	ret := _m.Called(listId)

	if len(ret) == 0 {
		panic("no return value specified for GetTodoRecords")
	}

	var r0 []*models.Todo
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]*models.Todo, error)); ok {
		return rf(listId)
	}
	if rf, ok := ret.Get(0).(func(string) []*models.Todo); ok {
		r0 = rf(listId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Todo)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(listId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IService_GetTodoRecords_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTodoRecords'
type IService_GetTodoRecords_Call struct {
	*mock.Call
}

// GetTodoRecords is a helper method to define mock.On call
//   - listId string
func (_e *IService_Expecter) GetTodoRecords(listId interface{}) *IService_GetTodoRecords_Call {
	return &IService_GetTodoRecords_Call{Call: _e.mock.On("GetTodoRecords", listId)}
}

func (_c *IService_GetTodoRecords_Call) Run(run func(listId string)) *IService_GetTodoRecords_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *IService_GetTodoRecords_Call) Return(_a0 []*models.Todo, _a1 error) *IService_GetTodoRecords_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *IService_GetTodoRecords_Call) RunAndReturn(run func(string) ([]*models.Todo, error)) *IService_GetTodoRecords_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateTodoRecord provides a mock function with given fields: listId, todoId, status
func (_m *IService) UpdateTodoRecord(listId string, todoId string, status constants.TodoStatus) error {
	ret := _m.Called(listId, todoId, status)

	if len(ret) == 0 {
		panic("no return value specified for UpdateTodoRecord")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, constants.TodoStatus) error); ok {
		r0 = rf(listId, todoId, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IService_UpdateTodoRecord_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateTodoRecord'
type IService_UpdateTodoRecord_Call struct {
	*mock.Call
}

// UpdateTodoRecord is a helper method to define mock.On call
//   - listId string
//   - todoId string
//   - status constants.TodoStatus
func (_e *IService_Expecter) UpdateTodoRecord(listId interface{}, todoId interface{}, status interface{}) *IService_UpdateTodoRecord_Call {
	return &IService_UpdateTodoRecord_Call{Call: _e.mock.On("UpdateTodoRecord", listId, todoId, status)}
}

func (_c *IService_UpdateTodoRecord_Call) Run(run func(listId string, todoId string, status constants.TodoStatus)) *IService_UpdateTodoRecord_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(constants.TodoStatus))
	})
	return _c
}

func (_c *IService_UpdateTodoRecord_Call) Return(_a0 error) *IService_UpdateTodoRecord_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *IService_UpdateTodoRecord_Call) RunAndReturn(run func(string, string, constants.TodoStatus) error) *IService_UpdateTodoRecord_Call {
	_c.Call.Return(run)
	return _c
}

// NewIService creates a new instance of IService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIService(t interface {
	mock.TestingT
	Cleanup(func())
}) *IService {
	mock := &IService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
