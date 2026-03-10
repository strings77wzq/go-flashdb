package core

import (
	"strconv"

	"goflashdb/pkg/resp"
)

// ZADD key score member [score member ...]
func execZAdd(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zadd' command")
	}

	key := string(args[0])
	zsetData, exists := db.GetZSetData(key)
	if !exists {
		zsetData = NewZSetData()
		db.SetZSet(key, zsetData)
	}

	count := 0
	for i := 1; i < len(args)-1; i += 2 {
		score, err := strconv.ParseFloat(string(args[i]), 64)
		if err != nil {
			return resp.NewErrorReply("ERR value is not a valid float")
		}
		member := args[i+1]
		added := zsetData.Add(score, member)
		if added {
			count++
		}
	}

	zsetData.Sort()

	return &resp.IntegerReply{Num: int64(count)}
}

// ZRANGE key start stop [WITHSCORES]
func execZRange(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zrange' command")
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

	withScores := false
	if len(args) > 3 && string(args[3]) == "WITHSCORES" {
		withScores = true
	}

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	length := len(zsetData.members)
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

	if withScores {
		replies := make([]resp.Reply, (stop-start+1)*2)
		for i := start; i <= stop; i++ {
			replies[(i-start)*2] = resp.NewBulkReply(zsetData.members[i].value)
			replies[(i-start)*2+1] = resp.NewBulkReply([]byte(strconv.FormatFloat(zsetData.members[i].score, 'f', -1, 64)))
		}
		return &resp.ArrayReply{Replies: replies}
	}

	replies := make([]resp.Reply, stop-start+1)
	for i := start; i <= stop; i++ {
		replies[i-start] = resp.NewBulkReply(zsetData.members[i].value)
	}

	return &resp.ArrayReply{Replies: replies}
}

// ZREVRANGE key start stop [WITHSCORES]
func execZRevRange(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zrevrange' command")
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

	withScores := false
	if len(args) > 3 && string(args[3]) == "WITHSCORES" {
		withScores = true
	}

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	length := len(zsetData.members)
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

	revStart := length - 1 - stop
	revStop := length - 1 - start

	if withScores {
		replies := make([]resp.Reply, (stop-start+1)*2)
		for i := revStart; i <= revStop; i++ {
			replies[(i-revStart)*2] = resp.NewBulkReply(zsetData.members[i].value)
			replies[(i-revStart)*2+1] = resp.NewBulkReply([]byte(strconv.FormatFloat(zsetData.members[i].score, 'f', -1, 64)))
		}
		return &resp.ArrayReply{Replies: replies}
	}

	replies := make([]resp.Reply, stop-start+1)
	for i := revStart; i <= revStop; i++ {
		replies[i-revStart] = resp.NewBulkReply(zsetData.members[i].value)
	}

	return &resp.ArrayReply{Replies: replies}
}

// ZRANK key member
func execZRank(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zrank' command")
	}

	key := string(args[0])
	member := string(args[1])

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return resp.NilBulkReply
	}

	memberData, ok := zsetData.scoreMap[member]
	if !ok {
		return resp.NilBulkReply
	}

	for i, m := range zsetData.members {
		if m == memberData {
			return &resp.IntegerReply{Num: int64(i)}
		}
	}

	return resp.NilBulkReply
}

// ZREVRANK key member
func execZRevRank(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zrevrank' command")
	}

	key := string(args[0])
	member := string(args[1])

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return resp.NilBulkReply
	}

	memberData, ok := zsetData.scoreMap[member]
	if !ok {
		return resp.NilBulkReply
	}

	for i, m := range zsetData.members {
		if m == memberData {
			revRank := len(zsetData.members) - 1 - i
			return &resp.IntegerReply{Num: int64(revRank)}
		}
	}

	return resp.NilBulkReply
}

// ZSCORE key member
func execZScore(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zscore' command")
	}

	key := string(args[0])
	member := string(args[1])

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return resp.NilBulkReply
	}

	memberData, ok := zsetData.scoreMap[member]
	if !ok {
		return resp.NilBulkReply
	}

	return resp.NewBulkReply([]byte(strconv.FormatFloat(memberData.score, 'f', -1, 64)))
}

// ZCARD key
func execZCard(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zcard' command")
	}

	key := string(args[0])
	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	return &resp.IntegerReply{Num: int64(len(zsetData.members))}
}

// ZCOUNT key min max
func execZCount(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zcount' command")
	}

	key := string(args[0])
	min, err := parseScore(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR min or max is not a valid float")
	}
	max, err := parseScore(string(args[2]))
	if err != nil {
		return resp.NewErrorReply("ERR min or max is not a valid float")
	}

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	count := int64(0)
	for _, m := range zsetData.members {
		if m.score >= min && m.score <= max {
			count++
		}
	}

	return &resp.IntegerReply{Num: count}
}

