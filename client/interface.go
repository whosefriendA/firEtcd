package client

//go:generate mockgen -source=interface.go -destination=../mocks/mock_client.go -package=mocks

import (
	"context"
	"time"

	"github.com/whosefriendA/firEtcd/common"
)

// KVStore 定义了基本的键值对读写操作。
type KVStore interface {
	Put(key string, value []byte, TTL time.Duration) error
	Get(key string) ([]byte, error)
	Append(key string, value []byte, TTL time.Duration) error
	Delete(key string) error
	DeleteWithPrefix(prefix string) error
	CAS(key string, origin, dest []byte, TTL time.Duration) (bool, error)
	GetWithPrefix(key string) ([][]byte, error)
	// Lease operations
	LeaseGrant(ttl time.Duration) (int64, error)
	LeaseRevoke(leaseID int64) error
	LeaseTimeToLive(leaseID int64, withKeys bool) (int64, []string, error)
	AutoKeepAlive(leaseID int64, interval time.Duration) (cancel func())
}

// Locker 定义了获取和释放分布式锁的操作。
type Locker interface {
	Lock(key string, TTL time.Duration) (id string, err error)
	Unlock(key, id string) (bool, error)
	LockWithKeepAlive(key string, TTL time.Duration) (id string, cancel func(), err error)
	// Lease operations for lock management
	LeaseGrant(ttl time.Duration) (int64, error)
	LeaseRevoke(leaseID int64) error
	LeaseTimeToLive(leaseID int64, withKeys bool) (int64, []string, error)
	AutoKeepAlive(leaseID int64, interval time.Duration) (cancel func())
}

// Pipeliner 定义了创建操作管道的能力。
type Pipeliner interface {
	// 返回一个可以进行批量操作的 Pipe。
	// Pipe 自身依赖 BatchWriter 接口。
	Pipeline() *Pipe
}

// BatchWriter 是一个内部接口，供 Pipe 使用。
type BatchWriter interface {
	BatchWrite(p *Pipe) error
}

// Watcher 定义了观察 key 变化的能力。
type Watcher interface {
	Watch(ctx context.Context, key string, opts ...WatchOption) (<-chan *WatchEvent, error)
}

// Querier 定义了批量查询键和键值对的能力。
type Querier interface {
	Keys() ([]common.Pair, error)
	KVs() ([]common.Pair, error)
	KeysWithPage(pageSize, pageIndex int) ([]common.Pair, error)
	KVsWithPage(pageSize, pageIndex int) ([]common.Pair, error)
}

var _ KVStore = (*Client)(nil)
var _ Locker = (*Client)(nil)
var _ Pipeliner = (*Client)(nil)
var _ BatchWriter = (*Client)(nil)
var _ Watcher = (*Client)(nil)
var _ Querier = (*Client)(nil)
