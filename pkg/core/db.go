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

// ServerState holds global server state for server commands
type ServerState struct {
	clients      map[int64]*ClientInfo
	clientMutex  sync.RWMutex
	clientIDGen  int64
	slowLog      []SlowLogEntry
	slowLogMutex sync.RWMutex
	slowLogMax   int
	startTime    time.Time
	shutdownChan chan struct{}
}

// ClientInfo holds information about a connected client
type ClientInfo struct {
	ID       int64
	Addr     string
	LibName  string
	DB       int
	IdleTime time.Time
	Command  string
}

// SlowLogEntry represents a slow log entry
type SlowLogEntry struct {
	ID       int64
	Time     time.Time
	Duration time.Duration
	Command  string
	Args     []string
}

var serverState = &ServerState{
	clients:      make(map[int64]*ClientInfo),
	clientIDGen:  0,
	slowLog:      make([]SlowLogEntry, 0),
	slowLogMax:   128,
	startTime:    time.Now(),
	shutdownChan: make(chan struct{}),
}

// SlowLogThreshold defines the threshold for slow log (microseconds)
var SlowLogThreshold = 10000 // 10ms default

type DB struct {
	index      int
	data       *ConcurrentDict
	ttlDict    *ConcurrentDict
	lock       sync.RWMutex
	persistMgr *persist.PersistManager

	tx     *Transaction
	txLock sync.Mutex

	watchMap map[string]map[int64]struct{}
	watchMu  sync.Mutex
}

type Transaction struct {
	commands  [][]byte
	discarded bool
}

func NewDB(index int) *DB {
	return &DB{
		index:    index,
		data:     NewConcurrentDict(),
		ttlDict:  NewConcurrentDict(),
		watchMap: make(map[string]map[int64]struct{}),
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

			db.watchMu.Lock()
			delete(db.watchMap, key)
			db.watchMu.Unlock()
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
	}, -1)
	RegisterCommand("flushall", execFlushAll, func(args [][]byte) ([]string, []string) {
		return []string{"__all__"}, nil
	}, -1)
	RegisterCommand("select", execSelect, nil, 2)
	RegisterCommand("quit", execQuit, nil, 1)
	RegisterCommand("multi", execMulti, nil, 1)
	RegisterCommand("exec", execExec, nil, 1)
	RegisterCommand("discard", execDiscard, nil, 1)
	RegisterCommand("watch", execWatch, nil, -2)
	RegisterCommand("unwatch", execUnwatch, nil, 1)
	RegisterCommand("info", execInfo, nil, -1)
	RegisterCommand("config", execConfig, nil, -3)
	RegisterCommand("time", execTime, nil, 1)
	RegisterCommand("shutdown", execShutdown, nil, -1)
	RegisterCommand("client", execClient, nil, -2)
	RegisterCommand("slowlog", execSlowLog, nil, -2)
	RegisterCommand("command", execCommand, nil, -1)
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
	async := false
	if len(args) > 0 {
		arg := strings.ToLower(string(args[0]))
		if arg == "async" {
			async = true
		}
	}

	if async {
		go func() {
			db.data = NewConcurrentDict()
			db.ttlDict = NewConcurrentDict()
		}()
		return resp.OkReply
	}

	db.data = NewConcurrentDict()
	db.ttlDict = NewConcurrentDict()
	return resp.OkReply
}

var flushAllFunc func() error

func RegisterFlushAll(fn func() error) {
	flushAllFunc = fn
}

func execFlushAll(db *DB, args [][]byte) resp.Reply {
	async := false
	if len(args) > 0 {
		arg := strings.ToLower(string(args[0]))
		if arg == "async" {
			async = true
		}
	}

	if flushAllFunc == nil {
		return resp.NewErrorReply("ERR FLUSHALL is not supported in single database mode")
	}

	if async {
		go func() {
			flushAllFunc()
		}()
		return resp.OkReply
	}

	if err := flushAllFunc(); err != nil {
		return resp.NewErrorReply("ERR " + err.Error())
	}
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

	db.watchMu.Lock()
	watchedKeys := make([]string, 0, len(db.watchMap))
	for key := range db.watchMap {
		watchedKeys = append(watchedKeys, key)
	}
	db.watchMu.Unlock()

	if len(watchedKeys) > 0 {
		for _, key := range watchedKeys {
			_, exists := db.data.Get(key)
			if !exists {
				continue
			}
			prevVal, hasPrev := db.data.Get(key)
			db.lock.RLock()
			db.lock.RUnlock()

			curVal, hasCur := db.data.Get(key)
			if !hasPrev && hasCur || hasPrev && !hasCur || (hasPrev && hasCur && prevVal != curVal) {
				db.watchMu.Lock()
				db.watchMap = make(map[string]map[int64]struct{})
				db.watchMu.Unlock()
				return &resp.ArrayReply{Replies: []resp.Reply{}}
			}
		}
	}

	if len(tx.commands) == 0 {
		db.watchMu.Lock()
		db.watchMap = make(map[string]map[int64]struct{})
		db.watchMu.Unlock()
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

	db.watchMu.Lock()
	db.watchMap = make(map[string]map[int64]struct{})
	db.watchMu.Unlock()

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

	db.watchMu.Lock()
	db.watchMap = make(map[string]map[int64]struct{})
	db.watchMu.Unlock()

	return resp.OkReply
}

func execWatch(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'watch' command")
	}

	clientID := int64(0)
	db.watchMu.Lock()
	defer db.watchMu.Unlock()

	for _, arg := range args {
		key := string(arg)
		if _, ok := db.watchMap[key]; !ok {
			db.watchMap[key] = make(map[int64]struct{})
		}
		db.watchMap[key][clientID] = struct{}{}
	}

	return resp.OkReply
}

