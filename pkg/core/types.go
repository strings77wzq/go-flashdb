package core

import (
	"sync"
)

type StringData struct {
	value    []byte
	expireAt int64
}

func (db *DB) GetStringData(key string) (*StringData, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	data, ok := val.(*StringData)
	if !ok {
		return nil, false
	}
	if data.expireAt > 0 && db.IsExpired(key) {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return nil, false
	}
	return data, true
}

func (db *DB) SetString(key string, value []byte) {
	db.data.Set(key, &StringData{
		value:    value,
		expireAt: 0,
	})
	db.RemoveExpire(key)
}

func (db *DB) SetStringWithExpire(key string, value []byte, expireAt int64) {
	db.data.Set(key, &StringData{
		value:    value,
		expireAt: expireAt,
	})
	db.Expire(key, expireAt)
}

type HashData struct {
	data     map[string][]byte
	expireAt int64
}

func NewHashData() *HashData {
	return &HashData{
		data:     make(map[string][]byte),
		expireAt: 0,
	}
}

func (db *DB) GetHashData(key string) (*HashData, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	data, ok := val.(*HashData)
	if !ok {
		return nil, false
	}
	if data.expireAt > 0 && db.IsExpired(key) {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return nil, false
	}
	return data, true
}

func (db *DB) SetHash(key string, hashData *HashData) {
	db.data.Set(key, hashData)
}

type ListData struct {
	items    [][]byte
	expireAt int64
}

func NewListData() *ListData {
	return &ListData{
		items:    make([][]byte, 0),
		expireAt: 0,
	}
}

func (db *DB) GetListData(key string) (*ListData, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	data, ok := val.(*ListData)
	if !ok {
		return nil, false
	}
	if data.expireAt > 0 && db.IsExpired(key) {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return nil, false
	}
	return data, true
}

func (db *DB) SetList(key string, listData *ListData) {
	db.data.Set(key, listData)
}

type SetData struct {
	members  map[string]struct{}
	expireAt int64
}

func NewSetData() *SetData {
	return &SetData{
		members:  make(map[string]struct{}),
		expireAt: 0,
	}
}

func (db *DB) GetSetData(key string) (*SetData, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	data, ok := val.(*SetData)
	if !ok {
		return nil, false
	}
	if data.expireAt > 0 && db.IsExpired(key) {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return nil, false
	}
	return data, true
}

func (db *DB) SetSet(key string, setData *SetData) {
	db.data.Set(key, setData)
}

type ZSetMember struct {
	score float64
	value []byte
}

type ZSetData struct {
	members  []*ZSetMember
	scoreMap map[string]*ZSetMember
	expireAt int64
	mu       sync.RWMutex
}

func NewZSetData() *ZSetData {
	return &ZSetData{
		members:  make([]*ZSetMember, 0),
		scoreMap: make(map[string]*ZSetMember),
		expireAt: 0,
	}
}

func (z *ZSetData) Add(score float64, value []byte) bool {
	z.mu.Lock()
	defer z.mu.Unlock()
	valueStr := string(value)
	if existing, ok := z.scoreMap[valueStr]; ok {
		existing.score = score
		return false
	}
	member := &ZSetMember{
		score: score,
		value: value,
	}
	z.members = append(z.members, member)
	z.scoreMap[valueStr] = member
	for i := len(z.members) - 1; i > 0 && z.members[i].score < z.members[i-1].score; i-- {
		z.members[i], z.members[i-1] = z.members[i-1], z.members[i]
	}
	return true
}

func (db *DB) GetZSetData(key string) (*ZSetData, bool) {
	val, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	data, ok := val.(*ZSetData)
	if !ok {
		return nil, false
	}
	if data.expireAt > 0 && db.IsExpired(key) {
		db.data.Delete(key)
		db.ttlDict.Delete(key)
		return nil, false
	}
	return data, true
}

func (db *DB) SetZSet(key string, zsetData *ZSetData) {
	db.data.Set(key, zsetData)
}
