package raft

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/whosefriendA/firEtcd/pkg/firconfig"
)

// Jepsen-style testing additions start here

// KVCommand defines the operation on the key-value store.
type KVCommand struct {
	Op    string // "Append"
	Key   string
	Value string
}

// HistoryEntry records a client operation and its result.
type HistoryEntry struct {
	Op       KVCommand
	Success  bool
	Start    time.Time
	End      time.Time
	ClientId int
}

// JepsenHarness manages the entire test setup.
type JepsenHarness struct {
	t *testing.T

	n          int // Number of raft peers
	rafts      []*Raft
	persisters []*Persister
	applyChs   []chan ApplyMsg
	kvStores   []map[string]string

	mu         sync.Mutex
	nodeStatus []bool // true if node is alive

	conf firconfig.RaftEnds

	history   []HistoryEntry
	historyMu sync.Mutex

	done chan struct{}
	wg   sync.WaitGroup
}

// NewJepsenHarness creates a new test harness.
func NewJepsenHarness(t *testing.T, n int) *JepsenHarness {
	// Suppress log output from the Raft implementation
	log.SetOutput(os.Stderr)

	h := &JepsenHarness{
		t:          t,
		n:          n,
		rafts:      make([]*Raft, n),
		persisters: make([]*Persister, n),
		applyChs:   make([]chan ApplyMsg, n),
		kvStores:   make([]map[string]string, n),
		nodeStatus: make([]bool, n),
		done:       make(chan struct{}),
	}

	endpoints := make([]firconfig.RaftEnd, n)
	for i := 0; i < n; i++ {
		// Use high port numbers to avoid conflicts
		port := 12345 + i
		endpoints[i] = firconfig.RaftEnd{
			Addr: "127.0.0.1",
			Port: ":" + strconv.Itoa(port),
		}
	}

	for i := 0; i < n; i++ {
		h.conf.Endpoints = endpoints
		h.conf.Me = i
	}

	return h
}

// StartNode (re)starts a single node.
func (h *JepsenHarness) StartNode(nodeId int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.applyChs[nodeId] = make(chan ApplyMsg, 100)
	// Create a new persister or use the existing one to simulate restart
	if h.persisters[nodeId] == nil {
		basePath := fmt.Sprintf("%s/firEtcd_test/%d/", os.TempDir(), nodeId)
		h.persisters[nodeId] = MakePersister("raftstate.bin", "snapshot.bin", basePath)
	}
	h.kvStores[nodeId] = make(map[string]string)

	nodeConf := h.conf
	nodeConf.Me = nodeId

	h.rafts[nodeId] = Make(nodeId, h.persisters[nodeId], h.applyChs[nodeId], nodeConf)
	h.nodeStatus[nodeId] = true

	h.wg.Add(1)
	go h.applyLoop(nodeId)
}

// StopNode kills a single node.
func (h *JepsenHarness) StopNode(nodeId int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.nodeStatus[nodeId] {
		return // Already stopped
	}

	h.rafts[nodeId].Kill()
	h.nodeStatus[nodeId] = false
	// We don't close the applyCh immediately, Shutdown will handle the main done channel
}

// applyLoop simulates the state machine for a single node.
func (h *JepsenHarness) applyLoop(nodeId int) {
	defer h.wg.Done()
	for {
		select {
		case <-h.done:
			return
		case msg := <-h.applyChs[nodeId]:
			if msg.CommandValid {
				var cmd KVCommand
				d := gob.NewDecoder(bytes.NewBuffer(msg.Command))
				if err := d.Decode(&cmd); err == nil {
					h.mu.Lock()
					if cmd.Op == "Append" {
						h.kvStores[nodeId][cmd.Key] += cmd.Value + ";"
					}
					h.mu.Unlock()
				}
			}
		}
	}
}

// StartCluster starts all raft nodes.
func (h *JepsenHarness) StartCluster() {
	for i := 0; i < h.n; i++ {
		h.StartNode(i)
	}
	// Give time for leader election
	time.Sleep(2 * time.Second)
}

// Shutdown kills all raft nodes and waits for goroutines to exit.
func (h *JepsenHarness) Shutdown() {
	close(h.done)
	for i := 0; i < h.n; i++ {
		h.mu.Lock()
		if h.nodeStatus[i] {
			h.rafts[i].Kill()
		}
		h.mu.Unlock()
	}
	h.wg.Wait()
}