func execUnwatch(db *DB, args [][]byte) resp.Reply {
	db.watchMu.Lock()
	defer db.watchMu.Unlock()

	db.watchMap = make(map[string]map[int64]struct{})

	return resp.OkReply
}

func execInfo(db *DB, args [][]byte) resp.Reply {
	section := ""
	if len(args) > 0 {
		section = strings.ToLower(string(args[0]))
	}

	var lines []string
	lines = append(lines, "# Server")
	lines = append(lines, "goflashdb_version:1.0.0")
	lines = append(lines, "goflashdb_mode:standalone")
	lines = append(lines, "tcp_port:6379")
	lines = append(lines, "uptime:"+itoa(int(time.Since(serverState.startTime).Seconds())))
	lines = append(lines, "")

	lines = append(lines, "# Clients")
	serverState.clientMutex.RLock()
	connectedClients := len(serverState.clients)
	serverState.clientMutex.RUnlock()
	lines = append(lines, "connected_clients:"+itoa(connectedClients))
	lines = append(lines, "")

	lines = append(lines, "# Memory")
	lines = append(lines, "used_memory:0")
	lines = append(lines, "")

	lines = append(lines, "# Persistence")
	lines = append(lines, "rdb_changes_since_last_save:0")
	lines = append(lines, "rdb_last_save_time:0")
	lines = append(lines, "")

	lines = append(lines, "# Stats")
	lines = append(lines, "total_connections_received:0")
	lines = append(lines, "total_commands_processed:0")
	lines = append(lines, "")

	lines = append(lines, "# SlowLog")
	serverState.slowLogMutex.RLock()
	slowLogLen := len(serverState.slowLog)
	serverState.slowLogMutex.RUnlock()
	lines = append(lines, "slowlog_length:"+itoa(slowLogLen))
	lines = append(lines, "slowlog_log_slower_than:"+itoa(SlowLogThreshold))
	lines = append(lines, "")

	lines = append(lines, "# CommandStats")
	lines = append(lines, "total_commands:0")

	if section == "clients" {
		lines = []string{"# Clients"}
		lines = append(lines, "connected_clients:"+itoa(connectedClients))
	} else if section == "memory" {
		lines = []string{"# Memory"}
		lines = append(lines, "used_memory:0")
	} else if section == "persistence" {
		lines = []string{"# Persistence"}
		lines = append(lines, "rdb_changes_since_last_save:0")
		lines = append(lines, "rdb_last_save_time:0")
	} else if section == "stats" {
		lines = []string{"# Stats"}
		lines = append(lines, "total_connections_received:0")
		lines = append(lines, "total_commands_processed:0")
	} else if section == "slowlog" {
		lines = []string{"# SlowLog"}
		lines = append(lines, "slowlog_length:"+itoa(slowLogLen))
		lines = append(lines, "slowlog_log_slower_than:"+itoa(SlowLogThreshold))
	}

	output := strings.Join(lines, "\r\n")
	return resp.NewBulkReply([]byte(output))
}

func execConfig(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'config' command")
	}

	subCmd := strings.ToLower(string(args[0]))
	if subCmd == "get" {
		if len(args) < 2 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'config get' command")
		}
		pattern := string(args[1])
		return configGet(pattern)
	} else if subCmd == "set" {
		if len(args) < 3 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'config set' command")
		}
		key := string(args[1])
		value := string(args[2])
		return configSet(key, value)
	}

	return resp.NewErrorReply("ERR CONFIG subcommand must be GET or SET")
}

