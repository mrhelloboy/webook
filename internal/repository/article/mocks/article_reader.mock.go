// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/article/article_reader.go
//
// Generated by this command:
//
//	mockgen -source=internal/repository/article/article_reader.go -package=repomocks -destination=internal/repository/article/mocks/article_reader.mock.go
//

// Package repomocks is a generated GoMock package.
package repomocks

import (
	context "context"
	reflect "reflect"

	domain "github.com/mrhelloboy/wehook/internal/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockReaderRepository is a mock of ReaderRepository interface.
type MockReaderRepository struct {
	ctrl     *gomock.Controller
	recorder *MockReaderRepositoryMockRecorder
}

// MockReaderRepositoryMockRecorder is the mock recorder for MockReaderRepository.
type MockReaderRepositoryMockRecorder struct {
	mock *MockReaderRepository
}

// NewMockReaderRepository creates a new mock instance.
func NewMockReaderRepository(ctrl *gomock.Controller) *MockReaderRepository {
	mock := &MockReaderRepository{ctrl: ctrl}
	mock.recorder = &MockReaderRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReaderRepository) EXPECT() *MockReaderRepositoryMockRecorder {
	return m.recorder
}

// Save mocks base method.
func (m *MockReaderRepository) Save(ctx context.Context, art domain.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, art)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockReaderRepositoryMockRecorder) Save(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockReaderRepository)(nil).Save), ctx, art)
}
