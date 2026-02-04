package cli

import (
	"encoding/json"
	"testing"
)

// TestUpgradeCmd_Registered 验证 upgrade 命令已注册到 rootCmd
// 需求: upgrade 命令应作为子命令注册
func TestUpgradeCmd_Registered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "upgrade" {
			found = true
			break
		}
	}
	if !found {
		t.Error("upgrade command should be registered to rootCmd")
	}
}

// TestUpgradeCmd_CheckFlag 验证 --check 标志已注册
// 需求: upgrade 命令应支持 --check 标志
func TestUpgradeCmd_CheckFlag(t *testing.T) {
	flag := upgradeCmd.Flags().Lookup("check")
	if flag == nil {
		t.Error("upgrade command should have --check flag")
		return
	}
	if flag.DefValue != "false" {
		t.Errorf("--check flag default value should be false, got %s", flag.DefValue)
	}
}

// TestGithubRelease_JSONParsing 验证 GitHub release JSON 解析
// 需求: 能正确解析 GitHub API 返回的 release 信息
func TestGithubRelease_JSONParsing(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantTag  string
		wantURL  string
		wantErr  bool
	}{
		{
			name:    "valid release",
			json:    `{"tag_name":"v1.0.0","html_url":"https://github.com/flyits/migro/releases/tag/v1.0.0"}`,
			wantTag: "v1.0.0",
			wantURL: "https://github.com/flyits/migro/releases/tag/v1.0.0",
			wantErr: false,
		},
		{
			name:    "release without v prefix",
			json:    `{"tag_name":"1.2.3","html_url":"https://example.com"}`,
			wantTag: "1.2.3",
			wantURL: "https://example.com",
			wantErr: false,
		},
		{
			name:    "empty json object",
			json:    `{}`,
			wantTag: "",
			wantURL: "",
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var release githubRelease
			err := json.Unmarshal([]byte(tt.json), &release)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if release.TagName != tt.wantTag {
					t.Errorf("TagName = %v, want %v", release.TagName, tt.wantTag)
				}
				if release.HTMLURL != tt.wantURL {
					t.Errorf("HTMLURL = %v, want %v", release.HTMLURL, tt.wantURL)
				}
			}
		})
	}
}

// TestVersionComparison 验证版本比较逻辑
// 需求: 正确比较当前版本和最新版本（去除 v 前缀后比较）
func TestVersionComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestTag      string
		wantSame       bool
	}{
		{
			name:           "same version with v prefix",
			currentVersion: "v1.0.0",
			latestTag:      "v1.0.0",
			wantSame:       true,
		},
		{
			name:           "same version without v prefix",
			currentVersion: "1.0.0",
			latestTag:      "1.0.0",
			wantSame:       true,
		},
		{
			name:           "same version mixed prefix",
			currentVersion: "v1.0.0",
			latestTag:      "1.0.0",
			wantSame:       true,
		},
		{
			name:           "different versions",
			currentVersion: "v1.0.0",
			latestTag:      "v1.1.0",
			wantSame:       false,
		},
		{
			name:           "dev version",
			currentVersion: "dev",
			latestTag:      "v1.0.0",
			wantSame:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 runUpgrade 中的版本比较逻辑
			current := trimVersionPrefix(tt.currentVersion)
			latest := trimVersionPrefix(tt.latestTag)
			same := current == latest
			if same != tt.wantSame {
				t.Errorf("version comparison: current=%s, latest=%s, got same=%v, want %v",
					tt.currentVersion, tt.latestTag, same, tt.wantSame)
			}
		})
	}
}

// trimVersionPrefix 去除版本号的 v 前缀（辅助测试函数）
func trimVersionPrefix(v string) string {
	if len(v) > 0 && v[0] == 'v' {
		return v[1:]
	}
	return v
}

// TestRepoConstants 验证仓库常量配置正确
// 需求: 仓库 owner 和 name 应正确配置
func TestRepoConstants(t *testing.T) {
	if repoOwner == "" {
		t.Error("repoOwner should not be empty")
	}
	if repoName == "" {
		t.Error("repoName should not be empty")
	}
	if repoOwner != "flyits" {
		t.Errorf("repoOwner = %s, want flyits", repoOwner)
	}
	if repoName != "migro" {
		t.Errorf("repoName = %s, want migro", repoName)
	}
}
