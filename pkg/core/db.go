package core

import (
	"goflashdb/pkg/persist"
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
	index      int
	data       *ConcurrentDict
	ttlDict    *ConcurrentDict
	lock       sync.RWMutex
	persistMgr *persist.PersistManager
}

func NewDB(index int) *DB {
	return &DB{
		index:   index,
		data:    NewConcurrentDict(),
		ttlDict: NewConcurrentDict(),
	}
}

func NewDBWithPersist(index int, pm *persist.PersistManager) *DB {
	return &DB{
		index:      index,
		data:       NewConcurrentDict(),
		ttlDict:    NewConcurrentDict(),
		persistMgr: pm,
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
	isWrite := false
	if cmd.prepare != nil {
		writeKeys, readKeys = cmd.prepare(args)
		isWrite = len(writeKeys) > 0
		if isWrite {
			db.lock.Lock()
			defer db.lock.Unlock()
		} else if len(readKeys) > 0 {
			db.lock.RLock()
			defer db.lock.RUnlock()
		}
	}

	reply := cmd.executor(db, args)

	// 写命令成功后追加 AOF
	if isWrite && db.persistMgr != nil {
		aofCmd := db.buildAOFCommand(cmdName, args)
		db.persistMgr.AppendAOF(aofCmd)
	}

	return reply
}

func (db *DB) buildAOFCommand(cmdName string, args [][]byte) []byte {
	var buf []byte
	buf = append(buf, '*')
	buf = append(buf, []byte(itoa(len(args)+1))...)
	buf = append(buf, '\r', '\n')

	buf = append(buf, '$')
	buf = append(buf, []byte(itoa(len(cmdName)))...)
	buf = append(buf, '\r', '\n')
	buf = append(buf, []byte(strings.ToUpper(cmdName))...)
	buf = append(buf, '\r', '\n')

	for _, arg := range args {
		buf = append(buf, '$')
		buf = append(buf, []byte(itoa(len(arg)))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, arg...)
		buf = append(buf, '\r', '\n')
	}

	return buf
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
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
	RegisterCommand("save", execSave, nil, 1)
}

func (db *DB) LoadFromPersist() error {
	if db.persistMgr == nil {
		return nil
	}

	pairs, err := db.persistMgr.LoadRDB()
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	for _, pair := range pairs {
		if pair.ExpireAt > 0 && pair.ExpireAt <= now {
			continue
		}
		db.data.Set(pair.Key, pair.Value)
		if pair.ExpireAt > 0 {
			db.ttlDict.Set(pair.Key, pair.ExpireAt)
		}
	}

	commands, err := db.persistMgr.LoadAOF()
	if err != nil {
		return err
	}

	for _, cmd := range commands {
		db.replayCommand(cmd)
	}

	return nil
}

func (db *DB) replayCommand(cmd []byte) {
	cmdName, args, err := resp.ParseCommand(cmd)
	if err != nil || cmdName == "" {
		return
	}

	db.Exec(cmdName, args)
}

func (db *DB) GetAllData() map[string][]byte {
	db.lock.RLock()
	defer db.lock.RUnlock()

	data := make(map[string][]byte)
	db.data.ForEach(func(key string, value interface{}) bool {
		if !db.IsExpired(key) {
			data[key] = value.([]byte)
		}
		return true
	})
	return data
}

func (db *DB) GetAllExpireTimes() map[string]int64 {
	db.lock.RLock()
	defer db.lock.RUnlock()

	expireTimes := make(map[string]int64)
	db.ttlDict.ForEach(func(key string, value interface{}) bool {
		expireTimes[key] = value.(int64)
		return true
	})
	return expireTimes
}

func execSave(db *DB, args [][]byte) resp.Reply {
	if db.persistMgr == nil {
		return resp.NewErrorReply("ERR persistence not enabled")
	}

	data := db.GetAllData()
	expireTimes := db.GetAllExpireTimes()

	if err := db.persistMgr.SaveRDB(data, expireTimes); err != nil {
		return resp.NewErrorReply("ERR " + err.Error())
	}

	return resp.OkReply
}
