package client

import (
	"context"
	"time"

	"github.com/whosefriendA/firEtcd/common"
)

// --- 角色 1: 基本 KV 存储 ---
// KVStore 定义了基本的键值对读写操作。
type KVStore interface {
	Put(key string, value []byte, TTL time.Duration) error
	Get(key string) ([]byte, error)
	Append(key string, value []byte, TTL time.Duration) error
	Delete(key string) error
	DeleteWithPrefix(prefix string) error
	CAS(key string, origin, dest []byte, TTL time.Duration) (bool, error)
	GetWithPrefix(key string) ([][]byte, error)
}

// --- 角色 2: 分布式锁 ---
// Locker 定义了获取和释放分布式锁的操作。
type Locker interface {
	Lock(key string, TTL time.Duration) (id string, err error)
	Unlock(key, id string) (bool, error)
}

// --- 角色 3: 批量操作 ---
// Pipeliner 定义了创建操作管道的能力。
type Pipeliner interface {
	// 返回一个可以进行批量操作的 Pipe。
	// Pipe 自身依赖 BatchWriter 接口。
	Pipeline() *Pipe
}

// BatchWriter 是一个内部接口，供 Pipe 使用。
// (这个我们之前已经定义过了)
type BatchWriter interface {
	BatchWrite(p *Pipe) error
}

// --- 角色 4: 变更观察者 ---
// Watcher 定义了观察 key 变化的能力。
type Watcher interface {
	Watch(ctx context.Context, key string, opts ...WatchOption) (<-chan *WatchEvent, error)
}

// --- 角色 5: 高级查询 ---
// Querier 定义了批量查询键和键值对的能力。
type Querier interface {
	Keys() ([]common.Pair, error)
	KVs() ([]common.Pair, error)
	KeysWithPage(pageSize, pageIndex int) ([]common.Pair, error)
	KVsWithPage(pageSize, pageIndex int) ([]common.Pair, error)
}

// --- 静态检查 ---
// 确保 *Clerk 类型实现了我们定义的所有角色接口。
var _ KVStore = (*Clerk)(nil)
var _ Locker = (*Clerk)(nil)
var _ Pipeliner = (*Clerk)(nil)
var _ BatchWriter = (*Clerk)(nil)
var _ Watcher = (*Clerk)(nil)
var _ Querier = (*Clerk)(nil)
