package extension

import (
	"testing"

	"goflashdb/pkg/resp"
)

type testSkill struct{}

func (t *testSkill) Name() string { return "test_skill" }

func (t *testSkill) Commands() []string { return []string{"test_cmd"} }

func (t *testSkill) Execute(cmd string, args [][]byte) resp.Reply {
	return resp.OkReply
}

type testPlugin struct{}

func (t *testPlugin) Init(config map[string]interface{}) error { return nil }

func (t *testPlugin) Start() error { return nil }

func (t *testPlugin) Stop() error { return nil }

func (t *testPlugin) Info() PluginInfo {
	return PluginInfo{
		Name:        "test_plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Author:      "test",
	}
}

func TestNewSkillRegistry(t *testing.T) {
	registry := NewSkillRegistry()
	if registry == nil {
		t.Error("NewSkillRegistry should not return nil")
	}
}

func TestSkillRegistryRegister(t *testing.T) {
	registry := NewSkillRegistry()
	skill := &testSkill{}
	registry.Register(skill)
}

func TestSkillRegistryGet(t *testing.T) {
	registry := NewSkillRegistry()
	skill := &testSkill{}
	registry.Register(skill)

	result, ok := registry.Get("test_skill")
	if !ok {
		t.Error("Expected to find skill")
	}
	if result.Name() != "test_skill" {
		t.Error("Expected skill name to match")
	}
}

func TestSkillRegistryGetNotFound(t *testing.T) {
	registry := NewSkillRegistry()

	_, ok := registry.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent skill")
	}
}

func TestSkillRegistryList(t *testing.T) {
	registry := NewSkillRegistry()
	registry.Register(&testSkill{})

	names := registry.List()
	if len(names) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(names))
	}
}

func TestNewMCPManager(t *testing.T) {
	mgr := NewMCPManager()
	if mgr == nil {
		t.Error("NewMCPManager should not return nil")
	}
}

func TestMCPManagerGetContext(t *testing.T) {
	mgr := NewMCPManager()

	ctx := mgr.GetContext("session1")
	if ctx == nil {
		t.Error("GetContext should not return nil")
	}
}

func TestMCPManagerSetContext(t *testing.T) {
	mgr := NewMCPManager()

	ctx := map[string]interface{}{"key": "value"}
	err := mgr.SetContext("session1", ctx)
	if err != nil {
		t.Error("SetContext should not return error")
	}

	result := mgr.GetContext("session1")
	if result["key"] != "value" {
		t.Error("Expected to retrieve set context")
	}
}

func TestMCPManagerClearContext(t *testing.T) {
	mgr := NewMCPManager()

	mgr.SetContext("session1", map[string]interface{}{"key": "value"})
	mgr.ClearContext("session1")

	ctx := mgr.GetContext("session1")
	if len(ctx) != 0 {
		t.Error("Expected context to be cleared")
	}
}

func TestNewPluginManager(t *testing.T) {
	pm := NewPluginManager()
	if pm == nil {
		t.Error("NewPluginManager should not return nil")
	}
}

func TestPluginManagerRegister(t *testing.T) {
	pm := NewPluginManager()
	plugin := &testPlugin{}

	err := pm.Register(plugin)
	if err != nil {
		t.Errorf("Register returned error: %v", err)
	}
}

func TestPluginManagerStartAll(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&testPlugin{})

	err := pm.StartAll()
	if err != nil {
		t.Errorf("StartAll returned error: %v", err)
	}
}

func TestPluginManagerStopAll(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&testPlugin{})

	err := pm.StopAll()
	if err != nil {
		t.Errorf("StopAll returned error: %v", err)
	}
}

func TestPluginManagerGet(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&testPlugin{})

	plugin, ok := pm.Get("test_plugin")
	if !ok {
		t.Error("Expected to find plugin")
	}
	if plugin.Info().Name != "test_plugin" {
		t.Error("Expected plugin name to match")
	}
}

func TestPluginManagerList(t *testing.T) {
	pm := NewPluginManager()
	pm.Register(&testPlugin{})

	infos := pm.List()
	if len(infos) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(infos))
	}
}
