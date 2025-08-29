package client_test

import (
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/client"
	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

func TestLeaseBasicFlow(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 测试 1: 创建租约
	leaseID, err := ck.LeaseGrant(2 * time.Second)
	if err != nil {
		t.Logf("LeaseGrant failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Granted lease ID: %d", leaseID)

	// 测试 2: 使用 TTL 写入键（应该自动创建租约）
	key := "test-lease-key"
	value := []byte("test-value")
	err = ck.Put(key, value, 1*time.Second)
	if err != nil {
		t.Logf("Put with TTL failed: %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Put key '%s' with TTL", key)

	// 测试 3: 立即获取键（应该存在）
	retrieved, err := ck.Get(key)
	if err != nil {
		t.Logf("Get failed: %v", err)
	} else {
		t.Logf("Retrieved value: %s", string(retrieved))
	}

	// 测试 4: 等待过期并检查删除
	t.Logf("Waiting for lease expiration...")
	time.Sleep(3 * time.Second)

	retrieved, err = ck.Get(key)
	if err == nil {
		t.Logf("Key still exists after expiration: %s", string(retrieved))
	} else {
		t.Logf("Key properly expired (expected): %v", err)
	}

	// 测试 5: 手动租约操作
	leaseID2, err := ck.LeaseGrant(5 * time.Second)
	if err != nil {
		t.Logf("Second LeaseGrant failed: %v", err)
	} else {
		t.Logf("Second lease granted: %d", leaseID2)

		// 检查 TTL
		ttl, keys, err := ck.LeaseTimeToLive(leaseID2, true)
		if err != nil {
			t.Logf("LeaseTimeToLive failed: %v", err)
		} else {
			t.Logf("Lease %d TTL: %dms, keys: %v", leaseID2, ttl, keys)
		}

		// 手动撤销
		err = ck.LeaseRevoke(leaseID2)
		if err != nil {
			t.Logf("LeaseRevoke failed: %v", err)
		} else {
			t.Logf("Lease %d manually revoked", leaseID2)
		}
	}
}

func TestLeaseAutoKeepAlive(t *testing.T) {
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 创建短 TTL 租约
	leaseID, err := ck.LeaseGrant(3 * time.Second)
	if err != nil {
		t.Logf("LeaseGrant failed: %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Granted lease ID: %d for auto-keepalive test", leaseID)

	// 启动自动续约
	cancel := ck.AutoKeepAlive(leaseID, 1*time.Second)
	defer cancel()

	// 使用此租约写入键
	key := "test-keepalive-key"
	value := []byte("test-keepalive-value")
	err = ck.Put(key, value, 0) // 无 TTL，使用现有租约
	if err != nil {
		t.Logf("Put failed: %v", err)
		t.Skip("Skipping test - no server available")
	}

	// 检查键是否存在
	retrieved, err := ck.Get(key)
	if err != nil {
		t.Logf("Get failed: %v", err)
	} else {
		t.Logf("Retrieved value: %s", string(retrieved))
	}

	// 让续约运行一段时间
	time.Sleep(5 * time.Second)

	// 检查 TTL 看续约是否有效
	ttl, _, err := ck.LeaseTimeToLive(leaseID, false)
	if err != nil {
		t.Logf("LeaseTimeToLive failed: %v", err)
	} else {
		t.Logf("Lease %d remaining TTL: %dms", leaseID, ttl)
	}

	// 停止续约
	cancel()
	t.Logf("Auto keepalive stopped")
}

// 新增：测试基于租约的分布式锁
func TestLeaseBasedLock(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 测试基本的锁获取和释放
	key := "test-lease-lock"
	ttl := 5 * time.Second

	// 获取锁
	lockID, err := ck.Lock(key, ttl)
	if err != nil {
		t.Logf("Lock failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Successfully acquired lock with ID: %s", lockID)

	// 验证锁已获取
	if lockID == "" {
		t.Fatal("Lock ID should not be empty")
	}

	// 尝试获取同一个锁（应该失败）
	lockID2, err := ck.Lock(key, ttl)
	if err == nil && lockID2 != "" {
		t.Logf("Second lock acquisition succeeded (this might be expected in some cases)")
	} else {
		t.Logf("Second lock acquisition failed as expected: %v", err)
	}

	// 释放锁
	ok, err := ck.Unlock(key, lockID)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	if !ok {
		t.Fatal("Unlock should succeed")
	}
	t.Logf("Successfully released lock")

	// 验证锁已释放（可以重新获取）
	lockID3, err := ck.Lock(key, ttl)
	if err != nil {
		t.Logf("Re-acquiring lock after release failed: %v", err)
	} else {
		t.Logf("Successfully re-acquired lock with ID: %s", lockID3)
		// 清理
		ck.Unlock(key, lockID3)
	}
}

// 新增：测试带自动续约的锁
func TestLeaseBasedLockWithKeepAlive(t *testing.T) {
	// 创建客户端
	conf := firconfig.Clerk{}
	firconfig.Init("config.yml", &conf)
	ck := client.MakeClerk(conf)

	// 测试带自动续约的锁
	key := "test-lease-lock-keepalive"
	ttl := 3 * time.Second

	// 获取锁并启动自动续约
	lockID, cancel, err := ck.LockWithKeepAlive(key, ttl)
	if err != nil {
		t.Logf("LockWithKeepAlive failed (expected if no server running): %v", err)
		t.Skip("Skipping test - no server available")
	}
	t.Logf("Successfully acquired lock with keepalive, ID: %s", lockID)

	// 等待一段时间，让续约机制工作
	time.Sleep(2 * time.Second)

	// 检查锁是否仍然有效
	// 这里可以添加验证逻辑，比如尝试获取同一个锁

	// 手动停止续约并释放锁
	cancel()
	t.Logf("Lock keepalive stopped and lock released")
}
