package core

import (
	"errors"

	"goflashdb/pkg/resp"
)

func execLPush(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'lpush' command")
	}

	key := string(args[0])
	listData, exists := db.GetListData(key)
	if !exists {
		listData = NewListData()
		db.SetList(key, listData)
	}

	for i := 1; i < len(args); i++ {
		listData.items = append([][]byte{args[i]}, listData.items...)
	}

	return &resp.IntegerReply{Num: int64(len(listData.items))}
}

func execRPush(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'rpush' command")
	}

	key := string(args[0])
	listData, exists := db.GetListData(key)
	if !exists {
		listData = NewListData()
		db.SetList(key, listData)
	}

	for i := 1; i < len(args); i++ {
		listData.items = append(listData.items, args[i])
	}

	return &resp.IntegerReply{Num: int64(len(listData.items))}
}

func execLPop(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'lpop' command")
	}

	key := string(args[0])
	listData, exists := db.GetListData(key)
	if !exists || len(listData.items) == 0 {
		return resp.NilBulkReply
	}

	count := 1
	if len(args) >= 2 {
		var err error
		count, err = parseInt(string(args[1]))
		if err != nil || count <= 0 {
			return resp.NewErrorReply("ERR value is not a positive integer")
		}
	}

	if count == 1 {
		value := listData.items[0]
		listData.items = listData.items[1:]
		if len(listData.items) == 0 {
			db.data.Delete(key)
			db.ttlDict.Delete(key)
		}
		return resp.NewBulkReply(value)
	}

	if count >= len(listData.items) {
		replies := make([]resp.Reply, len(listData.items))
		for i, item := range listData.items {
			replies[i] = resp.NewBulkReply(item)
		}
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return &resp.ArrayReply{Replies: replies}
	}

	replies := make([]resp.Reply, count)
	for i := 0; i < count; i++ {
		replies[i] = resp.NewBulkReply(listData.items[i])
	}
	listData.items = listData.items[count:]
	return &resp.ArrayReply{Replies: replies}
}

func execRPop(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'rpop' command")
	}

	key := string(args[0])
	listData, exists := db.GetListData(key)
	if !exists || len(listData.items) == 0 {
		return resp.NilBulkReply
	}

	count := 1
	if len(args) >= 2 {
		var err error
		count, err = parseInt(string(args[1]))
		if err != nil || count <= 0 {
			return resp.NewErrorReply("ERR value is not a positive integer")
		}
	}

	if count == 1 {
		idx := len(listData.items) - 1
		value := listData.items[idx]
		listData.items = listData.items[:idx]
		if len(listData.items) == 0 {
			db.data.Delete(key)
			db.ttlDict.Delete(key)
		}
		return resp.NewBulkReply(value)
	}

	if count >= len(listData.items) {
		replies := make([]resp.Reply, len(listData.items))
		for i, item := range listData.items {
			replies[len(listData.items)-1-i] = resp.NewBulkReply(item)
		}
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return &resp.ArrayReply{Replies: replies}
	}

	replies := make([]resp.Reply, count)
	for i := 0; i < count; i++ {
		idx := len(listData.items) - 1 - i
		replies[i] = resp.NewBulkReply(listData.items[idx])
	}
	listData.items = listData.items[:len(listData.items)-count]
	return &resp.ArrayReply{Replies: replies}
}

func execLRange(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'lrange' command")
	}

	key := string(args[0])
	start, err := parseInt(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	stop, err := parseInt(string(args[2]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}

	listData, exists := db.GetListData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	length := len(listData.items)
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	if start > stop || start >= length {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	replies := make([]resp.Reply, stop-start+1)
	for i := start; i <= stop; i++ {
		replies[i-start] = resp.NewBulkReply(listData.items[i])
	}

	return &resp.ArrayReply{Replies: replies}
}

func execLLen(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'llen' command")
	}

	key := string(args[0])
	listData, exists := db.GetListData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	return &resp.IntegerReply{Num: int64(len(listData.items))}
}

func execLIndex(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'lindex' command")
	}

	key := string(args[0])
	index, err := parseInt(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}

	listData, exists := db.GetListData(key)
	if !exists {
		return resp.NilBulkReply
	}

	length := len(listData.items)
	if index < 0 {
		index = length + index
	}

	if index < 0 || index >= length {
		return resp.NilBulkReply
	}

	return resp.NewBulkReply(listData.items[index])
}

func execLSet(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'lset' command")
	}

	key := string(args[0])
	index, err := parseInt(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	value := args[2]

	listData, exists := db.GetListData(key)
	if !exists {
		return resp.NewErrorReply("ERR no such key")
	}

	length := len(listData.items)
	if index < 0 {
		index = length + index
	}

	if index < 0 || index >= length {
		return resp.NewErrorReply("ERR index out of range")
	}

	listData.items[index] = value
	return resp.OkReply
}

func execLTrim(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'ltrim' command")
	}

	key := string(args[0])
	start, err := parseInt(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}
	stop, err := parseInt(string(args[2]))
	if err != nil {
		return resp.NewErrorReply("ERR value is not an integer")
	}

	listData, exists := db.GetListData(key)
	if !exists {
		return resp.OkReply
	}

	length := len(listData.items)
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	if start > stop || start >= length {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return resp.OkReply
	}

	listData.items = listData.items[start : stop+1]
	return resp.OkReply
}

func parseInt(s string) (int, error) {
	var n int
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
		n = n*10 + int(s[i]-'0')
	}

	if negative {
		n = -n
	}
	return n, nil
}

func init() {
	RegisterCommand("lpush", execLPush, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -3)
	RegisterCommand("rpush", execRPush, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -3)
	RegisterCommand("lpop", execLPop, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -2)
	RegisterCommand("rpop", execRPop, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -2)
	RegisterCommand("lrange", execLRange, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 4)
	RegisterCommand("llen", execLLen, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("lindex", execLIndex, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("lset", execLSet, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 4)
	RegisterCommand("ltrim", execLTrim, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 4)
}
