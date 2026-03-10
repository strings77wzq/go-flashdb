package replication

import (
	"testing"
)

func TestNewReplicationManager(t *testing.T) {
	rm := NewReplicationManager()

	if rm.Role() != RoleMaster {
		t.Errorf("expected RoleMaster, got %v", rm.Role())
	}

	if rm.GetReplID() == "" {
		t.Error("expected non-empty replication ID")
	}

	if rm.GetReplOffset() != 0 {
		t.Errorf("expected offset 0, got %d", rm.GetReplOffset())
	}
}

func TestReplicaOf_NoOne(t *testing.T) {
	rm := NewReplicationManager()

	rm.ReplicaOf("slavehost", 6380)
	if rm.Role() != RoleSlave {
		t.Errorf("expected RoleSlave, got %v", rm.Role())
	}

	rm.ReplicaOf("no", 0)
	if rm.Role() != RoleMaster {
		t.Errorf("expected RoleMaster after REPLICAOF NO ONE, got %v", rm.Role())
	}
}

func TestReplicaOf(t *testing.T) {
	rm := NewReplicationManager()

	err := rm.ReplicaOf("localhost", 6380)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rm.Role() != RoleSlave {
		t.Errorf("expected RoleSlave, got %v", rm.Role())
	}

	host, port := rm.GetMasterInfo()
	if host != "localhost" {
		t.Errorf("expected host 'localhost', got %s", host)
	}
	if port != 6380 {
		t.Errorf("expected port 6380, got %d", port)
	}
}

func TestIncrementOffset(t *testing.T) {
	rm := NewReplicationManager()

	rm.IncrementOffset(1)
	if rm.GetReplOffset() != 1 {
		t.Errorf("expected offset 1, got %d", rm.GetReplOffset())
	}

	rm.IncrementOffset(5)
	if rm.GetReplOffset() != 6 {
		t.Errorf("expected offset 6, got %d", rm.GetReplOffset())
	}
}

func TestAddCommand(t *testing.T) {
	rm := NewReplicationManager()

	cmd := []byte("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n")
	rm.AddCommand(cmd)

	if rm.GetReplOffset() != 1 {
		t.Errorf("expected offset 1, got %d", rm.GetReplOffset())
	}

	commands, _ := rm.GetCommandsSince(0)
	if len(commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(commands))
	}
}

func TestCommandBuffer(t *testing.T) {
	cb := NewCommandBuffer(5)

	for i := 0; i < 10; i++ {
		cb.Add([]byte{byte(i)}, int64(i))
	}

	if len(cb.commands) > 5 {
		t.Errorf("expected max 5 commands, got %d", len(cb.commands))
	}

	commands, offset := cb.GetSince(7)
	if len(commands) != 2 {
		t.Errorf("expected 2 commands since offset 7, got %d", len(commands))
	}
	if offset != 9 {
		t.Errorf("expected last offset 9, got %d", offset)
	}
}

func TestSlaveManagement(t *testing.T) {
	rm := NewReplicationManager()

	count := rm.GetSlaveCount()
	if count != 0 {
		t.Errorf("expected 0 slaves, got %d", count)
	}
}

func TestGetInfo_Master(t *testing.T) {
	rm := NewReplicationManager()

	info := rm.GetInfo()

	if info["role"] != "master" {
		t.Errorf("expected role 'master', got %v", info["role"])
	}

	if info["master_replid"] == "" {
		t.Error("expected non-empty master_replid")
	}
}

func TestGetInfo_Slave(t *testing.T) {
	rm := NewReplicationManager()
	rm.ReplicaOf("localhost", 6380)

	info := rm.GetInfo()

	if info["role"] != "slave" {
		t.Errorf("expected role 'slave', got %v", info["role"])
	}

	if info["master_host"] != "localhost" {
		t.Errorf("expected master_host 'localhost', got %v", info["master_host"])
	}
}

func TestRoleString(t *testing.T) {
	if RoleMaster.String() != "master" {
		t.Errorf("expected 'master', got %s", RoleMaster.String())
	}

	if RoleSlave.String() != "slave" {
		t.Errorf("expected 'slave', got %s", RoleSlave.String())
	}
}

func TestSlaveStateString(t *testing.T) {
	tests := []struct {
		state    SlaveState
		expected string
	}{
		{SlaveStateNone, "none"},
		{SlaveStateConnecting, "connect"},
		{SlaveStateSyncing, "sync"},
		{SlaveStateOnline, "online"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.state.String())
		}
	}
}
