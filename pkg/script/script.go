package script

import (
	"crypto/sha1"
	"encoding/hex"
	"sync"
	"time"

	"github.com/yuin/gopher-lua"
	"goflashdb/pkg/resp"
)

type LuaEngine struct {
	cache   *ScriptCache
	timeout time.Duration
	maxMem  int64
	mu      sync.RWMutex
}

type ScriptCache struct {
	mu      sync.RWMutex
	scripts map[string]*CompiledScript
}

type CompiledScript struct {
	source string
	sha1   string
}

func NewLuaEngine() *LuaEngine {
	return &LuaEngine{
		cache: &ScriptCache{
			scripts: make(map[string]*CompiledScript),
		},
		timeout: 5 * time.Second,
		maxMem:  4 * 1024 * 1024,
	}
}

func (e *LuaEngine) newSandboxedState() *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs:        true,
		RegistryMaxSize:     int(e.maxMem),
		RegistryGrowStep:    128,
		CallStackSize:       256,
		IncludeGoStackTrace: false,
	})

	L.OpenLibs()

	forbidden := []string{"io", "os", "file", "socket", "net", "debug", "package"}
	for _, name := range forbidden {
		L.SetGlobal(name, lua.LNil)
	}

	return L
}

func (e *LuaEngine) Eval(source string, keys []string, args [][]byte) (resp.Reply, error) {
	sha1Hash := sha1Hash(source)

	e.cache.mu.RLock()
	_, exists := e.cache.scripts[sha1Hash]
	e.cache.mu.RUnlock()

	if !exists {
		e.cache.mu.Lock()
		e.cache.scripts[sha1Hash] = &CompiledScript{
			source: source,
			sha1:   sha1Hash,
		}
		e.cache.mu.Unlock()
	}

	return e.EvalSHA(sha1Hash, keys, args)
}

func (e *LuaEngine) EvalSHA(sha1Hash string, keys []string, args [][]byte) (resp.Reply, error) {
	e.cache.mu.RLock()
	compiled, exists := e.cache.scripts[sha1Hash]
	e.cache.mu.RUnlock()

	if !exists {
		return resp.NewErrorReply("NOSCRIPT No matching script. Please use EVAL."), nil
	}

	L := e.newSandboxedState()
	defer L.Close()

	keysTable := L.NewTable()
	for i, key := range keys {
		L.SetTable(keysTable, lua.LNumber(i+1), lua.LString(key))
	}
	L.SetGlobal("KEYS", keysTable)

	argvTable := L.NewTable()
	for i, arg := range args {
		L.SetTable(argvTable, lua.LNumber(i+1), lua.LString(string(arg)))
	}
	L.SetGlobal("ARGV", argvTable)

	done := make(chan struct{})
	var result lua.LValue
	var err error

	go func() {
		defer close(done)
		err = L.DoString(compiled.source)
		if err == nil {
			result = L.Get(-1)
		}
	}()

	select {
	case <-done:
		if err != nil {
			return resp.NewErrorReply("ERR Error running script (call to " + sha1Hash + "): " + err.Error()), nil
		}
		return lValueToReply(result), nil
	case <-time.After(e.timeout):
		return resp.NewErrorReply("ERR Script execution timed out"), nil
	}
}

func (e *LuaEngine) LoadScript(source string) (string, error) {
	sha1Hash := sha1Hash(source)

	L := e.newSandboxedState()
	defer L.Close()

	err := L.DoString(source)
	if err != nil {
		return "", err
	}

	e.cache.mu.Lock()
	e.cache.scripts[sha1Hash] = &CompiledScript{
		source: source,
		sha1:   sha1Hash,
	}
	e.cache.mu.Unlock()

	return sha1Hash, nil
}

func (e *LuaEngine) Exists(sha1Hashes []string) []int {
	e.cache.mu.RLock()
	defer e.cache.mu.RUnlock()

	results := make([]int, len(sha1Hashes))
	for i, hash := range sha1Hashes {
		if _, exists := e.cache.scripts[hash]; exists {
			results[i] = 1
		} else {
			results[i] = 0
		}
	}
	return results
}

func (e *LuaEngine) Flush() {
	e.cache.mu.Lock()
	defer e.cache.mu.Unlock()

	e.cache.scripts = make(map[string]*CompiledScript)
}

func sha1Hash(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

func lValueToReply(value lua.LValue) resp.Reply {
	if value == nil || value == lua.LNil {
		return &resp.BulkReply{Arg: nil}
	}

	switch v := value.(type) {
	case lua.LBool:
		if bool(v) {
			return &resp.IntegerReply{Num: 1}
		}
		return &resp.IntegerReply{Num: 0}
	case lua.LNumber:
		return &resp.IntegerReply{Num: int64(v)}
	case lua.LString:
		return &resp.BulkReply{Arg: []byte(string(v))}
	case *lua.LTable:
		return lTableToReply(v)
	default:
		return &resp.BulkReply{Arg: []byte(v.String())}
	}
}

func lTableToReply(table *lua.LTable) resp.Reply {
	var replies []resp.Reply

	table.ForEach(func(_ lua.LValue, value lua.LValue) {
		replies = append(replies, lValueToReply(value))
	})

	return &resp.ArrayReply{Replies: replies}
}
