package core

import (
	"sync"
)

const segmentCount = 1 << 16 // 65536个分片，锁冲突概率极低

// ConcurrentDict 分段并发字典，高性能KV存储
type ConcurrentDict struct {
	segments []*Segment
	count    int
}

// Segment 字典分片，每个分片持有独立的读写锁
type Segment struct {
	m  map[string]interface{}
	mu sync.RWMutex
}

// NewConcurrentDict 创建新的并发字典
func NewConcurrentDict() *ConcurrentDict {
	segments := make([]*Segment, segmentCount)
	for i := 0; i < segmentCount; i++ {
		segments[i] = &Segment{
			m: make(map[string]interface{}),
		}
	}
	return &ConcurrentDict{
		segments: segments,
	}
}

// hash 计算key的哈希值，使用FNV哈希算法
func hash(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}
	return hash
}

// getSegment 获取key对应的分片
func (d *ConcurrentDict) getSegment(key string) *Segment {
	h := hash(key)
	return d.segments[h%segmentCount]
}

// Get 获取key对应的值
func (d *ConcurrentDict) Get(key string) (interface{}, bool) {
	seg := d.getSegment(key)
	seg.mu.RLock()
	defer seg.mu.RUnlock()
	val, ok := seg.m[key]
	return val, ok
}

// Set 设置key的值，返回是否是新增的key
func (d *ConcurrentDict) Set(key string, val interface{}) bool {
	seg := d.getSegment(key)
	seg.mu.Lock()
	defer seg.mu.Unlock()
	_, existed := seg.m[key]
	seg.m[key] = val
	if !existed {
		d.count++
	}
	return !existed
}

// SetIfAbsent 当key不存在时设置值
func (d *ConcurrentDict) SetIfAbsent(key string, val interface{}) bool {
	seg := d.getSegment(key)
	seg.mu.Lock()
	defer seg.mu.Unlock()
	_, existed := seg.m[key]
	if existed {
		return false
	}
	seg.m[key] = val
	d.count++
	return true
}

// Delete 删除key，返回是否删除成功
func (d *ConcurrentDict) Delete(key string) bool {
	seg := d.getSegment(key)
	seg.mu.Lock()
	defer seg.mu.Unlock()
	_, existed := seg.m[key]
	if existed {
		delete(seg.m, key)
		d.count--
	}
	return existed
}

// Len 返回字典中key的数量
func (d *ConcurrentDict) Len() int {
	return d.count
}

// Clear 清空字典
func (d *ConcurrentDict) Clear() {
	for _, seg := range d.segments {
		seg.mu.Lock()
		seg.m = make(map[string]interface{})
		seg.mu.Unlock()
	}
	d.count = 0
}

// ForEach 遍历字典中的所有key-value对
func (d *ConcurrentDict) ForEach(consumer func(key string, val interface{}) bool) {
	for _, seg := range d.segments {
		seg.mu.RLock()
		for k, v := range seg.m {
			if !consumer(k, v) {
				seg.mu.RUnlock()
				return
			}
		}
		seg.mu.RUnlock()
	}
}

// Keys 返回所有key的列表
func (d *ConcurrentDict) Keys() []string {
	keys := make([]string, 0, d.count)
	d.ForEach(func(key string, val interface{}) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}
