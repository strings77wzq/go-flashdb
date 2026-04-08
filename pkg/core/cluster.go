package core

import (
	"fmt"
	"strings"

	"goflashdb/pkg/cluster"
	"goflashdb/pkg/resp"
)

var clusterMgr *cluster.ClusterManager

func getClusterManager() *cluster.ClusterManager {
	if clusterMgr == nil {
		clusterMgr = cluster.NewClusterManager("127.0.0.1", 6379)
	}
	return clusterMgr
}

func execClusterSlots(db *DB, args [][]byte) resp.Reply {
	cm := getClusterManager()
	nodes := cm.GetNodes()

	var replies []resp.Reply

	if len(nodes) == 0 {
		slotRange := []resp.Reply{
			&resp.IntegerReply{Num: 0},
			&resp.IntegerReply{Num: 16383},
			&resp.BulkReply{Arg: []byte(cm.SelfIP())},
			&resp.IntegerReply{Num: int64(cm.SelfPort())},
		}
		replies = append(replies, &resp.ArrayReply{Replies: slotRange})
	} else {
		slotRange := []resp.Reply{
			&resp.IntegerReply{Num: 0},
			&resp.IntegerReply{Num: 16383},
			&resp.BulkReply{Arg: []byte(nodes[0].IP)},
			&resp.IntegerReply{Num: int64(nodes[0].Port)},
		}
		replies = append(replies, &resp.ArrayReply{Replies: slotRange})
	}

	return &resp.ArrayReply{Replies: replies}
}

func execClusterInfo(db *DB, args [][]byte) resp.Reply {
	cm := getClusterManager()
	nodes := cm.GetNodes()

	lines := []string{
		"cluster_state:ok",
		"cluster_slots_assigned:16384",
		"cluster_slots_ok:16384",
		"cluster_slots_fail:0",
		"cluster_known_nodes:" + fmt.Sprintf("%d", len(nodes)+1),
		"cluster_size:1",
		"cluster_current_epoch:0",
		"cluster_my_epoch:0",
		"cluster_stats_messages_sent:0",
		"cluster_stats_messages_received:0",
	}

	return resp.NewBulkReply([]byte(strings.Join(lines, "\r\n")))
}

func execClusterNodes(db *DB, args [][]byte) resp.Reply {
	cm := getClusterManager()

	var lines []string
	selfPort := cm.SelfPort()
	lines = append(lines, fmt.Sprintf("1234567890abcdef %s:%d %s@%d master - 0 0 connected 0-16383",
		cm.SelfIP(), selfPort, cm.SelfIP(), selfPort))

	nodes := cm.GetNodes()
	for _, node := range nodes {
		lines = append(lines, fmt.Sprintf("%s %s:%d %s@%d %s - 0 0 connected",
			node.NodeID, node.IP, node.Port, node.IP, node.Port, node.Role))
	}

	return resp.NewBulkReply([]byte(strings.Join(lines, "\n")))
}

func execClusterAddSlots(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'cluster addslots' command")
	}

	cm := getClusterManager()
	selfKey := fmt.Sprintf("%s:%d", cm.SelfIP(), cm.SelfPort())

	for _, arg := range args {
		slot, err := parseSlot(string(arg))
		if err != nil {
			return resp.NewErrorReply("ERR invalid slot number")
		}
		cm.AddSlot(slot, selfKey)
	}

	return resp.OkReply
}

func parseSlot(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid slot")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func init() {
	RegisterCommand("cluster", execCluster, nil, -2)
}

func execCluster(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'cluster' command")
	}

	subCmd := strings.ToLower(string(args[0]))
	switch subCmd {
	case "slots":
		return execClusterSlots(db, args[1:])
	case "info":
		return execClusterInfo(db, args[1:])
	case "nodes":
		return execClusterNodes(db, args[1:])
	case "addslots":
		return execClusterAddSlots(db, args[1:])
	}

	return resp.NewErrorReply("ERR CLUSTER subcommand must be SLOTS, INFO, NODES or ADDSLOTS")
}