// ZINCRBY key increment member
func execZIncrBy(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zincrby' command")
	}

	key := string(args[0])
	increment, err := strconv.ParseFloat(string(args[1]), 64)
	if err != nil {
		return resp.NewErrorReply("ERR value is not a valid float")
	}
	member := args[2]

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		zsetData = NewZSetData()
		db.SetZSet(key, zsetData)
	}

	memberStr := string(member)
	if existing, ok := zsetData.scoreMap[memberStr]; ok {
		existing.score += increment
		zsetData.Sort()
		return resp.NewBulkReply([]byte(strconv.FormatFloat(existing.score, 'f', -1, 64)))
	}

	zsetData.Add(increment, member)
	zsetData.Sort()

	memberData, _ := zsetData.scoreMap[memberStr]
	return resp.NewBulkReply([]byte(strconv.FormatFloat(memberData.score, 'f', -1, 64)))
}

// ZREM key member [member ...]
func execZRem(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zrem' command")
	}

	key := string(args[0])
	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return &resp.IntegerReply{Num: 0}
	}

	count := 0
	for i := 1; i < len(args); i++ {
		member := string(args[i])
		if memberData, ok := zsetData.scoreMap[member]; ok {
			delete(zsetData.scoreMap, member)
			for j := 0; j < len(zsetData.members); j++ {
				if zsetData.members[j] == memberData {
					zsetData.members = append(zsetData.members[:j], zsetData.members[j+1:]...)
					break
				}
			}
			count++
		}
	}

	if len(zsetData.members) == 0 {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
	}

	return &resp.IntegerReply{Num: int64(count)}
}

// ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count]
func execZRangeByScore(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return resp.NewErrorReply("ERR wrong number of arguments for 'zrangebyscore' command")
	}

	key := string(args[0])
	min, err := parseScore(string(args[1]))
	if err != nil {
		return resp.NewErrorReply("ERR min or max is not a valid float")
	}
	max, err := parseScore(string(args[2]))
	if err != nil {
		return resp.NewErrorReply("ERR min or max is not a valid float")
	}

	withScores := false
	offset := 0
	count := -1 // -1 means unlimited

	i := 3
	for i < len(args) {
		arg := string(args[i])
		if arg == "WITHSCORES" {
			withScores = true
			i++
		} else if arg == "LIMIT" {
			if i+2 >= len(args) {
				return resp.NewErrorReply("ERR syntax error")
			}
			offset, err = parseInt(string(args[i+1]))
			if err != nil {
				return resp.NewErrorReply("ERR value is not an integer")
			}
			count, err = parseInt(string(args[i+2]))
			if err != nil {
				return resp.NewErrorReply("ERR value is not an integer")
			}
			i += 3
		} else {
			i++
		}
	}

	zsetData, exists := db.GetZSetData(key)
	if !exists {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	var members []*ZSetMember
	for _, m := range zsetData.members {
		if m.score >= min && m.score <= max {
			members = append(members, m)
		}
	}

	if offset > 0 {
		if offset >= len(members) {
			return &resp.ArrayReply{Replies: []resp.Reply{}}
		}
		members = members[offset:]
	}

	if count > 0 && count < len(members) {
		members = members[:count]
	}

	if len(members) == 0 {
		return &resp.ArrayReply{Replies: []resp.Reply{}}
	}

	if withScores {
		replies := make([]resp.Reply, len(members)*2)
		for i, m := range members {
			replies[i*2] = resp.NewBulkReply(m.value)
			replies[i*2+1] = resp.NewBulkReply([]byte(strconv.FormatFloat(m.score, 'f', -1, 64)))
		}
		return &resp.ArrayReply{Replies: replies}
	}

	replies := make([]resp.Reply, len(members))
	for i, m := range members {
		replies[i] = resp.NewBulkReply(m.value)
	}

	return &resp.ArrayReply{Replies: replies}
}

func parseScore(s string) (float64, error) {
	if s == "-inf" {
		return float64(^uint(0) >> 1), nil
	}
	if s == "+inf" || s == "inf" {
		return float64(^uint(0)), nil
	}
	return strconv.ParseFloat(s, 64)
}

func (z *ZSetData) Sort() {
	for i := 1; i < len(z.members); i++ {
		current := z.members[i]
		j := i - 1
		for j >= 0 && z.members[j].score > current.score {
			z.members[j+1] = z.members[j]
			j--
		}
		z.members[j+1] = current
	}
}

func init() {
	RegisterCommand("zadd", execZAdd, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -4)
	RegisterCommand("zrange", execZRange, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -4)
	RegisterCommand("zrevrange", execZRevRange, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -4)
	RegisterCommand("zrank", execZRank, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("zrevrank", execZRevRank, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("zscore", execZScore, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 3)
	RegisterCommand("zcard", execZCard, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 2)
	RegisterCommand("zcount", execZCount, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, 4)
	RegisterCommand("zincrby", execZIncrBy, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, 4)
	RegisterCommand("zrem", execZRem, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return []string{string(args[0])}, nil
		}
		return nil, nil
	}, -3)
	RegisterCommand("zrangebyscore", execZRangeByScore, func(args [][]byte) ([]string, []string) {
		if len(args) > 0 {
			return nil, []string{string(args[0])}
		}
		return nil, nil
	}, -4)
}
