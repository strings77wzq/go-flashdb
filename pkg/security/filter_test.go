package security

import (
	"testing"
)

func TestNewCommandFilter(t *testing.T) {
	renamed := map[string]string{
		"flushall": "hidden_flushall",
	}
	filter := NewCommandFilter(renamed)
	if filter == nil {
		t.Error("NewCommandFilter should not return nil")
	}
	if !filter.enabled {
		t.Error("Filter should be enabled by default")
	}
}

func TestCommandFilterIsBlocked(t *testing.T) {
	filter := NewCommandFilter(nil)

	if !filter.IsBlocked("flushall") {
		t.Error("flushall should be blocked")
	}
	if !filter.IsBlocked("flushdb") {
		t.Error("flushdb should be blocked")
	}
	if filter.IsBlocked("get") {
		t.Error("get should not be blocked")
	}
	if filter.IsBlocked("GET") {
		t.Error("GET should be blocked (case insensitive)")
	}
}

func TestCommandFilterIsBlockedDisabled(t *testing.T) {
	filter := NewCommandFilter(nil)
	filter.enabled = false

	if filter.IsBlocked("flushall") {
		t.Error("flushall should not be blocked when filter is disabled")
	}
}

func TestCommandFilterRename(t *testing.T) {
	renamed := map[string]string{
		"flushall": "hidden_flushall",
	}
	filter := NewCommandFilter(renamed)

	result := filter.Rename("flushall")
	if result != "hidden_flushall" {
		t.Errorf("Expected 'hidden_flushall', got '%s'", result)
	}
}

func TestCommandFilterRenameNotRenamed(t *testing.T) {
	filter := NewCommandFilter(nil)

	result := filter.Rename("get")
	if result != "get" {
		t.Errorf("Expected 'get', got '%s'", result)
	}
}

func TestCommandFilterRenameDisabled(t *testing.T) {
	renamed := map[string]string{
		"flushall": "hidden_flushall",
	}
	filter := NewCommandFilter(renamed)
	filter.enabled = false

	result := filter.Rename("flushall")
	if result != "flushall" {
		t.Error("Should not rename when disabled")
	}
}

func TestCommandFilterBlockCommand(t *testing.T) {
	filter := NewCommandFilter(nil)
	filter.BlockCommand("mycommand")

	if !filter.IsBlocked("mycommand") {
		t.Error("mycommand should be blocked after BlockCommand")
	}
}

func TestCommandFilterUnblockCommand(t *testing.T) {
	filter := NewCommandFilter(nil)
	filter.UnblockCommand("flushall")

	if filter.IsBlocked("flushall") {
		t.Error("flushall should be unblocked after UnblockCommand")
	}
}

func TestCommandFilterSetEnabled(t *testing.T) {
	filter := NewCommandFilter(nil)
	filter.SetEnabled(false)

	if filter.enabled {
		t.Error("Filter should be disabled after SetEnabled(false)")
	}

	filter.SetEnabled(true)
	if !filter.enabled {
		t.Error("Filter should be enabled after SetEnabled(true)")
	}
}

func TestCommandFilterIsDangerous(t *testing.T) {
	filter := NewCommandFilter(nil)

	if !filter.IsDangerous("flushall") {
		t.Error("flushall should be dangerous")
	}
	if !filter.IsDangerous("flushdb") {
		t.Error("flushdb should be dangerous")
	}
	if !filter.IsDangerous("keys") {
		t.Error("keys should be dangerous")
	}
	if !filter.IsDangerous("debug") {
		t.Error("debug should be dangerous")
	}
	if filter.IsDangerous("get") {
		t.Error("get should not be dangerous")
	}
}

func TestCommandFilterIsDangerousCaseInsensitive(t *testing.T) {
	filter := NewCommandFilter(nil)

	if !filter.IsDangerous("FLUSHALL") {
		t.Error("FLUSHALL should be dangerous (case insensitive)")
	}
}

func TestCommandFilterGetBlockedCommands(t *testing.T) {
	filter := NewCommandFilter(nil)
	blocked := filter.GetBlockedCommands()

	if len(blocked) == 0 {
		t.Error("GetBlockedCommands should return non-empty slice")
	}

	found := false
	for _, cmd := range blocked {
		if cmd == "flushall" {
			found = true
			break
		}
	}
	if !found {
		t.Error("flushall should be in blocked commands")
	}
}

func TestCommandFilterGetRenamedCommands(t *testing.T) {
	renamed := map[string]string{
		"flushall": "hidden_flushall",
		"config":   "myconfig",
	}
	filter := NewCommandFilter(renamed)
	result := filter.GetRenamedCommands()

	if result["flushall"] != "hidden_flushall" {
		t.Error("flushall should be renamed to hidden_flushall")
	}
	if result["config"] != "myconfig" {
		t.Error("config should be renamed to myconfig")
	}
}

func TestCommandFilterGetRenamedCommandsEmpty(t *testing.T) {
	filter := NewCommandFilter(nil)
	result := filter.GetRenamedCommands()

	if result == nil {
		t.Error("GetRenamedCommands should not return nil")
	}
	if len(result) != 0 {
		t.Error("GetRenamedCommands should return empty map when no renames")
	}
}