func configGet(pattern string) resp.Reply {
	configs := []string{
		"maxmemory", "0",
		"maxclients", "10000",
		"timeout", "0",
		"tcp-keepalive", "300",
		"loglevel", "notice",
		"databases", "16",
		"save", "900 1 300 10 60 10000",
		"appendonly", "no",
		"appendfsync", "everysec",
		"slowlog-log-slower-than", "10000",
		"slowlog-max-len", "128",
	}

	var matches []resp.Reply
	for i := 0; i < len(configs); i += 2 {
		key := configs[i]
		if strings.Contains(key, pattern) || pattern == "*" {
			matches = append(matches, resp.NewBulkReply([]byte(key)))
			matches = append(matches, resp.NewBulkReply([]byte(configs[i+1])))
		}
	}

	if len(matches) == 0 {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	return &resp.ArrayReply{Replies: matches}
}

func configSet(key, value string) resp.Reply {
	key = strings.ToLower(key)

	switch key {
	case "slowlog-log-slower-than":
		threshold, err := parseInt64(value)
		if err != nil {
			return resp.NewErrorReply("ERR invalid value for 'slowlog-log-slower-than'")
		}
		SlowLogThreshold = int(threshold)
	case "slowlog-max-len":
		maxLen, err := parseInt64(value)
		if err != nil {
			return resp.NewErrorReply("ERR invalid value for 'slowlog-max-len'")
		}
		serverState.slowLogMutex.Lock()
		if int(maxLen) < len(serverState.slowLog) {
			serverState.slowLog = serverState.slowLog[:maxLen]
		}
		serverState.slowLogMax = int(maxLen)
		serverState.slowLogMutex.Unlock()
	default:
		return resp.NewErrorReply("ERR CONFIG SET '" + key + "' is not supported")
	}

	return resp.OkReply
}

func execTime(db *DB, args [][]byte) resp.Reply {
	unixTime := time.Now().Unix()
	microseconds := time.Now().UnixMicro() % 1000000

	unixStr := itoa(int(unixTime))
	microStr := itoa(int(microseconds))

	return &resp.ArrayReply{
		Replies: []resp.Reply{
			resp.NewBulkReply([]byte(unixStr)),
			resp.NewBulkReply([]byte(microStr)),
		},
	}
}

func execShutdown(db *DB, args [][]byte) resp.Reply {
	hasSave := false
	hasNosave := false

	for _, arg := range args {
		argStr := strings.ToLower(string(arg))
		if argStr == "save" {
			hasSave = true
		} else if argStr == "nosave" {
			hasNosave = true
		}
	}

	if hasSave && hasNosave {
		return resp.NewErrorReply("ERR SHUTDOWN SAVE and NOSAVE can't be used together")
	}

	if db.persistMgr != nil && !hasNosave {
		data := db.GetAllData()
		expireTimes := db.GetAllExpireTimes()
		db.persistMgr.SaveRDB(data, expireTimes)
	}

	select {
	case serverState.shutdownChan <- struct{}{}:
	default:
	}

	return nil
}

func execClient(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'client' command")
	}

	subCmd := strings.ToLower(string(args[0]))
	if subCmd == "list" {
		return clientList()
	} else if subCmd == "info" {
		return clientInfo()
	} else if subCmd == "kill" {
		if len(args) < 2 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'client kill' command")
		}
		return clientKill(string(args[1]))
	}

	return resp.NewErrorReply("ERR CLIENT subcommand must be LIST, INFO, or KILL")
}

func clientList() resp.Reply {
	serverState.clientMutex.RLock()
	defer serverState.clientMutex.RUnlock()

	if len(serverState.clients) == 0 {
		return resp.NewBulkReply([]byte(""))
	}

	var lines []string
	for _, client := range serverState.clients {
		idle := int(time.Since(client.IdleTime).Seconds())
		line := "id=" + itoa(int(client.ID)) +
			" addr=" + client.Addr +
			" db=" + itoa(client.DB) +
			" idle=" + itoa(idle)
		lines = append(lines, line)
	}

	output := strings.Join(lines, "\r\n")
	return resp.NewBulkReply([]byte(output))
}

func clientInfo() resp.Reply {
	serverState.clientMutex.RLock()
	defer serverState.clientMutex.RUnlock()

	if len(serverState.clients) == 0 {
		return resp.NewBulkReply([]byte(""))
	}

	for _, client := range serverState.clients {
		idle := int(time.Since(client.IdleTime).Seconds())
		output := "id=" + itoa(int(client.ID)) +
			" addr=" + client.Addr +
			" db=" + itoa(client.DB) +
			" idle=" + itoa(idle)
		return resp.NewBulkReply([]byte(output))
	}

	return resp.NewBulkReply([]byte(""))
}

func clientKill(addr string) resp.Reply {
	return resp.NewErrorReply("ERR CLIENT KILL not implemented")
}

