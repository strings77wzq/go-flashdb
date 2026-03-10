package persist

import (
	"os"
	"testing"
	"time"
)

func TestEnhancedRDB_BGSAVE(t *testing.T) {
	filename := "/tmp/test_enhanced_rdb.rdb"
	defer os.Remove(filename)

	rdb := NewEnhancedRDB(filename)

	data := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}
	expireTimes := map[string]int64{
		"key1": time.Now().Add(time.Hour).Unix(),
	}

	err := rdb.BGSAVE(data, expireTimes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status, err := rdb.BGSAVEStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 100 && status == BGSaveRunning; i++ {
		time.Sleep(10 * time.Millisecond)
		status, _ = rdb.BGSAVEStatus()
	}

	if status == BGSaveRunning {
		t.Error("BGSAVE should have completed")
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("RDB file should exist after BGSAVE")
	}
}

func TestEnhancedRDB_BGSAVE_AlreadyRunning(t *testing.T) {
	filename := "/tmp/test_enhanced_rdb2.rdb"
	defer os.Remove(filename)

	rdb := NewEnhancedRDB(filename)

	rdb.bgsaveStatus.Store(int32(BGSaveRunning))

	err := rdb.BGSAVE(nil, nil)
	if err == nil {
		t.Error("expected error for BGSAVE already in progress")
	}
}

func TestEnhancedRDB_LastSave(t *testing.T) {
	filename := "/tmp/test_enhanced_rdb3.rdb"
	defer os.Remove(filename)

	rdb := NewEnhancedRDB(filename)

	before := rdb.LastSave()

	rdb.SaveSync(map[string][]byte{"key": []byte("value")}, nil)

	after := rdb.LastSave()

	if !after.After(before) {
		t.Error("LastSave should be updated after save")
	}
}

func TestHybridPersistence_Create(t *testing.T) {
	aofFile := "/tmp/test_hybrid.aof"
	rdbFile := "/tmp/test_hybrid.rdb"
	defer os.Remove(aofFile)
	defer os.Remove(rdbFile)

	config := HybridConfig{
		AOFile:     aofFile,
		RDBFile:    rdbFile,
		AOFMode:    AOFEverysec,
		AOFEnabled: true,
		RDBEnabled: true,
	}

	h, err := NewHybridPersistence(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer h.Close()

	info := h.Info()
	if !info["aof_enabled"].(bool) {
		t.Error("AOF should be enabled")
	}
	if !info["rdb_enabled"].(bool) {
		t.Error("RDB should be enabled")
	}
}

func TestHybridPersistence_AOFOnly(t *testing.T) {
	aofFile := "/tmp/test_hybrid_aof_only.aof"
	defer os.Remove(aofFile)

	config := HybridConfig{
		AOFile:     aofFile,
		AOFMode:    AOFEverysec,
		AOFEnabled: true,
		RDBEnabled: false,
	}

	h, err := NewHybridPersistence(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer h.Close()

	h.AppendAOF([]byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"))

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(aofFile); os.IsNotExist(err) {
		t.Error("AOF file should exist")
	}
}

func TestHybridPersistence_RDBOnly(t *testing.T) {
	rdbFile := "/tmp/test_hybrid_rdb_only.rdb"
	defer os.Remove(rdbFile)

	config := HybridConfig{
		RDBFile:    rdbFile,
		AOFEnabled: false,
		RDBEnabled: true,
	}

	h, err := NewHybridPersistence(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer h.Close()

	data := map[string][]byte{"key": []byte("value")}
	err = h.Save(data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(rdbFile); os.IsNotExist(err) {
		t.Error("RDB file should exist")
	}
}

func TestHybridPersistence_BGSAVE(t *testing.T) {
	rdbFile := "/tmp/test_hybrid_bgsave.rdb"
	defer os.Remove(rdbFile)

	config := HybridConfig{
		RDBFile:    rdbFile,
		AOFEnabled: false,
		RDBEnabled: true,
	}

	h, err := NewHybridPersistence(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer h.Close()

	data := map[string][]byte{"key": []byte("value")}
	err = h.BGSAVE(data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if _, err := os.Stat(rdbFile); os.IsNotExist(err) {
		t.Error("RDB file should exist after BGSAVE")
	}
}

func TestHybridPersistence_CreateRDBSnapshot(t *testing.T) {
	rdbFile := "/tmp/test_snapshot.rdb"
	defer os.Remove(rdbFile)

	config := HybridConfig{
		RDBEnabled: true,
	}

	h, err := NewHybridPersistence(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer h.Close()

	data := map[string][]byte{"key": []byte("value")}
	err = h.CreateRDBSnapshot(rdbFile, data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(rdbFile); os.IsNotExist(err) {
		t.Error("Snapshot file should exist")
	}
}

func TestHybridPersistence_LoadFromRDB(t *testing.T) {
	rdbFile := "/tmp/test_load.rdb"
	defer os.Remove(rdbFile)

	config := HybridConfig{
		RDBFile:    rdbFile,
		RDBEnabled: true,
	}

	h, err := NewHybridPersistence(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer h.Close()

	data := map[string][]byte{"key1": []byte("value1"), "key2": []byte("value2")}
	expireTimes := map[string]int64{"key1": time.Now().Add(time.Hour).Unix()}

	err = h.Save(data, expireTimes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pairs, err := h.LoadFromRDB(rdbFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(pairs) != 2 {
		t.Errorf("expected 2 pairs, got %d", len(pairs))
	}
}

func TestAppendToFile(t *testing.T) {
	filename := "/tmp/test_append.txt"
	defer os.Remove(filename)

	err := AppendToFile(filename, []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = AppendToFile(filename, []byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got %s", string(content))
	}
}
