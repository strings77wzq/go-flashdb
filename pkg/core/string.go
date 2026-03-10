package core

import (
	"goflashdb/pkg/resp"
	"strconv"
	"time"
)

func init() {
	RegisterCommand("get", execGet, func(args [][]byte) ([]string, []string) {
		return nil, []string{string(args[0])}
	}, 2)
	RegisterCommand("set", execSet, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, -3)
	RegisterCommand("setnx", execSetNX, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 3)
	RegisterCommand("setex", execSetEX, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 4)
	RegisterCommand("psetex", execPSetEX, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 4)
	RegisterCommand("mset", execMSet, func(args [][]byte) ([]string, []string) {
		keys := make([]string, 0, len(args)/2)
		for i := 0; i < len(args); i += 2 {
			keys = append(keys, string(args[i]))
		}
		return keys, nil
	}, -3)
	RegisterCommand("mget", execMGet, func(args [][]byte) ([]string, []string) {
		keys := make([]string, len(args))
		for i := 0; i < len(args); i++ {
			keys[i] = string(args[i])
		}
		return nil, keys
	}, -2)
	RegisterCommand("incr", execIncr, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 2)
	RegisterCommand("decr", execDecr, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 2)
	RegisterCommand("incrby", execIncrBy, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 3)
	RegisterCommand("decrby", execDecrBy, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 3)
	RegisterCommand("append", execAppend, func(args [][]byte) ([]string, []string) {
		return []string{string(args[0])}, nil
	}, 3)
	RegisterCommand("strlen", execStrlen, func(args [][]byte) ([]string, []string) {
		return nil, []string{string(args[0])}
	}, 2)
}

func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	data, ok := db.GetStringData(key)
	if !ok {
		return resp.NilBulkReply
	}
	return resp.NewBulkReply(data.value)
}

func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	db.SetString(key, value)
	return resp.OkReply
}

func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	existed := db.data.SetIfAbsent(key, &StringData{
		value:    value,
		expireAt: 0,
	})
	if existed {
		return &resp.IntegerReply{Num: 0}
	}
	return &resp.IntegerReply{Num: 1}
}

func execSetEX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	seconds, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	expireAt := time.Now().Unix() + seconds
	db.SetStringWithExpire(key, value, expireAt*1000)
	return resp.OkReply
}

func execPSetEX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	milliseconds, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	expireAt := time.Now().UnixMilli() + milliseconds
	db.SetStringWithExpire(key, value, expireAt)
	return resp.OkReply
}

func execMSet(db *DB, args [][]byte) resp.Reply {
	for i := 0; i < len(args); i += 2 {
		key := string(args[i])
		value := args[i+1]
		db.SetString(key, value)
	}
	return resp.OkReply
}

func execMGet(db *DB, args [][]byte) resp.Reply {
	replies := make([]resp.Reply, len(args))
	for i := 0; i < len(args); i++ {
		key := string(args[i])
		data, ok := db.GetStringData(key)
		if !ok {
			replies[i] = resp.NilBulkReply
		} else {
			replies[i] = resp.NewBulkReply(data.value)
		}
	}
	return &resp.ArrayReply{Replies: replies}
}

func execIncr(db *DB, args [][]byte) resp.Reply {
	return incrBy(db, string(args[0]), 1)
}

func execDecr(db *DB, args [][]byte) resp.Reply {
	return incrBy(db, string(args[0]), -1)
}

func execIncrBy(db *DB, args [][]byte) resp.Reply {
	delta, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	return incrBy(db, string(args[0]), delta)
}

func execDecrBy(db *DB, args [][]byte) resp.Reply {
	delta, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	return incrBy(db, string(args[0]), -delta)
}

func incrBy(db *DB, key string, delta int64) resp.Reply {
	data, ok := db.GetStringData(key)
	if !ok {
		data = &StringData{
			value:    []byte("0"),
			expireAt: 0,
		}
	}
	val, err := strconv.ParseInt(string(data.value), 10, 64)
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	val += delta
	data.value = []byte(strconv.FormatInt(val, 10))
	db.SetString(key, data.value)
	return &resp.IntegerReply{Num: val}
}

func execAppend(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	data, ok := db.GetStringData(key)
	if !ok {
		data = &StringData{
			value:    []byte{},
			expireAt: 0,
		}
	}
	data.value = append(data.value, value...)
	db.SetString(key, data.value)
	return &resp.IntegerReply{Num: int64(len(data.value))}
}

func execStrlen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	data, ok := db.GetStringData(key)
	if !ok {
		return &resp.IntegerReply{Num: 0}
	}
	return &resp.IntegerReply{Num: int64(len(data.value))}
}
