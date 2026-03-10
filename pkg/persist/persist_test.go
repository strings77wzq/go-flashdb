package persist

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRDBEncoderEncodeHeader(t *testing.T) {
	encoder := NewRDBEncoder()
	header := encoder.EncodeHeader()

	if string(header[:5]) != RDBMagicString {
		t.Errorf("Expected magic string '%s', got '%s'", RDBMagicString, string(header[:5]))
	}
	if string(header[5:]) != RDBVersion {
		t.Errorf("Expected version '%s', got '%s'", RDBVersion, string(header[5:]))
	}
}

func TestRDBEncoderEncodeString(t *testing.T) {
	encoder := NewRDBEncoder()
	encoded := encoder.EncodeString("key", []byte("value"), 0)

	if len(encoded) == 0 {
		t.Error("Encoded string should not be empty")
	}

	if encoded[0] != 0x00 {
		t.Error("Expected opcode 0x00 for non-expired key")
	}
}

func TestRDBEncoderEncodeStringWithExpiry(t *testing.T) {
	encoder := NewRDBEncoder()
	expireAt := time.Now().Add(time.Hour).UnixMilli()
	encoded := encoder.EncodeString("key", []byte("value"), expireAt)

	if encoded[0] != 0xFC {
		t.Error("Expected opcode 0xFC for expired key")
	}
}

func TestRDBEncoderEncodeFooter(t *testing.T) {
	encoder := NewRDBEncoder()
	footer := encoder.EncodeFooter()

	if len(footer) != 1 {
		t.Errorf("Expected footer length 1, got %d", len(footer))
	}
	if footer[0] != 0xFF {
		t.Error("Expected footer 0xFF")
	}
}

func TestRDBEncoderEncodeLength(t *testing.T) {
	encoder := NewRDBEncoder()

	buf := encoder.encodeLength(nil, 10)
	if len(buf) != 1 || buf[0] != 10 {
		t.Error("Length < 64 should be encoded as single byte")
	}

	buf = encoder.encodeLength(nil, 100)
	if len(buf) != 2 || buf[0]&0x40 == 0 {
		t.Error("Length < 16384 should have high bit set")
	}

	buf = encoder.encodeLength(nil, 20000)
	if len(buf) != 5 || buf[0] != 0x80 {
		t.Error("Length >= 16384 should have 0x80 prefix")
	}
}

func TestNewRDBSaver(t *testing.T) {
	saver := NewRDBSaver("test.rdb")
	if saver == nil {
		t.Error("NewRDBSaver should not return nil")
	}
	if saver.encoder == nil {
		t.Error("encoder should not be nil")
	}
}

func TestRDBSaverSave(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.rdb")
	saver := NewRDBSaver(tmpFile)

	data := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}
	expireTimes := map[string]int64{
		"key1": time.Now().Add(time.Hour).UnixMilli(),
	}

	err := saver.Save(data, expireTimes)
	if err != nil {
		t.Errorf("Save returned error: %v", err)
	}

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("RDB file should be created")
	}
}

func TestRDBSaverSaveEmpty(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.rdb")
	saver := NewRDBSaver(tmpFile)

	err := saver.Save(map[string][]byte{}, map[string]int64{})
	if err != nil {
		t.Errorf("Save returned error: %v", err)
	}
}

func TestNewRDBLoader(t *testing.T) {
	loader := NewRDBLoader("test.rdb")
	if loader == nil {
		t.Error("NewRDBLoader should not return nil")
	}
	if loader.filename != "test.rdb" {
		t.Errorf("Expected filename 'test.rdb', got '%s'", loader.filename)
	}
}

func TestRDBLoaderLoadNotExist(t *testing.T) {
	loader := NewRDBLoader("/nonexistent/path/test.rdb")
	pairs, err := loader.Load()
	if err != nil {
		t.Errorf("Load returned error: %v", err)
	}
	if pairs != nil {
		t.Error("Load should return nil for non-existent file")
	}
}

func TestRDBLoaderLoad(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.rdb")
	saver := NewRDBSaver(tmpFile)

	data := map[string][]byte{
		"mykey": []byte("myvalue"),
	}
	err := saver.Save(data, map[string]int64{})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loader := NewRDBLoader(tmpFile)
	pairs, err := loader.Load()
	if err != nil {
		t.Errorf("Load returned error: %v", err)
	}
	if len(pairs) != 1 {
		t.Errorf("Expected 1 pair, got %d", len(pairs))
	}
	if pairs[0].Key != "mykey" {
		t.Errorf("Expected key 'mykey', got '%s'", pairs[0].Key)
	}
	if string(pairs[0].Value) != "myvalue" {
		t.Errorf("Expected value 'myvalue', got '%s'", string(pairs[0].Value))
	}
}

func TestRDBLoaderExists(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "exists.rdb")

	loader := NewRDBLoader(tmpFile)
	if loader.Exists() {
		t.Error("Should not exist before file creation")
	}

	_, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	loader = NewRDBLoader(tmpFile)
	if !loader.Exists() {
		t.Error("Should exist after file creation")
	}
}

func TestNewAOFPersister(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.aof")
	persister, err := NewAOFPersister(tmpFile, AOFEverysec)
	if err != nil {
		t.Errorf("NewAOFPersister returned error: %v", err)
	}
	if persister == nil {
		t.Error("persister should not be nil")
	}
	persister.Close()
}

