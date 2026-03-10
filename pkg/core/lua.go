package core

import (
	"strconv"
	"sync"

	"goflashdb/pkg/resp"
	"goflashdb/pkg/script"
)

var (
	luaEngine     *script.LuaEngine
	luaEngineOnce sync.Once
)

func getLuaEngine() *script.LuaEngine {
	luaEngineOnce.Do(func() {
		luaEngine = script.NewLuaEngine()
	})
	return luaEngine
}

func execEval(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'eval' command")
	}

	scriptSource := string(args[0])
	numKeysStr := string(args[1])

	numKeys, err := strconv.Atoi(numKeysStr)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer or out of range")
	}

	if numKeys < 0 {
		return resp.NewErrorReply("ERR number of keys must be positive")
	}

	if len(args) < 2+numKeys {
		return resp.NewErrorReply("ERR wrong number of arguments for 'eval' command")
	}

	keys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = string(args[2+i])
	}

	argvStart := 2 + numKeys
	argv := make([][]byte, len(args)-argvStart)
	for i := argvStart; i < len(args); i++ {
		argv[i-argvStart] = args[i]
	}

	engine := getLuaEngine()
	reply, err := engine.Eval(scriptSource, keys, argv)
	if err != nil {
		return resp.NewErrorReply("ERR " + err.Error())
	}

	return reply
}

func execEvalSHA(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'evalsha' command")
	}

	sha1Hash := string(args[0])
	numKeysStr := string(args[1])

	numKeys, err := strconv.Atoi(numKeysStr)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer or out of range")
	}

	if numKeys < 0 {
		return resp.NewErrorReply("ERR number of keys must be positive")
	}

	if len(args) < 2+numKeys {
		return resp.NewErrorReply("ERR wrong number of arguments for 'evalsha' command")
	}

	keys := make([]string, numKeys)
	for i := 0; i < numKeys; i++ {
		keys[i] = string(args[2+i])
	}

	argvStart := 2 + numKeys
	argv := make([][]byte, len(args)-argvStart)
	for i := argvStart; i < len(args); i++ {
		argv[i-argvStart] = args[i]
	}

	engine := getLuaEngine()
	reply, err := engine.EvalSHA(sha1Hash, keys, argv)
	if err != nil {
		return resp.NewErrorReply("ERR " + err.Error())
	}

	return reply
}

func execScript(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'script' command")
	}

	subcommand := string(args[0])
	engine := getLuaEngine()

	switch subcommand {
	case "load":
		if len(args) < 2 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'script|load' command")
		}
		sha1, err := engine.LoadScript(string(args[1]))
		if err != nil {
			return resp.NewErrorReply("ERR Error compiling script: " + err.Error())
		}
		return &resp.BulkReply{Arg: []byte(sha1)}

	case "exists":
		if len(args) < 2 {
			return resp.NewErrorReply("ERR wrong number of arguments for 'script|exists' command")
		}
		sha1Hashes := make([]string, len(args)-1)
		for i := 1; i < len(args); i++ {
			sha1Hashes[i-1] = string(args[i])
		}
		results := engine.Exists(sha1Hashes)

		replies := make([]resp.Reply, len(results))
		for i, r := range results {
			replies[i] = &resp.IntegerReply{Num: int64(r)}
		}
		return &resp.ArrayReply{Replies: replies}

	case "flush":
		engine.Flush()
		return resp.OkReply

	default:
		return resp.NewErrorReply("ERR unknown subcommand '" + subcommand + "'")
	}
}

func init() {
	RegisterCommand("eval", execEval, nil, -3)
	RegisterCommand("evalsha", execEvalSHA, nil, -3)
	RegisterCommand("script", execScript, nil, -2)
}
