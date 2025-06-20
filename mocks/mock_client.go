// Code generated by MockGen. DO NOT EDIT.
// Source: client/interface.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	client "github.com/whosefriendA/firEtcd/client"
	common "github.com/whosefriendA/firEtcd/common"
)

// MockKVStore is a mock of KVStore interface.
type MockKVStore struct {
	ctrl     *gomock.Controller
	recorder *MockKVStoreMockRecorder
}

// MockKVStoreMockRecorder is the mock recorder for MockKVStore.
type MockKVStoreMockRecorder struct {
	mock *MockKVStore
}

// NewMockKVStore creates a new mock instance.
func NewMockKVStore(ctrl *gomock.Controller) *MockKVStore {
	mock := &MockKVStore{ctrl: ctrl}
	mock.recorder = &MockKVStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKVStore) EXPECT() *MockKVStoreMockRecorder {
	return m.recorder
}

// Append mocks base method.
func (m *MockKVStore) Append(key string, value []byte, TTL time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Append", key, value, TTL)
	ret0, _ := ret[0].(error)
	return ret0
}

// Append indicates an expected call of Append.
func (mr *MockKVStoreMockRecorder) Append(key, value, TTL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Append", reflect.TypeOf((*MockKVStore)(nil).Append), key, value, TTL)
}

// CAS mocks base method.
func (m *MockKVStore) CAS(key string, origin, dest []byte, TTL time.Duration) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CAS", key, origin, dest, TTL)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CAS indicates an expected call of CAS.
func (mr *MockKVStoreMockRecorder) CAS(key, origin, dest, TTL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CAS", reflect.TypeOf((*MockKVStore)(nil).CAS), key, origin, dest, TTL)
}

// Delete mocks base method.
func (m *MockKVStore) Delete(key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", key)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockKVStoreMockRecorder) Delete(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockKVStore)(nil).Delete), key)
}

// DeleteWithPrefix mocks base method.
func (m *MockKVStore) DeleteWithPrefix(prefix string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteWithPrefix", prefix)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteWithPrefix indicates an expected call of DeleteWithPrefix.
func (mr *MockKVStoreMockRecorder) DeleteWithPrefix(prefix interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteWithPrefix", reflect.TypeOf((*MockKVStore)(nil).DeleteWithPrefix), prefix)
}

// Get mocks base method.
func (m *MockKVStore) Get(key string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", key)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockKVStoreMockRecorder) Get(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockKVStore)(nil).Get), key)
}

// GetWithPrefix mocks base method.
func (m *MockKVStore) GetWithPrefix(key string) ([][]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithPrefix", key)
	ret0, _ := ret[0].([][]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithPrefix indicates an expected call of GetWithPrefix.
func (mr *MockKVStoreMockRecorder) GetWithPrefix(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithPrefix", reflect.TypeOf((*MockKVStore)(nil).GetWithPrefix), key)
}

// Put mocks base method.
func (m *MockKVStore) Put(key string, value []byte, TTL time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", key, value, TTL)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockKVStoreMockRecorder) Put(key, value, TTL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockKVStore)(nil).Put), key, value, TTL)
}

// MockLocker is a mock of Locker interface.
type MockLocker struct {
	ctrl     *gomock.Controller
	recorder *MockLockerMockRecorder
}

// MockLockerMockRecorder is the mock recorder for MockLocker.
type MockLockerMockRecorder struct {
	mock *MockLocker
}

// NewMockLocker creates a new mock instance.
func NewMockLocker(ctrl *gomock.Controller) *MockLocker {
	mock := &MockLocker{ctrl: ctrl}
	mock.recorder = &MockLockerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLocker) EXPECT() *MockLockerMockRecorder {
	return m.recorder
}

// Lock mocks base method.
func (m *MockLocker) Lock(key string, TTL time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Lock", key, TTL)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Lock indicates an expected call of Lock.
func (mr *MockLockerMockRecorder) Lock(key, TTL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockLocker)(nil).Lock), key, TTL)
}

// Unlock mocks base method.
func (m *MockLocker) Unlock(key, id string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unlock", key, id)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Unlock indicates an expected call of Unlock.
func (mr *MockLockerMockRecorder) Unlock(key, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockLocker)(nil).Unlock), key, id)
}

// MockPipeliner is a mock of Pipeliner interface.
type MockPipeliner struct {
	ctrl     *gomock.Controller
	recorder *MockPipelinerMockRecorder
}

// MockPipelinerMockRecorder is the mock recorder for MockPipeliner.
type MockPipelinerMockRecorder struct {
	mock *MockPipeliner
}

// NewMockPipeliner creates a new mock instance.
func NewMockPipeliner(ctrl *gomock.Controller) *MockPipeliner {
	mock := &MockPipeliner{ctrl: ctrl}
	mock.recorder = &MockPipelinerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPipeliner) EXPECT() *MockPipelinerMockRecorder {
	return m.recorder
}

// Pipeline mocks base method.
func (m *MockPipeliner) Pipeline() *client.Pipe {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pipeline")
	ret0, _ := ret[0].(*client.Pipe)
	return ret0
}

// Pipeline indicates an expected call of Pipeline.
func (mr *MockPipelinerMockRecorder) Pipeline() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pipeline", reflect.TypeOf((*MockPipeliner)(nil).Pipeline))
}

// MockBatchWriter is a mock of BatchWriter interface.
type MockBatchWriter struct {
	ctrl     *gomock.Controller
	recorder *MockBatchWriterMockRecorder
}