func TestAOFPersisterAppend(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.aof")
	persister, err := NewAOFPersister(tmpFile, AOFAlways)
	if err != nil {
		t.Fatalf("NewAOFPersister returned error: %v", err)
	}

	persister.Append([]byte("test command"))

	time.Sleep(50 * time.Millisecond)

	persister.Close()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to read AOF file: %v", err)
	}
	if len(data) == 0 {
		t.Error("AOF file should not be empty")
	}
}

func TestAOFPersisterLoad(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_load.aof")
	persister, err := NewAOFPersister(tmpFile, AOFAlways)
	if err != nil {
		t.Fatalf("NewAOFPersister returned error: %v", err)
	}

	persister.Append([]byte("test command 1\r\n"))
	time.Sleep(50 * time.Millisecond)
	persister.Close()

	loader, err := NewAOFPersister(tmpFile, AOFAlways)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	commands, err := loader.Load()
	if err != nil {
		t.Errorf("Load returned error: %v", err)
	}
	if len(commands) == 0 {
		t.Error("Should have loaded at least one command")
	}
	loader.Close()
}

func TestAOFPersisterLoadNotExist(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "nonexistent.aof")
	persister, err := NewAOFPersister(tmpFile, AOFAlways)
	if err != nil {
		t.Fatalf("NewAOFPersister returned error: %v", err)
	}

	commands, err := persister.Load()
	if err != nil {
		t.Errorf("Load returned error: %v", err)
	}
	if commands != nil {
		t.Error("Should return nil for non-existent file")
	}
	persister.Close()
}

func TestAOFPersisterRewrite(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_rewrite.aof")
	persister, err := NewAOFPersister(tmpFile, AOFAlways)
	if err != nil {
		t.Fatalf("NewAOFPersister returned error: %v", err)
	}

	data := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}

	err = persister.Rewrite(data)
	if err != nil {
		t.Errorf("Rewrite returned error: %v", err)
	}

	persister.Close()

	data2, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("Failed to read AOF file: %v", err)
	}
	if len(data2) == 0 {
		t.Error("Rewritten AOF file should not be empty")
	}
}

func TestNewPersistManager(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, false, time.Minute)
	if err != nil {
		t.Errorf("NewPersistManager returned error: %v", err)
	}
	if pm == nil {
		t.Error("pm should not be nil")
	}
	pm.Close()
}

func TestNewPersistManagerWithAOF(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, true, time.Minute)
	if err != nil {
		t.Errorf("NewPersistManager returned error: %v", err)
	}
	if !pm.aofEnabled {
		t.Error("AOF should be enabled")
	}
	pm.Close()
}

func TestPersistManagerAppendAOF(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, true, time.Minute)
	if err != nil {
		t.Fatalf("NewPersistManager returned error: %v", err)
	}

	pm.AppendAOF([]byte("test"))

	time.Sleep(50 * time.Millisecond)

	pm.Close()

	data, err := os.ReadFile(aofFile)
	if err != nil {
		t.Errorf("Failed to read AOF file: %v", err)
	}
	if len(data) == 0 {
		t.Error("AOF file should have data")
	}
}

func TestPersistManagerAppendAOFDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test_disabled.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, false, time.Minute)
	if err != nil {
		t.Fatalf("NewPersistManager returned error: %v", err)
	}

	pm.AppendAOF([]byte("test"))

	pm.Close()

	if _, err := os.Stat(aofFile); os.IsNotExist(err) {
		// Expected - AOF disabled should not create file
	} else {
		t.Error("AOF file should not be created when disabled")
	}
}

func TestPersistManagerSaveRDB(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, false, time.Minute)
	if err != nil {
		t.Fatalf("NewPersistManager returned error: %v", err)
	}

	data := map[string][]byte{
		"key1": []byte("value1"),
	}
	expireTimes := map[string]int64{}

	err = pm.SaveRDB(data, expireTimes)
	if err != nil {
		t.Errorf("SaveRDB returned error: %v", err)
	}

	pm.Close()

	if _, err := os.Stat(rdbFile); os.IsNotExist(err) {
		t.Error("RDB file should be created")
	}
}

func TestPersistManagerLoadRDB(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, false, time.Minute)
	if err != nil {
		t.Fatalf("NewPersistManager returned error: %v", err)
	}

	data := map[string][]byte{
		"loadkey": []byte("loadvalue"),
	}
	pm.SaveRDB(data, map[string]int64{})

	pairs, err := pm.LoadRDB()
	if err != nil {
		t.Errorf("LoadRDB returned error: %v", err)
	}
	if len(pairs) != 1 {
		t.Errorf("Expected 1 pair, got %d", len(pairs))
	}

	pm.Close()
}

func TestPersistManagerLoadAOF(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, true, time.Minute)
	if err != nil {
		t.Fatalf("NewPersistManager returned error: %v", err)
	}

	pm.AppendAOF([]byte("cmd1\r\n"))
	time.Sleep(50 * time.Millisecond)

	commands, err := pm.LoadAOF()
	if err != nil {
		t.Errorf("LoadAOF returned error: %v", err)
	}
	if len(commands) == 0 {
		t.Error("Should have loaded commands")
	}

	pm.Close()
}

func TestPersistManagerLoadAOFDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	aofFile := tmpDir + "/test.aof"
	rdbFile := tmpDir + "/test.rdb"

	pm, err := NewPersistManager(aofFile, rdbFile, false, time.Minute)
	if err != nil {
		t.Fatalf("NewPersistManager returned error: %v", err)
	}

	commands, err := pm.LoadAOF()
	if err != nil {
		t.Errorf("LoadAOF returned error: %v", err)
	}
	if commands != nil {
		t.Error("Should return nil when AOF disabled")
	}

	pm.Close()
}
