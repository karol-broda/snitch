package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/karol-broda/snitch/internal/errutil"
	"github.com/karol-broda/snitch/internal/tui"
)

const (
	repoOwner = "karol-broda"
	repoName  = "snitch"
	githubAPI = "https://api.github.com"
	firstUpgradeVersion = "0.1.8"
)

var (
	upgradeYes     bool
	upgradeVersion string
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Check for updates and optionally upgrade snitch",
	Long: `Check for available updates and show upgrade instructions.

Use --yes to perform an in-place upgrade automatically.
Use --version to install a specific version.`,
	RunE: runUpgrade,
}

func init() {
	upgradeCmd.Flags().BoolVarP(&upgradeYes, "yes", "y", false, "Perform the upgrade automatically")
	upgradeCmd.Flags().StringVarP(&upgradeVersion, "version", "v", "", "Install a specific version (e.g., v0.1.7)")
	rootCmd.AddCommand(upgradeCmd)
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type githubCommit struct {
	SHA string `json:"sha"`
}

type githubCompare struct {
	Status      string `json:"status"`
	AheadBy     int    `json:"ahead_by"`
	BehindBy    int    `json:"behind_by"`
	TotalCommits int   `json:"total_commits"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	current := Version
	nixInstall := isNixInstall()
	nixVersion := isNixVersion(current)

	if upgradeVersion != "" {
		if nixInstall || nixVersion {
			return handleNixSpecificVersion(current, upgradeVersion)
		}
		return handleSpecificVersion(current, upgradeVersion)
	}

	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if nixInstall || nixVersion {
		return handleNixUpgrade(current, latest)
	}

	currentClean := strings.TrimPrefix(current, "v")
	latestClean := strings.TrimPrefix(latest, "v")

	printVersionComparison(current, latest)

	if currentClean == latestClean {
		green := color.New(color.FgGreen)
		errutil.Println(green, tui.SymbolSuccess+" you are running the latest version")
		return nil
	}

	if current == "dev" {
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" you are running a development build")
		fmt.Println()
		fmt.Println("use one of the methods below to install a release version:")
		fmt.Println()
		printUpgradeInstructions()
		return nil
	}

	green := color.New(color.FgGreen, color.Bold)
	errutil.Printf(green, tui.SymbolSuccess+" update available: %s "+tui.SymbolArrowRight+" %s\n", current, latest)
	fmt.Println()

	if !upgradeYes {
		printUpgradeInstructions()
		fmt.Println()
		faint := color.New(color.Faint)
		cmdStyle := color.New(color.FgCyan)
		errutil.Print(faint, "  in-place      ")
		errutil.Println(cmdStyle, "snitch upgrade --yes")
		return nil
	}

	return performUpgrade(latest)
}

func handleSpecificVersion(current, target string) error {
	if !strings.HasPrefix(target, "v") {
		target = "v" + target
	}
	targetClean := strings.TrimPrefix(target, "v")

	printVersionComparisonTarget(current, target)

	if isVersionLower(targetClean, firstUpgradeVersion) {
		yellow := color.New(color.FgYellow)
		errutil.Printf(yellow, tui.SymbolWarning+" warning: the upgrade command was introduced in v%s\n", firstUpgradeVersion)
		faint := color.New(color.Faint)
		errutil.Printf(faint, "  version %s does not include this command\n", target)
		errutil.Println(faint, "  you will need to use other methods to upgrade from that version")
		fmt.Println()
	}

	currentClean := strings.TrimPrefix(current, "v")
	if currentClean == targetClean {
		green := color.New(color.FgGreen)
		errutil.Println(green, tui.SymbolSuccess+" you are already running this version")
		return nil
	}

	if !upgradeYes {
		faint := color.New(color.Faint)
		cmdStyle := color.New(color.FgCyan)
		if isVersionLower(targetClean, currentClean) {
			yellow := color.New(color.FgYellow)
			errutil.Printf(yellow, tui.SymbolArrowDown+" this will downgrade from %s to %s\n", current, target)
		} else {
			green := color.New(color.FgGreen)
			errutil.Printf(green, tui.SymbolArrowUp+" this will upgrade from %s to %s\n", current, target)
		}
		fmt.Println()
		errutil.Print(faint, "run ")
		errutil.Printf(cmdStyle, "snitch upgrade --version %s --yes", target)
		errutil.Println(faint, " to proceed")
		return nil
	}

	return performUpgrade(target)
}

func handleNixUpgrade(current, latest string) error {
	faint := color.New(color.Faint)
	version := color.New(color.FgCyan)

	currentCommit := extractCommitFromVersion(current)
	dirty := isNixDirty(current)

	errutil.Print(faint, "current  ")
	errutil.Print(version, current)
	if currentCommit != "" {
		errutil.Printf(faint, " (commit %s)", currentCommit)
	}
	fmt.Println()

	errutil.Print(faint, "latest   ")
	errutil.Println(version, latest)
	fmt.Println()

	if dirty {
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" you are running a dirty nix build (uncommitted changes)")
		fmt.Println()
		printNixUpgradeInstructions()
		return nil
	}

	if currentCommit == "" {
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" this is a nix installation")
		errutil.Println(faint, "  nix store is immutable; use nix commands to upgrade")
		fmt.Println()
		printNixUpgradeInstructions()
		return nil
	}

	releaseCommit, err := fetchCommitForTag(latest)
	if err != nil {
		errutil.Printf(faint, "  (could not fetch release commit: %v)\n", err)
		fmt.Println()
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" this is a nix installation")
		errutil.Println(faint, "  nix store is immutable; use nix commands to upgrade")
		fmt.Println()
		printNixUpgradeInstructions()
		return nil
	}

	releaseShort := releaseCommit
	if len(releaseShort) > 7 {
		releaseShort = releaseShort[:7]
	}

	if strings.HasPrefix(releaseCommit, currentCommit) || strings.HasPrefix(currentCommit, releaseShort) {
		green := color.New(color.FgGreen)
		errutil.Printf(green, tui.SymbolSuccess+" you are running %s (commit %s)\n", latest, releaseShort)
		return nil
	}

	comparison, err := compareCommits(latest, currentCommit)
	if err != nil {
		green := color.New(color.FgGreen, color.Bold)
		errutil.Printf(green, tui.SymbolSuccess+" update available: %s "+tui.SymbolArrowRight+" %s\n", currentCommit, latest)
		errutil.Printf(faint, "  your commit: %s\n", currentCommit)
		errutil.Printf(faint, "  release:     %s (%s)\n", releaseShort, latest)
		fmt.Println()
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" this is a nix installation")
		errutil.Println(faint, "  nix store is immutable; use nix commands to upgrade")
		fmt.Println()
		printNixUpgradeInstructions()
		return nil
	}

	if comparison.AheadBy > 0 {
		cyan := color.New(color.FgCyan)
		errutil.Printf(cyan, tui.SymbolArrowUp+" you are %d commit(s) ahead of %s\n", comparison.AheadBy, latest)
		errutil.Printf(faint, "  your commit: %s\n", currentCommit)
		errutil.Printf(faint, "  release:     %s (%s)\n", releaseShort, latest)
		fmt.Println()
		errutil.Println(faint, "you are running a newer build than the latest release")
		return nil
	}

	if comparison.BehindBy > 0 {
		green := color.New(color.FgGreen, color.Bold)
		errutil.Printf(green, tui.SymbolSuccess+" update available: %d commit(s) behind %s\n", comparison.BehindBy, latest)
		errutil.Printf(faint, "  your commit: %s\n", currentCommit)
		errutil.Printf(faint, "  release:     %s (%s)\n", releaseShort, latest)
		fmt.Println()
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" this is a nix installation")
		errutil.Println(faint, "  nix store is immutable; use nix commands to upgrade")
		fmt.Println()
		printNixUpgradeInstructions()
		return nil
	}

	green := color.New(color.FgGreen)
	errutil.Printf(green, tui.SymbolSuccess+" you are running %s (commit %s)\n", latest, releaseShort)
	return nil
}

func handleNixSpecificVersion(current, target string) error {
	if !strings.HasPrefix(target, "v") {
		target = "v" + target
	}

	printVersionComparisonTarget(current, target)

	yellow := color.New(color.FgYellow)
	errutil.Println(yellow, tui.SymbolWarning+" this is a nix installation")
	faint := color.New(color.Faint)
	errutil.Println(faint, "  nix store is immutable; in-place upgrades are not supported")
	fmt.Println()

	bold := color.New(color.Bold)
	cmd := color.New(color.FgCyan)

	errutil.Println(bold, "to install a specific version with nix:")
	fmt.Println()

	errutil.Print(faint, "  specific ref    ")
	errutil.Printf(cmd, "nix profile install github:%s/%s/%s\n", repoOwner, repoName, target)

	errutil.Print(faint, "  latest          ")
	errutil.Printf(cmd, "nix profile install github:%s/%s\n", repoOwner, repoName)

	return nil
}

func isVersionLower(v1, v2 string) bool {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	for i := 0; i < 3; i++ {
		if parts1[i] < parts2[i] {
			return true
		}
		if parts1[i] > parts2[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	var parts [3]int
	segments := strings.Split(v, ".")

	for i := 0; i < len(segments) && i < 3; i++ {
		n, err := strconv.Atoi(segments[i])
		if err == nil {
			parts[i] = n
		}
	}
	return parts
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPI, repoOwner, repoName)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer errutil.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}

	return release.TagName, nil
}

func printVersionComparison(current, latest string) {
	faint := color.New(color.Faint)
	version := color.New(color.FgCyan)

	errutil.Print(faint, "current  ")
	errutil.Println(version, current)
	errutil.Print(faint, "latest   ")
	errutil.Println(version, latest)
	fmt.Println()
}

func printVersionComparisonTarget(current, target string) {
	faint := color.New(color.Faint)
	version := color.New(color.FgCyan)

	errutil.Print(faint, "current  ")
	errutil.Println(version, current)
	errutil.Print(faint, "target   ")
	errutil.Println(version, target)
	fmt.Println()
}

func printUpgradeInstructions() {
	bold := color.New(color.Bold)
	faint := color.New(color.Faint)
	cmd := color.New(color.FgCyan)

	errutil.Println(bold, "upgrade options:")
	fmt.Println()

	errutil.Print(faint, "  go install    ")
	errutil.Printf(cmd, "go install github.com/%s/%s@latest\n", repoOwner, repoName)

	errutil.Print(faint, "  shell script  ")
	errutil.Printf(cmd, "curl -sSL https://raw.githubusercontent.com/%s/%s/master/install.sh | sh\n", repoOwner, repoName)

	errutil.Print(faint, "  arch (aur)    ")
	errutil.Println(cmd, "yay -S snitch-bin")

	errutil.Print(faint, "  nix           ")
	errutil.Printf(cmd, "nix profile upgrade --inputs-from github:%s/%s\n", repoOwner, repoName)
}

func performUpgrade(version string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	if strings.HasPrefix(execPath, "/nix/store/") {
		yellow := color.New(color.FgYellow)
		errutil.Println(yellow, tui.SymbolWarning+" cannot perform in-place upgrade for nix installation")
		fmt.Println()
		printNixUpgradeInstructions()
		return nil
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	versionClean := strings.TrimPrefix(version, "v")
	archiveName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", repoName, versionClean, goos, goarch)
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		repoOwner, repoName, version, archiveName)

	faint := color.New(color.Faint)
	cyan := color.New(color.FgCyan)
	errutil.Print(faint, tui.SymbolDownload+" downloading ")
	errutil.Printf(cyan, "%s", archiveName)
	errutil.Println(faint, "...")

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer errutil.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "snitch-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer errutil.RemoveAll(tmpDir)

	binaryPath, err := extractBinaryFromTarGz(resp.Body, tmpDir)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	if goos == "darwin" {
		removeQuarantine(binaryPath)
	}

	// check if we can write to the target location
	targetDir := filepath.Dir(execPath)
	if !isWritable(targetDir) {
		yellow := color.New(color.FgYellow)
		cmdStyle := color.New(color.FgCyan)

		errutil.Printf(yellow, tui.SymbolWarning+" elevated permissions required to install to %s\n", targetDir)
		fmt.Println()
		errutil.Println(faint, "run with sudo or install to a user-writable location:")
		fmt.Println()
		errutil.Print(faint, "  sudo         ")
		errutil.Println(cmdStyle, "sudo snitch upgrade --yes")
		errutil.Print(faint, "  custom dir   ")
		errutil.Printf(cmdStyle, "curl -sSL https://raw.githubusercontent.com/%s/%s/master/install.sh | INSTALL_DIR=~/.local/bin sh\n",
			repoOwner, repoName)
		return nil
	}

	// replace the binary
	backupPath := execPath + ".bak"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := copyFile(binaryPath, execPath); err != nil {
		// try to restore backup
		if restoreErr := os.Rename(backupPath, execPath); restoreErr != nil {
			return fmt.Errorf("failed to install new binary and restore backup: %w (restore error: %v)", err, restoreErr)
		}
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	if err := os.Chmod(execPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	if err := os.Remove(backupPath); err != nil {
		// non-fatal, just warn
		yellow := color.New(color.FgYellow)
		errutil.Fprintf(yellow, os.Stderr, tui.SymbolWarning+" warning: failed to remove backup file %s: %v\n", backupPath, err)
	}

	green := color.New(color.FgGreen, color.Bold)
	errutil.Printf(green, tui.SymbolSuccess+" successfully upgraded to %s\n", version)
	return nil
}

func extractBinaryFromTarGz(r io.Reader, destDir string) (string, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer errutil.Close(gzr)

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		// look for the snitch binary
		name := filepath.Base(header.Name)
		if name != repoName {
			continue
		}

		destPath := filepath.Join(destDir, name)
		outFile, err := os.Create(destPath)
		if err != nil {
			return "", err
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			errutil.Close(outFile)
			return "", err
		}
		errutil.Close(outFile)

		return destPath, nil
	}

	return "", fmt.Errorf("binary not found in archive")
}

func isWritable(path string) bool {
	testFile := filepath.Join(path, ".snitch-write-test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	errutil.Close(f)
	errutil.Remove(testFile)
	return true
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer errutil.Close(srcFile)

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer errutil.Close(dstFile)

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}

func removeQuarantine(path string) {
	cmd := exec.Command("xattr", "-d", "com.apple.quarantine", path)
	if err := cmd.Run(); err == nil {
		faint := color.New(color.Faint)
		errutil.Println(faint, "  removed macOS quarantine attribute")
	}
}

func isNixInstall() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}

	resolved, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(resolved, "/nix/store/")
}

var nixVersionPattern = regexp.MustCompile(`^nix-([a-f0-9]+)(-dirty)?$`)
var commitHashPattern = regexp.MustCompile(`^[a-f0-9]{7,40}$`)

func isNixVersion(version string) bool {
	if nixVersionPattern.MatchString(version) {
		return true
	}
	if commitHashPattern.MatchString(version) {
		return true
	}
	return false
}

func extractCommitFromVersion(version string) string {
	matches := nixVersionPattern.FindStringSubmatch(version)
	if len(matches) >= 2 {
		return matches[1]
	}
	if commitHashPattern.MatchString(version) {
		return version
	}
	return ""
}

func isNixDirty(version string) bool {
	return strings.HasSuffix(version, "-dirty")
}

func fetchCommitForTag(tag string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s", githubAPI, repoOwner, repoName, tag)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer errutil.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var commit githubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", err
	}

	return commit.SHA, nil
}

func compareCommits(base, head string) (*githubCompare, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/compare/%s...%s", githubAPI, repoOwner, repoName, base, head)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer errutil.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var compare githubCompare
	if err := json.NewDecoder(resp.Body).Decode(&compare); err != nil {
		return nil, err
	}

	return &compare, nil
}

func printNixUpgradeInstructions() {
	bold := color.New(color.Bold)
	faint := color.New(color.Faint)
	cmd := color.New(color.FgCyan)

	errutil.Println(bold, "nix upgrade options:")
	fmt.Println()

	errutil.Print(faint, "  flake profile   ")
	errutil.Printf(cmd, "nix profile install github:%s/%s\n", repoOwner, repoName)

	errutil.Print(faint, "  flake update    ")
	errutil.Println(cmd, "nix flake update snitch  (in your system/home-manager config)")

	errutil.Print(faint, "  rebuild         ")
	errutil.Println(cmd, "nixos-rebuild switch  or  home-manager switch")
}

