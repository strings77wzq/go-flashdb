package persist

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type BGSaveStatus int32

const (
	BGSaveNone BGSaveStatus = iota
	BGSaveRunning
	BGSaveDone
	BGSaveError
)

type EnhancedAOF struct {
	*AOFPersister
	filename     string
	rewriteMutex sync.Mutex
	rewriteCh    chan error
	lastRewrite  time.Time
	rewriteSize  int64
}

func NewEnhancedAOF(filename string, mode AOFMode) (*EnhancedAOF, error) {
	aof, err := NewAOFPersister(filename, mode)
	if err != nil {
		return nil, err
	}

	return &EnhancedAOF{
		AOFPersister: aof,
		filename:     filename,
		rewriteCh:    make(chan error, 1),
	}, nil
}

func (e *EnhancedAOF) BGREWRITE(data map[string][]byte) error {
	e.rewriteMutex.Lock()
	defer e.rewriteMutex.Unlock()

	go func() {
		err := e.Rewrite(data)
		if err != nil {
			e.rewriteCh <- err
			return
		}
		e.lastRewrite = time.Now()
		e.rewriteCh <- nil
	}()

	return nil
}

func (e *EnhancedAOF) RewriteStatus() (bool, error) {
	select {
	case err := <-e.rewriteCh:
		return false, err
	default:
		return true, nil
	}
}

type EnhancedRDB struct {
	*RDBSaver
	filename     string
	bgsaveStatus atomic.Int32
	bgsaveCh     chan error
	lastSave     time.Time
}

func NewEnhancedRDB(filename string) *EnhancedRDB {
	return &EnhancedRDB{
		RDBSaver: NewRDBSaver(filename),
		filename: filename,
		bgsaveCh: make(chan error, 1),
	}
}

func (r *EnhancedRDB) BGSAVE(data map[string][]byte, expireTimes map[string]int64) error {
	currentStatus := BGSaveStatus(r.bgsaveStatus.Load())
	if currentStatus == BGSaveRunning {
		return fmt.Errorf("BGSAVE already in progress")
	}

	r.bgsaveStatus.Store(int32(BGSaveRunning))

	go func() {
		err := r.RDBSaver.Save(data, expireTimes)
		if err != nil {
			r.bgsaveStatus.Store(int32(BGSaveError))
			r.bgsaveCh <- err
			return
		}
		r.bgsaveStatus.Store(int32(BGSaveDone))
		r.lastSave = time.Now()
		r.bgsaveCh <- nil
	}()

	return nil
}

func (r *EnhancedRDB) SaveSync(data map[string][]byte, expireTimes map[string]int64) error {
	err := r.RDBSaver.Save(data, expireTimes)
	if err != nil {
		return err
	}
	r.lastSave = time.Now()
	return nil
}

func (r *EnhancedRDB) BGSAVEStatus() (BGSaveStatus, error) {
	status := BGSaveStatus(r.bgsaveStatus.Load())

	select {
	case err := <-r.bgsaveCh:
		if err != nil {
			return BGSaveError, err
		}
		return BGSaveDone, nil
	default:
		return status, nil
	}
}

func (r *EnhancedRDB) LastSave() time.Time {
	return r.lastSave
}

type HybridPersistence struct {
	aof        *EnhancedAOF
	rdb        *EnhancedRDB
	aofEnabled bool
	rdbEnabled bool
	mu         sync.RWMutex
}

type HybridConfig struct {
	AOFile     string
	RDBFile    string
	AOFMode    AOFMode
	AOFEnabled bool
	RDBEnabled bool
}

func NewHybridPersistence(config HybridConfig) (*HybridPersistence, error) {
	h := &HybridPersistence{
		aofEnabled: config.AOFEnabled,
		rdbEnabled: config.RDBEnabled,
	}

	if config.AOFEnabled {
		aof, err := NewEnhancedAOF(config.AOFile, config.AOFMode)
		if err != nil {
			return nil, fmt.Errorf("failed to create AOF: %v", err)
		}
		h.aof = aof
	}

	if config.RDBEnabled {
		h.rdb = NewEnhancedRDB(config.RDBFile)
	}

	return h, nil
}

func (h *HybridPersistence) AppendAOF(cmd []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.aofEnabled && h.aof != nil {
		h.aof.Append(cmd)
	}
}

func (h *HybridPersistence) Save(data map[string][]byte, expireTimes map[string]int64) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.rdbEnabled && h.rdb != nil {
		return h.rdb.Save(data, expireTimes)
	}
	return nil
}

func (h *HybridPersistence) BGSAVE(data map[string][]byte, expireTimes map[string]int64) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.rdbEnabled && h.rdb != nil {
		return h.rdb.BGSAVE(data, expireTimes)
	}
	return fmt.Errorf("RDB not enabled")
}

func (h *HybridPersistence) BGREWRITE(data map[string][]byte) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.aofEnabled && h.aof != nil {
		return h.aof.BGREWRITE(data)
	}
	return fmt.Errorf("AOF not enabled")
}

func (h *HybridPersistence) Load() (map[string][]byte, map[string]int64, error) {
	data := make(map[string][]byte)
	expireTimes := make(map[string]int64)

	if h.rdbEnabled && h.rdb != nil {
		loader := NewRDBLoader(h.rdb.filename)
		if loader.Exists() {
			pairs, err := loader.Load()
			if err != nil {
				return nil, nil, err
			}
			for _, pair := range pairs {
				data[pair.Key] = pair.Value
				if pair.ExpireAt > 0 {
					expireTimes[pair.Key] = pair.ExpireAt
				}
			}
		}
	}

	if h.aofEnabled && h.aof != nil {
		commands, err := h.aof.Load()
		if err != nil {
			return nil, nil, err
		}
		for _, cmd := range commands {
			_ = cmd
		}
	}

	return data, expireTimes, nil
}

func (h *HybridPersistence) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.aof != nil {
		h.aof.Close()
	}
}

func (h *HybridPersistence) Info() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	info := make(map[string]interface{})
	info["aof_enabled"] = h.aofEnabled
	info["rdb_enabled"] = h.rdbEnabled

	if h.rdb != nil {
		status, _ := h.rdb.BGSAVEStatus()
		info["rdb_bgsave_in_progress"] = status == BGSaveRunning
		info["rdb_last_save_time"] = h.rdb.LastSave().Unix()
	}

	return info
}

func (h *HybridPersistence) CreateRDBSnapshot(filename string, data map[string][]byte, expireTimes map[string]int64) error {
	saver := NewRDBSaver(filename)
	return saver.Save(data, expireTimes)
}

func (h *HybridPersistence) LoadFromRDB(filename string) ([]KVPair, error) {
	loader := NewRDBLoader(filename)
	return loader.Load()
}

func AppendToFile(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	return writer.Flush()
}
