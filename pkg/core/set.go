package core

import (
	"goflashdb/pkg/resp"
)

func execSAdd(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'sadd' command")
	}

	key := string(args[0])
	setData, exists := db.GetSetData(key)
	if !exists {
		setData = NewSetData()
		db.SetSet(key, setData)
	}

	count := 0
	for i := 1; i < len(args); i++ {
		member := string(args[i])
		if _, ok := setData.members[member]; !ok {
			setData.members[member] = struct{}{}
			count++
		}
	}

	return &resp.IntegerReply{Num: int64(count)}
}

func execSRem(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'srem' command")
	}

	key := string(args[0])
	setData, exists := db.GetSetData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	count := 0
	for i := 1; i < len(args); i++ {
		member := string(args[i])
		if _, ok := setData.members[member]; ok {
			delete(setData.members, member)
			count++
		}
	}

	if len(setData.members) == 0 {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
	}

	return &resp.IntegerReply{Num: int64(count)}
}

func execSIsMember(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'sismember' command")
	}

	key := string(args[0])
	member := string(args[1])

	setData, exists := db.GetSetData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	if _, ok := setData.members[member]; ok {
		return &resp.IntegerReply{Num: 1}
	}
	return &resp.IntegerReply{Num: 0}
}

func execSMembers(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'smembers' command")
	}

	key := string(args[0])
	setData, exists := db.GetSetData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	replies := make([]resp.Reply, 0, len(setData.members))
	for member := range setData.members {
		replies = append(replies, resp.NewBulkReply([]byte(member)))
	}

	return &resp.ArrayReply{Replies: replies}
}

func execSCard(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'scard' command")
	}

	key := string(args[0])
	setData, exists := db.GetSetData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	return &resp.IntegerReply{Num: int64(len(setData.members))}
}

func execSPop(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'spop' command")
	}

	key := string(args[0])
	setData, exists := db.GetSetData(key)
	if !exists || len(setData.members) == 0 {
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

	if count >= len(setData.members) {
		replies := make([]resp.Reply, 0, len(setData.members))
		for member := range setData.members {
			replies = append(replies, resp.NewBulkReply([]byte(member)))
		}
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return &resp.ArrayReply{Replies: replies}
	}

	replies := make([]resp.Reply, 0, count)
	i := 0
	for member := range setData.members {
		if i >= count {
			break
		}
		replies = append(replies, resp.NewBulkReply([]byte(member)))
		delete(setData.members, member)
		i++
	}

	return &resp.ArrayReply{Replies: replies}
}

func execSRandMember(db *DB, args [][]byte) resp.Reply {
	if len(args) < 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'srandmember' command")
	}

	key := string(args[0])
	setData, exists := db.GetSetData(key)
	if !exists || len(setData.members) == 0 {
		return resp.NilBulkReply
	}

	count := 1
	if len(args) >= 2 {
		var err error
		count, err = parseInt(string(args[1]))
		if err != nil {
			return resp.NewErrorReply("ERR value is not an integer")
		}
	}

	if count == 0 {
		return resp.NilBulkReply
	}

	members := make([]string, 0, len(setData.members))
	for m := range setData.members {
		members = append(members, m)
	}

	if count == 1 {
		return resp.NewBulkReply([]byte(members[0]))
	}

	if count < 0 {
		count = -count
		replies := make([]resp.Reply, count)
		for i := 0; i < count; i++ {
			replies[i] = resp.NewBulkReply([]byte(members[i%len(members)]))
		}
		return &resp.ArrayReply{Replies: replies}
	}

	if count > len(members) {
		count = len(members)
	}
	replies := make([]resp.Reply, count)
	for i := 0; i < count; i++ {
		replies[i] = resp.NewBulkReply([]byte(members[i]))
	}
	return &resp.ArrayReply{Replies: replies}
}

func init() {
	RegisterCommand("sadd", execSAdd, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -3)
	RegisterCommand("srem", execSRem, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -3)
	RegisterCommand("sismember", execSIsMember, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("smembers", execSMembers, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("scard", execSCard, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("spop", execSPop, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -2)
	RegisterCommand("srandmember", execSRandMember, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -2)
}