func execSlowLog(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'slowlog' command")
	}

	subCmd := strings.ToLower(string(args[0]))
	if subCmd == "get" {
		count := 10
		if len(args) > 1 {
			c, err := parseInt64(string(args[1]))
			if err == nil {
				count = int(c)
			}
		}
		return slowLogGet(count)
	} else if subCmd == "len" {
		serverState.slowLogMutex.RLock()
		length := len(serverState.slowLog)
		serverState.slowLogMutex.RUnlock()
		return &resp.IntegerReply{Num: int64(length)}
	} else if subCmd == "reset" {
		serverState.slowLogMutex.Lock()
		serverState.slowLog = make([]SlowLogEntry, 0)
		serverState.slowLogMutex.Unlock()
		return resp.OkReply
	}

	return resp.NewErrorReply("ERR SLOWLOG subcommand must be GET, LEN, or RESET")
}

func slowLogGet(count int) resp.Reply {
	serverState.slowLogMutex.RLock()
	defer serverState.slowLogMutex.RUnlock()

	entries := serverState.slowLog
	if count > 0 && count < len(entries) {
		entries = entries[len(entries)-count:]
	}

	var results []resp.Reply
	for _, entry := range entries {
		args := make([]resp.Reply, 4)
		args[0] = &resp.IntegerReply{Num: entry.ID}
		args[1] = &resp.IntegerReply{Num: entry.Time.Unix()}
		args[2] = &resp.IntegerReply{Num: int64(entry.Duration.Microseconds())}

		cmdStr := entry.Command
		for _, a := range entry.Args {
			cmdStr += " " + a
		}
		args[3] = resp.NewBulkReply([]byte(cmdStr))

		results = append(results, &resp.ArrayReply{Replies: args})
	}

	return &resp.ArrayReply{Replies: results}
}

func execCommand(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return commandInfo()
	}

	subCmd := strings.ToLower(string(args[0]))
	if subCmd == "info" {
		if len(args) < 2 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'command info' command")
		}
		return commandInfoCmd(string(args[1]))
	} else if subCmd == "count" {
		return &resp.IntegerReply{Num: int64(len(cmdTable))}
	} else if subCmd == "list" {
		return commandList()
	}

	return resp.NewErrorReply("ERR COMMAND subcommand must be INFO, COUNT, or LIST")
}

func commandInfo() resp.Reply {
	return &resp.IntegerReply{Num: int64(len(cmdTable))}
}

func commandInfoCmd(cmdName string) resp.Reply {
	cmd, ok := cmdTable[strings.ToLower(cmdName)]
	if !ok {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	var replies []resp.Reply
	replies = append(replies, resp.NewBulkReply([]byte(strings.ToLower(cmdName))))
	replies = append(replies, &resp.IntegerReply{Num: int64(cmd.arity)})
	replies = append(replies, &resp.ArrayReply{Replies: []resp.Reply{
		resp.NewBulkReply([]byte("write")),
	}})
	replies = append(replies, &resp.IntegerReply{Num: 1})
	replies = append(replies, &resp.IntegerReply{Num: 1})
	replies = append(replies, &resp.ArrayReply{Replies: []resp.Reply{}})

	return &resp.ArrayReply{Replies: replies}
}

func commandList() resp.Reply {
	var replies []resp.Reply
	for name := range cmdTable {
		replies = append(replies, resp.NewBulkReply([]byte(name)))
	}
	return &resp.ArrayReply{Replies: replies}
}

func AddClient(addr string, dbIndex int) int64 {
	serverState.clientMutex.Lock()
	defer serverState.clientMutex.Unlock()

	serverState.clientIDGen++
	id := serverState.clientIDGen
	serverState.clients[id] = &ClientInfo{
		ID:       id,
		Addr:     addr,
		DB:       dbIndex,
		IdleTime: time.Now(),
	}
	return id
}

func UpdateClient(id int64, cmd string) {
	serverState.clientMutex.Lock()
	defer serverState.clientMutex.Unlock()

	if client, ok := serverState.clients[id]; ok {
		client.Command = cmd
		client.IdleTime = time.Now()
	}
}

func RemoveClient(id int64) {
	serverState.clientMutex.Lock()
	defer serverState.clientMutex.Unlock()

	delete(serverState.clients, id)
}

func AddSlowLog(command string, args []string, duration time.Duration) {
	if int(duration.Microseconds()) < SlowLogThreshold {
		return
	}

	serverState.slowLogMutex.Lock()
	defer serverState.slowLogMutex.Unlock()

	entry := SlowLogEntry{
		ID:       int64(len(serverState.slowLog)) + 1,
		Time:     time.Now(),
		Duration: duration,
		Command:  command,
		Args:     args,
	}

	serverState.slowLog = append(serverState.slowLog, entry)

	if len(serverState.slowLog) > serverState.slowLogMax {
		serverState.slowLog = serverState.slowLog[1:]
	}
}

func GetShutdownChan() <-chan struct{} {
	return serverState.shutdownChan
}
