package core

import (
	"errors"
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

	tx     *Transaction
	txLock sync.Mutex
}

type Transaction struct {
	commands  [][]byte
	discarded bool
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
	RegisterCommand("expire", execExpire, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 3)
	RegisterCommand("pexpire", execPExpire, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 3)
	RegisterCommand("expireat", execExpireAt, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 3)
	RegisterCommand("pexpireat", execPExpireAt, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 3)
	RegisterCommand("ttl", execTTL, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("pttl", execPTTL, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("persist", execPersist, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 2)
	RegisterCommand("echo", execEcho, nil, 2)
	RegisterCommand("dbsize", execDBSize, nil, 1)
	RegisterCommand("flushdb", execFlushDB, func(args [][]byte) ([]string, []string) {
		return []string{"__all__"}, nil
	}, 1)
	RegisterCommand("select", execSelect, nil, 2)
	RegisterCommand("quit", execQuit, nil, 1)
	RegisterCommand("multi", execMulti, nil, 1)
	RegisterCommand("exec", execExec, nil, 1)
	RegisterCommand("discard", execDiscard, nil, 1)
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
			switch v := value.(type) {
			case []byte:
				data[key] = v
			case *StringData:
				data[key] = v.value
			case *HashData:
				data[key] = nil
			case *ListData:
				data[key] = nil
			case *SetData:
				data[key] = nil
			case *ZSetData:
				data[key] = nil
			}
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
		if v, ok := value.(int64); ok {
			expireTimes[key] = v
		}
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

func execExpire(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	seconds, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer or out of range")
	}

	if _, ok := db.data.Get(key); !ok {
		return &resp.IntegerReply{Num: 0}
	}

	expireAt := time.Now().Unix() + seconds
	db.Expire(key, expireAt*1000)
	return &resp.IntegerReply{Num: 1}
}

func execPExpire(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	milliseconds, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer or out of range")
	}

	if _, ok := db.data.Get(key); !ok {
		return &resp.IntegerReply{Num: 0}
	}

	expireAt := time.Now().UnixMilli() + milliseconds
	db.Expire(key, expireAt)
	return &resp.IntegerReply{Num: 1}
}

func execExpireAt(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	timestamp, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer or out of range")
	}

	if _, ok := db.data.Get(key); !ok {
		return &resp.IntegerReply{Num: 0}
	}

	db.Expire(key, timestamp*1000)
	return &resp.IntegerReply{Num: 1}
}

func execPExpireAt(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	timestamp, err := parseInt64(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer or out of range")
	}

	if _, ok := db.data.Get(key); !ok {
		return &resp.IntegerReply{Num: 0}
	}

	db.Expire(key, timestamp)
	return &resp.IntegerReply{Num: 1}
}

func execTTL(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	_, ok := db.data.Get(key)
	if !ok {
		return &resp.IntegerReply{Num: -2}
	}

	expireTime, ok := db.ttlDict.Get(key)
	if !ok {
		return &resp.IntegerReply{Num: -1}
	}

	expireMs := expireTime.(int64)
	remaining := expireMs - time.Now().UnixMilli()
	if remaining <= 0 {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return &resp.IntegerReply{Num: -2}
	}

	return &resp.IntegerReply{Num: remaining / 1000}
}

func execPTTL(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	_, ok := db.data.Get(key)
	if !ok {
		return &resp.IntegerReply{Num: -2}
	}

	expireTime, ok := db.ttlDict.Get(key)
	if !ok {
		return &resp.IntegerReply{Num: -1}
	}

	expireMs := expireTime.(int64)
	remaining := expireMs - time.Now().UnixMilli()
	if remaining <= 0 {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return &resp.IntegerReply{Num: -2}
	}

	return &resp.IntegerReply{Num: remaining}
}

func execPersist(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	if _, ok := db.data.Get(key); !ok {
		return &resp.IntegerReply{Num: 0}
	}

	if _, ok := db.ttlDict.Get(key); !ok {
		return &resp.IntegerReply{Num: 0}
	}

	db.ttlDict.Delete(key)
	return &resp.IntegerReply{Num: 1}
}

func execEcho(db *DB, args [][]byte) resp.Reply {
	return resp.NewBulkReply(args[0])
}

func execDBSize(db *DB, args [][]byte) resp.Reply {
	count := 0
	db.data.ForEach(func(key string, value interface{}) bool {
		if !db.IsExpired(key) {
			count++
		}
		return true
	})
	return &resp.IntegerReply{Num: int64(count)}
}

func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.data = NewConcurrentDict()
	db.ttlDict = NewConcurrentDict()
	return resp.OkReply
}

func execSelect(db *DB, args [][]byte) resp.Reply {
	index, err := parseInt64(string(args[0]))
	if err != nil {
		return resp.NewErrorReply("ERR invalid DB index")
	}
	db.index = int(index)
	return resp.OkReply
}

func execQuit(db *DB, args [][]byte) resp.Reply {
	return resp.OkReply
}

func parseInt64(s string) (int64, error) {
	var n int64
	var negative bool
	i := 0

	if len(s) > 0 && s[0] == '-' {
		negative = true
		i = 1
	}

	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errors.New("not an integer")
		}
		n = n*10 + int64(s[i]-'0')
	}

	if negative {
		n = -n
	}
	return n, nil
}

func execMulti(db *DB, args [][]byte) resp.Reply {
	db.txLock.Lock()
	defer db.txLock.Unlock()

	if db.tx != nil {
		return resp.NewErrorReply("ERR MULTI calls can not be nested")
	}

	db.tx = &Transaction{
		commands: make([][]byte, 0),
	}
	return resp.OkReply
}

func execExec(db *DB, args [][]byte) resp.Reply {
	db.txLock.Lock()
	tx := db.tx
	db.tx = nil
	db.txLock.Unlock()

	if tx == nil {
		return resp.NewErrorReply("ERR EXEC without MULTI")
	}

	if tx.discarded {
		return resp.NewErrorReply("ERR DISCARD called")
	}

	if len(tx.commands) == 0 {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	replies := make([]resp.Reply, 0, len(tx.commands))
	for _, cmd := range tx.commands {
		cmdName, cmdArgs, err := resp.ParseCommand(cmd)
		if err != nil {
			replies = append(replies, resp.NewErrorReply("ERR "+err.Error()))
			continue
		}

		result := db.Exec(cmdName, cmdArgs)
		replies = append(replies, result)
	}

	return &resp.ArrayReply{Replies: replies}
}

func execDiscard(db *DB, args [][]byte) resp.Reply {
	db.txLock.Lock()
	defer db.txLock.Unlock()

	if db.tx == nil {
		return resp.NewErrorReply("ERR DISCARD without MULTI")
	}

	db.tx.discarded = true
	db.tx = nil
	return resp.OkReply
}
