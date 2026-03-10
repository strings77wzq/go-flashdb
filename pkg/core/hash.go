package core

import (
	"goflashdb/pkg/resp"
)

func execHSet(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hset' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)
	if !exists {
		hashData = NewHashData()
		db.SetHash(key, hashData)
	}

	count := 0
	for i := 1; i < len(args)-1; i += 2 {
		field := string(args[i])
		value := args[i+1]
		if _, ok := hashData.data[field]; !ok {
			count++
		}
		hashData.data[field] = value
	}

	return &resp.IntegerReply{Num: int64(count)}
}

func execHGet(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hget' command")
	}

	key := string(args[0])
	field := string(args[1])

	hashData, exists := db.GetHashData(key)
	if !exists {
		return resp.NilBulkReply
	}

	value, ok := hashData.data[field]
	if !ok {
		return resp.NilBulkReply
	}

	return resp.NewBulkReply(value)
}

func execHDel(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hdel' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	count := 0
	for i := 1; i < len(args); i++ {
		field := string(args[i])
		if _, ok := hashData.data[field]; ok {
			delete(hashData.data, field)
			count++
		}
	}

	if len(hashData.data) == 0 {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
	}

	return &resp.IntegerReply{Num: int64(count)}
}

func execHMGet(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hmget' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)

	replies := make([]resp.Reply, len(args)-1)
	for i := 1; i < len(args); i++ {
		field := string(args[i])
		if !exists {
			replies[i-1] = resp.NilBulkReply
			continue
		}
		value, ok := hashData.data[field]
		if !ok {
			replies[i-1] = resp.NilBulkReply
			continue
		}
		replies[i-1] = resp.NewBulkReply(value)
	}

	return &resp.ArrayReply{Replies: replies}
}

func execHGetAll(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hgetall' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	replies := make([]resp.Reply, 0, len(hashData.data)*2)
	for field, value := range hashData.data {
		replies = append(replies, resp.NewBulkReply([]byte(field)))
		replies = append(replies, resp.NewBulkReply(value))
	}

	return &resp.ArrayReply{Replies: replies}
}

func execHExists(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hexists' command")
	}

	key := string(args[0])
	field := string(args[1])

	hashData, exists := db.GetHashData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	_, ok := hashData.data[field]
	if ok {
		return &resp.IntegerReply{Num: 1}
	}
	return &resp.IntegerReply{Num: 0}
}

func execHLen(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hlen' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	return &resp.IntegerReply{Num: int64(len(hashData.data))}
}

func execHKeys(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hkeys' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	replies := make([]resp.Reply, 0, len(hashData.data))
	for field := range hashData.data {
		replies = append(replies, resp.NewBulkReply([]byte(field)))
	}

	return &resp.ArrayReply{Replies: replies}
}

func execHVals(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'hvals' command")
	}

	key := string(args[0])
	hashData, exists := db.GetHashData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	replies := make([]resp.Reply, 0, len(hashData.data))
	for _, value := range hashData.data {
		replies = append(replies, resp.NewBulkReply(value))
	}

	return &resp.ArrayReply{Replies: replies}
}

func init() {
	RegisterCommand("hset", execHSet, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -4)
	RegisterCommand("hget", execHGet, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("hdel", execHDel, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -3)
	RegisterCommand("hmget", execHMGet, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -3)
	RegisterCommand("hgetall", execHGetAll, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("hexists", execHExists, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("hlen", execHLen, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("hkeys", execHKeys, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("hvals", execHVals, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
}
