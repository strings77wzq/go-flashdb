package core

import (
	"strconv"
	"strings"
	"sync"

	"goflashdb/pkg/replication"
	"goflashdb/pkg/resp"
)

var (
	replManager     *replication.ReplicationManager
	replManagerOnce sync.Once
)

func getReplManager() *replication.ReplicationManager {
	replManagerOnce.Do(func() {
		replManager = replication.NewReplicationManager()
	})
	return replManager
}

func execReplicaOf(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'replicaof' command")
	}

	host := strings.ToLower(string(args[0]))
	portStr := string(args[1])

	var port int
	var err error

	if host == "no" {
		port = 0
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return resp.NewErrorReply("ERR value is not an integer or out of range")
		}
	}

	rm := getReplManager()
	err = rm.ReplicaOf(host, port)
	if err != nil {
		return resp.NewErrorReply("ERR " + err.Error())
	}

	return resp.OkReply
}

func execSlaveOf(db *DB, args [][]byte) resp.Reply {
	return execReplicaOf(db, args)
}

func execRole(db *DB, args [][]byte) resp.Reply {
	rm := getReplManager()
	role := rm.Role()

	var replies []resp.Reply

	if role == replication.RoleMaster {
		replies = append(replies, &resp.BulkReply{Arg: []byte("master")})

		slaves := rm.GetConnectedSlaves()
		slaveReplies := make([]resp.Reply, len(slaves))
		for i, slave := range slaves {
			slaveReplies[i] = &resp.ArrayReply{
				Replies: []resp.Reply{
					&resp.BulkReply{Arg: []byte(slave)},
				},
			}
		}
		replies = append(replies, &resp.ArrayReply{Replies: slaveReplies})
		replies = append(replies, &resp.IntegerReply{Num: rm.GetReplOffset()})
	} else {
		replies = append(replies, &resp.BulkReply{Arg: []byte("slave")})

		host, port := rm.GetMasterInfo()
		state := rm.GetSlaveState()

		replies = append(replies, &resp.ArrayReply{
			Replies: []resp.Reply{
				&resp.BulkReply{Arg: []byte(host)},
				&resp.IntegerReply{Num: int64(port)},
				&resp.BulkReply{Arg: []byte(state.String())},
			},
		})
		replies = append(replies, &resp.IntegerReply{Num: rm.GetReplOffset()})
	}

	return &resp.ArrayReply{Replies: replies}
}

func execInfoReplication(db *DB, args [][]byte) resp.Reply {
	rm := getReplManager()
	info := rm.GetInfo()

	var lines []string
	lines = append(lines, "# Replication")
	lines = append(lines, "role:"+info["role"].(string))
	lines = append(lines, "connected_slaves:"+strconv.Itoa(info["connected_slaves"].(int)))
	lines = append(lines, "master_replid:"+info["master_replid"].(string))
	lines = append(lines, "master_repl_offset:"+strconv.FormatInt(info["master_repl_offset"].(int64), 10))

	if masterHost, ok := info["master_host"]; ok {
		lines = append(lines, "master_host:"+masterHost.(string))
	}
	if masterPort, ok := info["master_port"]; ok {
		lines = append(lines, "master_port:"+strconv.Itoa(masterPort.(int)))
	}

	return &resp.BulkReply{Arg: []byte(strings.Join(lines, "\r\n"))}
}

func init() {
	RegisterCommand("replicaof", execReplicaOf, nil, 3)
	RegisterCommand("slaveof", execSlaveOf, nil, 3)
	RegisterCommand("role", execRole, nil, 1)
}
