package client_test // 包名修改

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	// !!! 确认 client 和 raft 包的导入路径正确
	"github.com/whosefriendA/firEtcd/client" // <-- 新增导入
	"github.com/whosefriendA/firEtcd/mocks"
	"github.com/whosefriendA/firEtcd/raft"
)

// --- 自定义 Gomock 匹配器 ---

// pipeMatcher 用于验证传入的 *client.Pipe 对象的 ops 字段。
// 注意：这个匹配器能成功的前提是 Pipe 结构体中的 ops 字段是公开的 (Ops)。
// 如果 ops 是非公开的，你需要一个公开的 getter 方法 (e.g., p.Ops()) 并在匹配器中调用它。
// 这里我们假设 ops 是公开的。
type pipeMatcher struct {
	expectedOps []raft.Op
}

// Matches 是 gomock.Matcher 接口的核心方法。
func (m *pipeMatcher) Matches(x interface{}) bool {
	// 类型断言修改为 *client.Pipe
	p, ok := x.(*client.Pipe)
	if !ok {
		return false
	}
	// 假设 client.Pipe 有一个公开的 `Ops` 字段或 `Ops()` 方法
	// 如果字段名是小写的 `ops`，你需要一个公开的 getter，例如：
	// return reflect.DeepEqual(p.GetOps(), m.expectedOps)
	// 这里我们假设字段是公开的 `Ops`。
	return reflect.DeepEqual(p.Ops, m.expectedOps)
}

// String 方法用于在测试失败时提供清晰的描述。
func (m *pipeMatcher) String() string {
	// 描述信息更新
	return fmt.Sprintf("is a *client.Pipe with ops matching: %v", m.expectedOps)
}

// PipeWithOps 是一个辅助函数，让创建匹配器更方便。
func PipeWithOps(ops []raft.Op) gomock.Matcher {
	return &pipeMatcher{expectedOps: ops}
}

func TestPipe_Exec_WithMock(t *testing.T) {
	t.Run("Exec with ops and successful clerk", func(t *testing.T) {
		// 1. Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCk := mocks.NewMockBatchWriter(ctrl)
		// !!! 使用构造函数创建 Pipe 实例，因为无法从外部访问非公开字段 ck
		p := client.NewPipe(mockCk)

		p.Put("k1", []byte("v1"), 0)
		p.Delete("k2")

		// 假设 Pipe 有一个公开的方法来获取其内部的 ops
		// 如果没有，此部分逻辑可能需要调整
		expectedOps := p.Ops // 假设 `Ops` 字段是公开的

		// 2. 设定预期 (使用我们自定义的匹配器)
		mockCk.EXPECT().
			BatchWrite(PipeWithOps(expectedOps)). // 使用自定义匹配器
			Return(nil).
			Times(1)

		// 3. Action
		err := p.Exec()

		// 4. Assert
		if err != nil {
			t.Fatalf("Pipe.Exec() returned unexpected error: %v", err)
		}
	})

	t.Run("Exec with ops and failing clerk", func(t *testing.T) {
		// 1. Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCk := mocks.NewMockBatchWriter(ctrl)
		p := client.NewPipe(mockCk) // 使用构造函数
		p.Put("k1", []byte("v1"), 0)

		// 2. 设定预期
		expectedErr := errors.New("simulated clerk failure")
		mockCk.EXPECT().BatchWrite(gomock.Any()).Return(expectedErr).Times(1)

		// 3. Action
		err := p.Exec()

		// 4. Assert
		if err == nil {
			t.Fatal("Pipe.Exec() should have returned an error, but did not")
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})

	t.Run("Exec with no ops", func(t *testing.T) {
		// 1. Setup
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCk := mocks.NewMockBatchWriter(ctrl)
		p := client.NewPipe(mockCk) // 使用构造函数

		// 2. 设定预期: 无。
		// 由于没有为 mockCk 设置任何 EXPECT()，任何对它的方法调用都会导致测试失败。

		// 3. Action
		err := p.Exec()

		// 4. Assert
		if err != nil {
			t.Fatalf("Pipe.Exec() with no ops returned an error: %v", err)
		}
	})
}
