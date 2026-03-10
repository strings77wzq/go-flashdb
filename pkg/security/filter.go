package security

import (
	"strings"
)

type CommandFilter struct {
	blockedCommands map[string]bool
	renamedCommands map[string]string
	enabled         bool
}

func NewCommandFilter(renamedCommands map[string]string) *CommandFilter {
	blocked := make(map[string]bool)
	blockedCommands := []string{
		"flushall",
		"flushdb",
		"debug",
		"config",
		"shutdown",
		"bgrewriteaof",
		"bgsave",
		"save",
		"slaveof",
		"replicaof",
	}

	for _, cmd := range blockedCommands {
		blocked[cmd] = true
	}

	return &CommandFilter{
		blockedCommands: blocked,
		renamedCommands: renamedCommands,
		enabled:         true,
	}
}

func (cf *CommandFilter) IsBlocked(cmd string) bool {
	if !cf.enabled {
		return false
	}
	return cf.blockedCommands[strings.ToLower(cmd)]
}

func (cf *CommandFilter) Rename(cmd string) string {
	if !cf.enabled {
		return cmd
	}
	if renamed, ok := cf.renamedCommands[strings.ToLower(cmd)]; ok {
		return renamed
	}
	return cmd
}

func (cf *CommandFilter) BlockCommand(cmd string) {
	cf.blockedCommands[strings.ToLower(cmd)] = true
}

func (cf *CommandFilter) UnblockCommand(cmd string) {
	delete(cf.blockedCommands, strings.ToLower(cmd))
}

func (cf *CommandFilter) SetEnabled(enabled bool) {
	cf.enabled = enabled
}

func (cf *CommandFilter) IsDangerous(cmd string) bool {
	dangerousCommands := []string{
		"flushall",
		"flushdb",
		"keys",
		"debug",
		"config",
		"shutdown",
	}
	cmd = strings.ToLower(cmd)
	for _, dangerous := range dangerousCommands {
		if cmd == dangerous {
			return true
		}
	}
	return false
}

func (cf *CommandFilter) GetBlockedCommands() []string {
	cmds := make([]string, 0, len(cf.blockedCommands))
	for cmd := range cf.blockedCommands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (cf *CommandFilter) GetRenamedCommands() map[string]string {
	result := make(map[string]string)
	for k, v := range cf.renamedCommands {
		result[k] = v
	}
	return result
}
