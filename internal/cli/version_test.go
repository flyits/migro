package cli

import (
	"testing"
)

// TestVersionVariables 验证版本变量的默认值
// 需求: 版本信息变量应有合理的默认值
func TestVersionVariables(t *testing.T) {
	tests := []struct {
		name     string
		variable string
		value    string
		wantNot  string
	}{
		{
			name:     "Version has default value",
			variable: "Version",
			value:    Version,
			wantNot:  "",
		},
		{
			name:     "GitCommit has default value",
			variable: "GitCommit",
			value:    GitCommit,
			wantNot:  "",
		},
		{
			name:     "BuildDate has default value",
			variable: "BuildDate",
			value:    BuildDate,
			wantNot:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == tt.wantNot {
				t.Errorf("%s should not be empty", tt.variable)
			}
		})
	}
}

// TestVersionCmd_Registered 验证 version 命令已注册到 rootCmd
// 需求: version 命令应作为子命令注册
func TestVersionCmd_Registered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	if !found {
		t.Error("version command should be registered to rootCmd")
	}
}
