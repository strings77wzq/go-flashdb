package extension

import "goflashdb/pkg/resp"

type OpenClaw interface {
	Execute(cmd string, args ...string) (interface{}, error)
	GetStatus() map[string]interface{}
	Configure(key string, value interface{}) error
}

type Skill interface {
	Name() string
	Commands() []string
	Execute(cmd string, args [][]byte) resp.Reply
}

type SkillRegistry struct {
	skills map[string]Skill
}

func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]Skill),
	}
}

func (r *SkillRegistry) Register(skill Skill) {
	r.skills[skill.Name()] = skill
}

func (r *SkillRegistry) Get(name string) (Skill, bool) {
	skill, ok := r.skills[name]
	return skill, ok
}

func (r *SkillRegistry) List() []string {
	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	return names
}

type MCP interface {
	GetContext(sessionID string) map[string]interface{}
	SetContext(sessionID string, ctx map[string]interface{}) error
	ClearContext(sessionID string) error
}

type MCPManager struct {
	contexts map[string]map[string]interface{}
}

func NewMCPManager() *MCPManager {
	return &MCPManager{
		contexts: make(map[string]map[string]interface{}),
	}
}

func (m *MCPManager) GetContext(sessionID string) map[string]interface{} {
	if ctx, ok := m.contexts[sessionID]; ok {
		return ctx
	}
	return make(map[string]interface{})
}

func (m *MCPManager) SetContext(sessionID string, ctx map[string]interface{}) error {
	m.contexts[sessionID] = ctx
	return nil
}

func (m *MCPManager) ClearContext(sessionID string) error {
	delete(m.contexts, sessionID)
	return nil
}

type Plugin interface {
	Init(config map[string]interface{}) error
	Start() error
	Stop() error
	Info() PluginInfo
}

type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
}

type PluginManager struct {
	plugins map[string]Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

func (pm *PluginManager) Register(plugin Plugin) error {
	if err := plugin.Init(nil); err != nil {
		return err
	}
	pm.plugins[plugin.Info().Name] = plugin
	return nil
}

func (pm *PluginManager) StartAll() error {
	for _, plugin := range pm.plugins {
		if err := plugin.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (pm *PluginManager) StopAll() error {
	for _, plugin := range pm.plugins {
		if err := plugin.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (pm *PluginManager) Get(name string) (Plugin, bool) {
	plugin, ok := pm.plugins[name]
	return plugin, ok
}

func (pm *PluginManager) List() []PluginInfo {
	infos := make([]PluginInfo, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		infos = append(infos, plugin.Info())
	}
	return infos
}
