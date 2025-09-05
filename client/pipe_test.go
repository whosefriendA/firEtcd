package client_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/mocks"
	"github.com/whosefriendA/firEtcd/raft"
)

type pipeMatcher struct {
	expectedOps []raft.Op
}

func (m *pipeMatcher) Matches(x interface{}) bool {
	p, ok := x.(*client.Pipe)
	if !ok {
		return false
	}
	return reflect.DeepEqual(p.Ops, m.expectedOps)
}

func (m *pipeMatcher) String() string {
	return fmt.Sprintf("is a *client.Pipe with ops matching: %v", m.expectedOps)
}

func PipeWithOps(ops []raft.Op) gomock.Matcher {
	return &pipeMatcher{expectedOps: ops}
}

func TestPipe_Exec_WithMock(t *testing.T) {
	t.Run("Exec with ops and successful clerk", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCk := mocks.NewMockBatchWriter(ctrl)
		p := client.NewPipe(mockCk)

		p.Put("k1", []byte("v1"), 0)
		p.Delete("k2")

		expectedOps := p.Ops

		mockCk.EXPECT().
			BatchWrite(PipeWithOps(expectedOps)).
			Return(nil).
			Times(1)

		err := p.Exec()

		if err != nil {
			t.Fatalf("Pipe.Exec() returned unexpected error: %v", err)
		}
	})

	t.Run("Exec with ops and failing clerk", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCk := mocks.NewMockBatchWriter(ctrl)
		p := client.NewPipe(mockCk)
		p.Put("k1", []byte("v1"), 0)

		expectedErr := errors.New("simulated clerk failure")
		mockCk.EXPECT().BatchWrite(gomock.Any()).Return(expectedErr).Times(1)

		err := p.Exec()

		if err == nil {
			t.Fatal("Pipe.Exec() should have returned an error, but did not")
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})

	t.Run("Exec with no ops", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCk := mocks.NewMockBatchWriter(ctrl)
		p := client.NewPipe(mockCk)

		err := p.Exec()

		if err != nil {
			t.Fatalf("Pipe.Exec() with no ops returned an error: %v", err)
		}
	})
}
