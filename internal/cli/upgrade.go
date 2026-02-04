package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	repoOwner = "flyits"
	repoName  = "migro"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade migro to the latest version",
	Long:  `Check for updates and upgrade migro to the latest version.`,
	RunE:  runUpgrade,
}

var checkOnly bool

func init() {
	upgradeCmd.Flags().BoolVar(&checkOnly, "check", false, "only check for updates without installing")
	rootCmd.AddCommand(upgradeCmd)
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking for updates...")

	latest, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	currentVersion := strings.TrimPrefix(Version, "v")
	latestVersion := strings.TrimPrefix(latest.TagName, "v")

	if currentVersion == latestVersion {
		fmt.Printf("You are already using the latest version (%s)\n", Version)
		return nil
	}

	fmt.Printf("Current version: %s\n", Version)
	fmt.Printf("Latest version:  %s\n", latest.TagName)
	fmt.Printf("Release URL:     %s\n", latest.HTMLURL)

	if checkOnly {
		fmt.Println("\nRun 'migro upgrade' to install the latest version.")
		return nil
	}

	fmt.Println("\nUpgrading...")
	if err := doUpgrade(); err != nil {
		return fmt.Errorf("upgrade failed: %w", err)
	}

	fmt.Println("Upgrade completed successfully!")
	return nil
}

func getLatestVersion() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func doUpgrade() error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go command not found: %w", err)
	}

	installCmd := exec.Command(goPath, "install", fmt.Sprintf("github.com/%s/%s/cmd/migro@latest", repoOwner, repoName))
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	installCmd.Env = append(os.Environ(), "GOPROXY=https://goproxy.cn,direct")

	if runtime.GOOS == "windows" {
		installCmd.Env = append(installCmd.Env, "GOOS=windows")
	}

	return installCmd.Run()
}
