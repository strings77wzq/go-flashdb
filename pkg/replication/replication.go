package replication

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
	"time"
)

type Role int

const (
	RoleMaster Role = iota
	RoleSlave
)

func (r Role) String() string {
	if r == RoleMaster {
		return "master"
	}
	return "slave"
}

type SlaveState int

const (
	SlaveStateNone SlaveState = iota
	SlaveStateConnecting
	SlaveStateSyncing
	SlaveStateOnline
)

func (s SlaveState) String() string {
	switch s {
	case SlaveStateConnecting:
		return "connect"
	case SlaveStateSyncing:
		return "sync"
	case SlaveStateOnline:
		return "online"
	default:
		return "none"
	}
}

type ReplicationManager struct {
	mu         sync.RWMutex
	role       Role
	replID     string
	replOffset int64

	masterInfo *MasterInfo
	slaveInfo  *SlaveInfo

	commandBuffer *CommandBuffer
}

type MasterInfo struct {
	slaves      map[string]*SlaveConn
	replOffset  int64
	replBacklog []byte
}

type SlaveInfo struct {
	masterHost string
	masterPort int
	masterConn net.Conn
	state      SlaveState
	replOffset int64
}

type SlaveConn struct {
	conn       net.Conn
	offset     int64
	lastAck    time.Time
	remoteAddr string
}

type CommandBuffer struct {
	mu       sync.RWMutex
	commands [][]byte
	offsets  []int64
	maxSize  int
}

func NewCommandBuffer(maxSize int) *CommandBuffer {
	return &CommandBuffer{
		commands: make([][]byte, 0),
		offsets:  make([]int64, 0),
		maxSize:  maxSize,
	}
}

func (cb *CommandBuffer) Add(command []byte, offset int64) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.commands = append(cb.commands, command)
	cb.offsets = append(cb.offsets, offset)

	if len(cb.commands) > cb.maxSize {
		cb.commands = cb.commands[1:]
		cb.offsets = cb.offsets[1:]
	}
}

func (cb *CommandBuffer) GetSince(offset int64) ([][]byte, int64) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	var result [][]byte
	var lastOffset int64

	for i, off := range cb.offsets {
		if off > offset {
			result = append(result, cb.commands[i])
			lastOffset = off
		}
	}

	return result, lastOffset
}

func NewReplicationManager() *ReplicationManager {
	replID := generateReplID()
	return &ReplicationManager{
		role:          RoleMaster,
		replID:        replID,
		replOffset:    0,
		commandBuffer: NewCommandBuffer(10000),
		masterInfo: &MasterInfo{
			slaves: make(map[string]*SlaveConn),
		},
	}
}

func generateReplID() string {
	b := make([]byte, 20)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (rm *ReplicationManager) Role() Role {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.role
}

func (rm *ReplicationManager) SetRole(role Role) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.role = role
}

func (rm *ReplicationManager) GetReplID() string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.replID
}

func (rm *ReplicationManager) GetReplOffset() int64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.replOffset
}

func (rm *ReplicationManager) IncrementOffset(delta int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.replOffset += delta
}

func (rm *ReplicationManager) AddCommand(command []byte) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.replOffset++
	rm.commandBuffer.Add(command, rm.replOffset)
}

func (rm *ReplicationManager) ReplicaOf(host string, port int) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if host == "no" && port == 0 || host == "NO" && port == 0 {
		rm.role = RoleMaster
		rm.slaveInfo = nil
		rm.replID = generateReplID()
		rm.replOffset = 0
		return nil
	}

	rm.role = RoleSlave
	rm.slaveInfo = &SlaveInfo{
		masterHost: host,
		masterPort: port,
		state:      SlaveStateConnecting,
	}

	return nil
}

func (rm *ReplicationManager) GetMasterInfo() (string, int) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.slaveInfo == nil {
		return "", 0
	}
	return rm.slaveInfo.masterHost, rm.slaveInfo.masterPort
}

func (rm *ReplicationManager) GetSlaveState() SlaveState {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.slaveInfo == nil {
		return SlaveStateNone
	}
	return rm.slaveInfo.state
}

func (rm *ReplicationManager) SetSlaveState(state SlaveState) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.slaveInfo != nil {
		rm.slaveInfo.state = state
	}
}

func (rm *ReplicationManager) AddSlave(conn net.Conn) string {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	addr := conn.RemoteAddr().String()
	rm.masterInfo.slaves[addr] = &SlaveConn{
		conn:       conn,
		offset:     0,
		lastAck:    time.Now(),
		remoteAddr: addr,
	}

	return addr
}

func (rm *ReplicationManager) RemoveSlave(addr string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.masterInfo.slaves, addr)
}

func (rm *ReplicationManager) GetSlaveCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.masterInfo.slaves)
}

func (rm *ReplicationManager) GetConnectedSlaves() []string {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var slaves []string
	for addr := range rm.masterInfo.slaves {
		slaves = append(slaves, addr)
	}
	return slaves
}

func (rm *ReplicationManager) UpdateSlaveOffset(addr string, offset int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if slave, ok := rm.masterInfo.slaves[addr]; ok {
		slave.offset = offset
		slave.lastAck = time.Now()
	}
}

func (rm *ReplicationManager) GetCommandsSince(offset int64) ([][]byte, int64) {
	return rm.commandBuffer.GetSince(offset)
}

func (rm *ReplicationManager) GetInfo() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	info := map[string]interface{}{
		"role":               rm.role.String(),
		"connected_slaves":   len(rm.masterInfo.slaves),
		"master_replid":      rm.replID,
		"master_repl_offset": rm.replOffset,
	}

	if rm.role == RoleSlave && rm.slaveInfo != nil {
		info["master_host"] = rm.slaveInfo.masterHost
		info["master_port"] = rm.slaveInfo.masterPort
		info["master_link_status"] = rm.slaveInfo.state.String()
	}

	return info
}