// ClientLoop simulates a client sending operations.
func (h *JepsenHarness) ClientLoop(clientId int, key string) {
	h.wg.Add(1)
	defer h.wg.Done()

	opId := 0
	for {
		select {
		case <-h.done:
			return
		default:
			opId++
			value := fmt.Sprintf("%d-%d", clientId, opId)
			cmd := KVCommand{Op: "Append", Key: key, Value: value}

			var cmdBytes bytes.Buffer
			e := gob.NewEncoder(&cmdBytes)
			e.Encode(cmd)

			start := time.Now()
			success := false
			// Try to send the command for a while
			for i := 0; i < 10 && !success; i++ {
				h.mu.Lock()
				// Find a node that is currently alive to send the command to
				leaderHint := rand.Intn(h.n)
				if !h.nodeStatus[leaderHint] {
					h.mu.Unlock()
					time.Sleep(10 * time.Millisecond)
					continue
				}
				rf := h.rafts[leaderHint]
				h.mu.Unlock()

				_, _, isLeader := rf.Start(cmdBytes.Bytes())

				if isLeader {
					success = true
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			end := time.Now()

			h.historyMu.Lock()
			h.history = append(h.history, HistoryEntry{
				Op:       cmd,
				Success:  success,
				Start:    start,
				End:      end,
				ClientId: clientId,
			})
			h.historyMu.Unlock()

			time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
		}
	}
}

// NemesisLoop introduces failures (killing and restarting nodes).
func (h *JepsenHarness) NemesisLoop() {
	h.wg.Add(1)
	defer h.wg.Done()

	for {
		select {
		case <-h.done:
			return
		case <-time.After(time.Duration(500+rand.Intn(500)) * time.Millisecond):
			nodeId := rand.Intn(h.n)
			h.t.Logf("Nemesis: Killing node %d", nodeId)
			h.StopNode(nodeId)

			select {
			case <-h.done:
				return
			case <-time.After(time.Duration(1000+rand.Intn(1000)) * time.Millisecond):
				h.t.Logf("Nemesis: Restarting node %d", nodeId)
				h.StartNode(nodeId)
			}
		}
	}
}

// CheckConsistency verifies the results after the test.
func (h *JepsenHarness) CheckConsistency(key string) {
	h.t.Logf("--- Starting Consistency Check ---")
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. Check that all live nodes have the same state
	var finalState string
	var firstNodeId = -1
	for i := 0; i < h.n; i++ {
		if h.nodeStatus[i] {
			if firstNodeId == -1 {
				firstNodeId = i
				finalState = h.kvStores[i][key]
			} else {
				if h.kvStores[i][key] != finalState {
					h.t.Errorf("Inconsistency detected! Node %d state differs from Node %d.\nNode %d: %s\nNode %d: %s",
						i, firstNodeId, i, h.kvStores[i][key], firstNodeId, finalState)
				}
			}
		}
	}
	if firstNodeId == -1 {
		h.t.Log("All nodes were dead at the end. No state to compare.")
		return
	}
	h.t.Logf("Final consistent state on live nodes: %s", finalState)

	// 2. Check for lost acknowledged writes
	h.historyMu.Lock()
	defer h.historyMu.Unlock()

	acknowledgedWrites := make(map[string]bool)
	for _, entry := range h.history {
		if entry.Success && entry.Op.Op == "Append" {
			acknowledgedWrites[entry.Op.Value] = true
		}
	}

	finalParts := strings.Split(strings.TrimSuffix(finalState, ";"), ";")
	finalStateMap := make(map[string]bool)
	for _, part := range finalParts {
		if part != "" {
			finalStateMap[part] = true
		}
	}

	lostWrites := 0
	for value := range acknowledgedWrites {
		if !finalStateMap[value] {
			h.t.Errorf("LOST WRITE DETECTED: Operation %s was acknowledged but is not in the final state.", value)
			lostWrites++
		}
	}

	if lostWrites == 0 {
		h.t.Logf("PASSED: No lost acknowledged writes. %d successful writes verified.", len(acknowledgedWrites))
	} else {
		h.t.Errorf("FAILED: Detected %d lost writes.", lostWrites)
	}
}

// TestJepsenStyleCrashTolerance is the main entry point for the Jepsen-style test.
func TestJepsenStyleCrashTolerance(t *testing.T) {
	// Set a timeout for the entire test to prevent it from running forever
	testTimeout := time.After(30 * time.Second)
	done := make(chan bool)

	go func() {
		h := NewJepsenHarness(t, 5)
		h.StartCluster()

		// Start a client and the nemesis
		h.wg.Add(1)
		go h.ClientLoop(1, "testKey")
		h.wg.Add(1)
		go h.NemesisLoop()

		// Let the test run for a while
		time.Sleep(25 * time.Second)

		h.t.Log("--- Test duration finished. Shutting down. ---")
		h.Shutdown()
		h.CheckConsistency("testKey")
		done <- true
	}()

	select {
	case <-testTimeout:
		t.Fatal("Test timed out!")
	case <-done:
		// Test finished successfully
	}
}
