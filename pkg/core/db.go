package core

import (
	"goflashdb/pkg/resp"
	"strings"
	"sync"
	"time"
)

type ExecFunc func(db *DB, args [][]byte) resp.Reply

type PreFunc func(args [][]byte) (writeKeys []string, readKeys []string)

type command struct {
	executor ExecFunc
	prepare  PreFunc
	arity    int
}

var cmdTable = make(map[string]*command)

type DB struct {
	index   int
	data    *ConcurrentDict
	ttlDict *ConcurrentDict
	lock    sync.RWMutex
}

func NewDB(index int) *DB {
	return &DB{
		index:   index,
		data:    NewConcurrentDict(),
		ttlDict: NewConcurrentDict(),
	}
}

func (db *DB) Exec(cmdName string, args [][]byte) resp.Reply {
	cmd, ok := cmdTable[strings.ToLower(cmdName)]
	if !ok {
		return resp.NewErrorReply("ERR unknown command '" + cmdName + "'")
	}

	arity := cmd.arity
	if arity > 0 && len(args) != arity-1 {
		return resp.NewErrorReply("ERR wrong number of arguments for '" + cmdName + "' command")
	}
	if arity < 0 && len(args) < -arity-1 {
		return resp.NewErrorReply("ERR wrong number of arguments for '" + cmdName + "' command")
	}

	var writeKeys, readKeys []string
	if cmd.prepare != nil {
		writeKeys, readKeys = cmd.prepare(args)
		if len(writeKeys) > 0 {
			db.lock.Lock()
			defer db.lock.Unlock()
		} else if len(readKeys) > 0 {
			db.lock.RLock()
			defer db.lock.RUnlock()
		}
	}

	return cmd.executor(db, args)
}

func RegisterCommand(name string, executor ExecFunc, prepare PreFunc, arity int) {
	cmdTable[strings.ToLower(name)] = &command{
		executor: executor,
		prepare:  prepare,
		arity:    arity,
	}
}

func (db *DB) IsExpired(key string) bool {
	expireTime, ok := db.ttlDict.Get(key)
	if !ok {
		return false
	}
	expireMs := expireTime.(int64)
	return time.Now().UnixMilli() > expireMs
}

func (db *DB) RemoveExpire(key string) {
	db.ttlDict.Delete(key)
}

func (db *DB) Expire(key string, expireAt int64) {
	db.ttlDict.Set(key, expireAt)
}

func execDel(db *DB, args [][]byte) resp.Reply {
	count := 0
	for _, arg := range args {
		key := string(arg)
		if db.data.Delete(key) {
			count++
			db.ttlDict.Delete(key)
		}
	}
	return &resp.IntegerReply{Num: int64(count)}
}

func execExists(db *DB, args [][]byte) resp.Reply {
	count := 0
	for _, arg := range args {
		key := string(arg)
		if _, ok := db.data.Get(key); ok {
			count++
		}
	}
	return &resp.IntegerReply{Num: int64(count)}
}

func execPing(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return resp.PongReply
	}
	return resp.NewBulkReply(args[0])
}

func init() {
	RegisterCommand("del", execDel, func(args [][]byte) ([]string, []string) {
		keys := make([]string, len(args))
		for i, arg := range args {
			keys[i] = string(arg)
		}
		return keys, nil
	}, -2)
	RegisterCommand("exists", execExists, func(args [][]byte) ([]string, []string) {
		keys := make([]string, len(args))
		for i, arg := range args {
			keys[i] = string(arg)
		}
		return nil, keys
	}, -2)
	RegisterCommand("ping", execPing, nil, -1)
}