// MockBatchWriterMockRecorder is the mock recorder for MockBatchWriter.
type MockBatchWriterMockRecorder struct {
	mock *MockBatchWriter
}

// NewMockBatchWriter creates a new mock instance.
func NewMockBatchWriter(ctrl *gomock.Controller) *MockBatchWriter {
	mock := &MockBatchWriter{ctrl: ctrl}
	mock.recorder = &MockBatchWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBatchWriter) EXPECT() *MockBatchWriterMockRecorder {
	return m.recorder
}

// BatchWrite mocks base method.
func (m *MockBatchWriter) BatchWrite(p *client.Pipe) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchWrite", p)
	ret0, _ := ret[0].(error)
	return ret0
}

// BatchWrite indicates an expected call of BatchWrite.
func (mr *MockBatchWriterMockRecorder) BatchWrite(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchWrite", reflect.TypeOf((*MockBatchWriter)(nil).BatchWrite), p)
}

// MockWatcher is a mock of Watcher interface.
type MockWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockWatcherMockRecorder
}

// MockWatcherMockRecorder is the mock recorder for MockWatcher.
type MockWatcherMockRecorder struct {
	mock *MockWatcher
}

// NewMockWatcher creates a new mock instance.
func NewMockWatcher(ctrl *gomock.Controller) *MockWatcher {
	mock := &MockWatcher{ctrl: ctrl}
	mock.recorder = &MockWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatcher) EXPECT() *MockWatcherMockRecorder {
	return m.recorder
}

// Watch mocks base method.
func (m *MockWatcher) Watch(ctx context.Context, key string, opts ...client.WatchOption) (<-chan *client.WatchEvent, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, key}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Watch", varargs...)
	ret0, _ := ret[0].(<-chan *client.WatchEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Watch indicates an expected call of Watch.
func (mr *MockWatcherMockRecorder) Watch(ctx, key interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, key}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockWatcher)(nil).Watch), varargs...)
}

// MockQuerier is a mock of Querier interface.
type MockQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockQuerierMockRecorder
}

// MockQuerierMockRecorder is the mock recorder for MockQuerier.
type MockQuerierMockRecorder struct {
	mock *MockQuerier
}

// NewMockQuerier creates a new mock instance.
func NewMockQuerier(ctrl *gomock.Controller) *MockQuerier {
	mock := &MockQuerier{ctrl: ctrl}
	mock.recorder = &MockQuerierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuerier) EXPECT() *MockQuerierMockRecorder {
	return m.recorder
}

// KVs mocks base method.
func (m *MockQuerier) KVs() ([]common.Pair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KVs")
	ret0, _ := ret[0].([]common.Pair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// KVs indicates an expected call of KVs.
func (mr *MockQuerierMockRecorder) KVs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KVs", reflect.TypeOf((*MockQuerier)(nil).KVs))
}

// KVsWithPage mocks base method.
func (m *MockQuerier) KVsWithPage(pageSize, pageIndex int) ([]common.Pair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KVsWithPage", pageSize, pageIndex)
	ret0, _ := ret[0].([]common.Pair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// KVsWithPage indicates an expected call of KVsWithPage.
func (mr *MockQuerierMockRecorder) KVsWithPage(pageSize, pageIndex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KVsWithPage", reflect.TypeOf((*MockQuerier)(nil).KVsWithPage), pageSize, pageIndex)
}

// Keys mocks base method.
func (m *MockQuerier) Keys() ([]common.Pair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Keys")
	ret0, _ := ret[0].([]common.Pair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Keys indicates an expected call of Keys.
func (mr *MockQuerierMockRecorder) Keys() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Keys", reflect.TypeOf((*MockQuerier)(nil).Keys))
}

// KeysWithPage mocks base method.
func (m *MockQuerier) KeysWithPage(pageSize, pageIndex int) ([]common.Pair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KeysWithPage", pageSize, pageIndex)
	ret0, _ := ret[0].([]common.Pair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// KeysWithPage indicates an expected call of KeysWithPage.
func (mr *MockQuerierMockRecorder) KeysWithPage(pageSize, pageIndex interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeysWithPage", reflect.TypeOf((*MockQuerier)(nil).KeysWithPage), pageSize, pageIndex)
}

// MockWatchDoger is a mock of WatchDoger interface.
type MockWatchDoger struct {
	ctrl     *gomock.Controller
	recorder *MockWatchDogerMockRecorder
}

// MockWatchDogerMockRecorder is the mock recorder for MockWatchDoger.
type MockWatchDogerMockRecorder struct {
	mock *MockWatchDoger
}

// NewMockWatchDoger creates a new mock instance.
func NewMockWatchDoger(ctrl *gomock.Controller) *MockWatchDoger {
	mock := &MockWatchDoger{ctrl: ctrl}
	mock.recorder = &MockWatchDogerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatchDoger) EXPECT() *MockWatchDogerMockRecorder {
	return m.recorder
}

// WatchDog mocks base method.
func (m *MockWatchDoger) WatchDog(key string, value []byte) func() {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchDog", key, value)
	ret0, _ := ret[0].(func())
	return ret0
}

// WatchDog indicates an expected call of WatchDog.
func (mr *MockWatchDogerMockRecorder) WatchDog(key, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchDog", reflect.TypeOf((*MockWatchDoger)(nil).WatchDog), key, value)
}
