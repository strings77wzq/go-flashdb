package cluster

import (
	"fmt"
	"sync"
)

type NodeInfo struct {
	NodeID    string
	IP        string
	Port      int
	Role      string
	Slots     []int
	Connected bool
}

type ClusterManager struct {
	nodes    map[string]*NodeInfo
	selfIP   string
	selfPort int
	slots    map[int]string
	mu       sync.RWMutex
}

func NewClusterManager(ip string, port int) *ClusterManager {
	return &ClusterManager{
		nodes:    make(map[string]*NodeInfo),
		selfIP:   ip,
		selfPort: port,
		slots:    make(map[int]string),
	}
}

func (cm *ClusterManager) AddNode(nodeID, ip string, port int, role string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	nodeKey := fmt.Sprintf("%s:%d", ip, port)
	cm.nodes[nodeKey] = &NodeInfo{
		NodeID:    nodeID,
		IP:        ip,
		Port:      port,
		Role:      role,
		Connected: true,
	}
}

func (cm *ClusterManager) AddSlot(slot int, nodeKey string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.slots[slot] = nodeKey
}

func (cm *ClusterManager) GetNodes() []*NodeInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*NodeInfo, 0, len(cm.nodes))
	for _, node := range cm.nodes {
		result = append(result, node)
	}
	return result
}

func (cm *ClusterManager) GetNodeForSlot(slot int) (string, int, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	nodeKey, ok := cm.slots[slot]
	if !ok {
		return cm.selfIP, cm.selfPort, false
	}
	node, ok := cm.nodes[nodeKey]
	if !ok {
		return cm.selfIP, cm.selfPort, false
	}
	return node.IP, node.Port, true
}

func (cm *ClusterManager) SelfIP() string {
	return cm.selfIP
}

func (cm *ClusterManager) SelfPort() int {
	return cm.selfPort
}
