package lease

import (
	"sync"
	"time"
)

// Lease represents a single lease.
type Lease struct {
	ID        int64
	TTL       time.Duration
	ExpiresAt time.Time
	Keys      map[string]struct{} // Set of keys associated with this lease
	mu        sync.Mutex
}

// LeaseManager manages leases and their associated keys.
type LeaseManager struct {
	sync.RWMutex
	leases      map[int64]*Lease
	nextLeaseID int64
	minLeaseTTL time.Duration
	revokeC     chan int64 // Channel to signal lease revocation
	stopC       chan struct{}
	wg          sync.WaitGroup
}

// NewLeaseManager creates a new LeaseManager.
func NewLeaseManager(minLeaseTTL time.Duration) *LeaseManager {
	lm := &LeaseManager{
		leases:      make(map[int64]*Lease),
		nextLeaseID: 1,
		minLeaseTTL: minLeaseTTL,
		revokeC:     make(chan int64),
		stopC:       make(chan struct{}),
	}
	lm.wg.Add(1)
	go lm.run()
	return lm
}

// run is the main loop for the LeaseManager.
func (lm *LeaseManager) run() {
	defer lm.wg.Done()

	ticker := time.NewTicker(lm.minLeaseTTL / 2) // Check for expirations more frequently
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lm.checkExpirations()
		case leaseID := <-lm.revokeC:
			lm.revokeLease(leaseID)
		case <-lm.stopC:
			return
		}
	}
}

// Grant creates a new lease with the given TTL.
func (lm *LeaseManager) Grant(ttl time.Duration) (*Lease, error) {
	lm.Lock()
	defer lm.Unlock()

	if ttl < lm.minLeaseTTL {
		ttl = lm.minLeaseTTL
	}

	lease := &Lease{
		ID:        lm.nextLeaseID,
		TTL:       ttl,
		ExpiresAt: time.Now().Add(ttl),
		Keys:      make(map[string]struct{}),
	}
	lm.leases[lease.ID] = lease
	lm.nextLeaseID++

	// TODO: Persist lease state to Raft

	return lease, nil
}

// Revoke revokes a lease and deletes all associated keys.
func (lm *LeaseManager) Revoke(id int64) error {
	lm.revokeC <- id
	return nil
}

// revokeLease performs the actual lease revocation.
func (lm *LeaseManager) revokeLease(id int64) {
	lm.Lock()
	defer lm.Unlock()

	_, ok := lm.leases[id]
	if !ok {
		return // Lease not found
	}

	delete(lm.leases, id)

	// TODO: Delete all keys associated with this lease
	// TODO: Persist lease state to Raft
}

// KeepAlive renews a lease.
func (lm *LeaseManager) KeepAlive(id int64) (*Lease, error) {
	lm.Lock()
	defer lm.Unlock()

	lease, ok := lm.leases[id]
	if !ok {
		return nil, nil // Lease not found
	}

	lease.ExpiresAt = time.Now().Add(lease.TTL)

	// TODO: Persist lease state to Raft

	return lease, nil
}

// AttachKey attaches a key to a lease.
func (lm *LeaseManager) AttachKey(leaseID int64, key string) error {
	lm.Lock()
	defer lm.Unlock()

	lease, ok := lm.leases[leaseID]
	if !ok {
		return nil // Lease not found
	}

	lease.mu.Lock()
	lease.Keys[key] = struct{}{}
	lease.mu.Unlock()

	// TODO: Persist key-lease association to Raft

	return nil
}

// DetachKey detaches a key from a lease.
func (lm *LeaseManager) DetachKey(leaseID int64, key string) error {
	lm.Lock()
	defer lm.Unlock()

	lease, ok := lm.leases[leaseID]
	if !ok {
		return nil // Lease not found
	}

	lease.mu.Lock()
	delete(lease.Keys, key)
	lease.mu.Unlock()

	// TODO: Persist key-lease association to Raft

	return nil
}

// checkExpirations checks for expired leases and revokes them.
func (lm *LeaseManager) checkExpirations() {
	lm.Lock()
	defer lm.Unlock()

	now := time.Now()
	for id, lease := range lm.leases {
		if now.After(lease.ExpiresAt) {
			lm.revokeC <- id
		}
	}
}

// ExpiredLeases returns IDs of leases expired at the given time (does not modify state).
func (lm *LeaseManager) ExpiredLeases(now time.Time) []int64 {
	lm.RLock()
	defer lm.RUnlock()
	ret := make([]int64, 0)
	for id, l := range lm.leases {
		if now.After(l.ExpiresAt) {
			ret = append(ret, id)
		}
	}
	return ret
}

// Stop stops the LeaseManager.
func (lm *LeaseManager) Stop() {
	close(lm.stopC)
	lm.wg.Wait()
}

// GetLease returns a lease by ID.
func (lm *LeaseManager) GetLease(id int64) *Lease {
	lm.RLock()
	defer lm.RUnlock()
	return lm.leases[id]
}

// GetLeaseKeys returns the keys associated with a lease.
func (lm *LeaseManager) GetLeaseKeys(id int64) []string {
	lm.RLock()
	defer lm.RUnlock()

	lease, ok := lm.leases[id]
	if !ok {
		return nil
	}

	lease.mu.Lock()
	defer lease.mu.Unlock()

	keys := make([]string, 0, len(lease.Keys))
	for key := range lease.Keys {
		keys = append(keys, key)
	}
	return keys
}
