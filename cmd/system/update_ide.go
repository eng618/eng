package system

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// UpdateIdeCmd represents the command to update Antigravity IDE.
var UpdateIdeCmd = &cobra.Command{
	Use:     "ide [archive-path-or-url]",
	Aliases: []string{"agy-ide", "antigravity-ide", "antigravity"},
	Short:   "Update or install Antigravity IDE",
	Long: `Download, validate, and install the latest Antigravity IDE release package.
Supports automated download, detecting downloaded archives in ~/Downloads, or specifying an archive directly.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		urlPath, _ := cmd.Flags().GetString("url")
		autoApprove, _ := cmd.Flags().GetBool("yes")

		target := ""
		if len(args) > 0 {
			target = args[0]
		} else if filePath != "" {
			target = filePath
		} else if urlPath != "" {
			target = urlPath
		}

		return RunIdeUpdate(cmd.Context(), target, cmdutil.IsVerbose(cmd), autoApprove)
	},
}

func init() {
	UpdateIdeCmd.Flags().StringP("file", "f", "", "Path to local Antigravity IDE tarball (.tar.gz)")
	UpdateIdeCmd.Flags().StringP("url", "u", "", "Direct download URL for the Antigravity IDE tarball")
	UpdateIdeCmd.Flags().BoolP("yes", "y", false, "Auto-approve update operations without prompting")
}

// RunIdeUpdate orchestrates finding, downloading, extracting, validating, and installing the IDE.
func RunIdeUpdate(ctx context.Context, target string, verbose, autoApprove bool) error {
	if runtime.GOOS != "linux" {
		log.Warn("Antigravity IDE automated tarball installation is currently configured for Linux.")
		return nil
	}

	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	archivePath := target

	// 1. If target is a URL, download it
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		downloaded, err := downloadArchive(ctx, target, homeDir, verbose)
		if err != nil {
			return fmt.Errorf("failed to download archive: %w", err)
		}
		archivePath = downloaded
		defer os.Remove(downloaded)
	} else if archivePath == "" {
		// 2. Check if a configured URL exists
		configURL := strings.TrimSpace(viper.GetString("antigravity.ide_download_url"))
		if configURL != "" {
			log.Info("Using configured download URL: %s", configURL)
			downloaded, err := downloadArchive(ctx, configURL, homeDir, verbose)
			if err != nil {
				log.Warn("Download from configured URL failed: %v", err)
			} else {
				archivePath = downloaded
				defer os.Remove(downloaded)
			}
		}
	}

	// 3. Fallback: Search ~/Downloads for recent IDE archives
	if archivePath == "" {
		downloadsDir := filepath.Join(homeDir, "Downloads")
		detected := findLatestIdeArchive(downloadsDir)
		if detected != "" {
			log.Message("Found Antigravity IDE archive in Downloads: %s", detected)
			archivePath = detected
		}
	}

	// 4. Fallback: Prompt user to open download page if still not found
	if archivePath == "" {
		log.Warn("No Antigravity IDE archive found in ~/Downloads and no download URL specified.")
		if !autoApprove {
			confirm, err := ui.Confirm("Would you like to open the official Antigravity download page in your browser?", true)
			if err == nil && confirm {
				_ = openURL("https://antigravity.google/download")
				log.Message("Please download the Linux Antigravity IDE package to ~/Downloads.")
				log.Message("After downloading, re-run 'eng system update ide' to complete the update.")
				return nil
			}
		}
		return errors.New("no Antigravity IDE archive available to install")
	}

	return installIdeArchive(archivePath, homeDir, verbose)
}

func findLatestIdeArchive(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	type fileWithTime struct {
		path    string
		modTime time.Time
	}

	var ideFiles []fileWithTime
	var otherAntigravityFiles []fileWithTime

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".tgz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		if strings.Contains(name, "antigravity") && strings.Contains(name, "ide") {
			ideFiles = append(ideFiles, fileWithTime{path: fullPath, modTime: info.ModTime()})
		} else if strings.HasPrefix(name, "antigravity") {
			otherAntigravityFiles = append(otherAntigravityFiles, fileWithTime{path: fullPath, modTime: info.ModTime()})
		}
	}

	if len(ideFiles) > 0 {
		sort.Slice(ideFiles, func(i, j int) bool {
			return ideFiles[i].modTime.After(ideFiles[j].modTime)
		})
		return ideFiles[0].path
	}

	// If only other antigravity archives exist, return the latest for structural inspection
	if len(otherAntigravityFiles) > 0 {
		sort.Slice(otherAntigravityFiles, func(i, j int) bool {
			return otherAntigravityFiles[i].modTime.After(otherAntigravityFiles[j].modTime)
		})
		return otherAntigravityFiles[0].path
	}

	return ""
}

type progressWriter struct {
	total      int64
	downloaded int64
	spinner    *ui.Spinner
	lastUpdate time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.downloaded += int64(n)

	if time.Since(pw.lastUpdate) > 200*time.Millisecond {
		pw.lastUpdate = time.Now()
		if pw.spinner != nil {
			if pw.total > 0 {
				percent := float64(pw.downloaded) / float64(pw.total)
				msg := fmt.Sprintf("Downloading Antigravity IDE... %s / %s (%.0f%%)",
					humanize.Bytes(uint64(pw.downloaded)),
					humanize.Bytes(uint64(pw.total)),
					percent*100.0)
				pw.spinner.SetProgressBar(percent, msg)
			} else {
				pw.spinner.UpdateMessage(fmt.Sprintf("Downloading Antigravity IDE... %s",
					humanize.Bytes(uint64(pw.downloaded))))
			}
		}
	}
	return n, nil
}

func downloadArchive(ctx context.Context, url, homeDir string, verbose bool) (string, error) {
	log.Start("Downloading Antigravity IDE from: %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "eng-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	tmpFile, err := os.CreateTemp("", "antigravity-ide-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	var spinner *ui.Spinner
	if !ui.DisableProgress {
		if resp.ContentLength > 0 {
			spinner = ui.NewProgressSpinner("Downloading Antigravity IDE...")
		} else {
			spinner = ui.NewSpinner("Downloading Antigravity IDE...")
		}
		spinner.Start()
	}

	pw := &progressWriter{
		total:      resp.ContentLength,
		spinner:    spinner,
		lastUpdate: time.Now(),
	}

	_, err = io.Copy(tmpFile, io.TeeReader(resp.Body, pw))
	if spinner != nil {
		spinner.Stop()
	}
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to save download: %w", err)
	}

	log.Success("Download completed: %s", humanize.Bytes(uint64(pw.downloaded)))
	return tmpFile.Name(), nil
}

func installIdeArchive(archivePath, homeDir string, verbose bool) error {
	log.Start("Extracting archive for inspection: %s", filepath.Base(archivePath))

	tmpExtractDir, err := os.MkdirTemp("", "antigravity-ide-extract-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	defer os.RemoveAll(tmpExtractDir)

	if err := extractTarGz(archivePath, tmpExtractDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Locate root folder if packed with a single root directory
	extractedRoot := tmpExtractDir
	entries, err := os.ReadDir(tmpExtractDir)
	if err == nil && len(entries) == 1 && entries[0].IsDir() {
		extractedRoot = filepath.Join(tmpExtractDir, entries[0].Name())
	}

	// Validate: Ensure this is Antigravity IDE (VS Code based), NOT Antigravity Desktop App
	isDesktopHub := false
	if _, err := os.Stat(filepath.Join(extractedRoot, "resources", "app.asar")); err == nil {
		isDesktopHub = true
	}

	hasIdeCLI := false
	if _, err := os.Stat(filepath.Join(extractedRoot, "resources", "app", "out", "cli.js")); err == nil {
		hasIdeCLI = true
	}
	if _, err := os.Stat(filepath.Join(extractedRoot, "bin", "antigravity-ide")); err == nil {
		hasIdeCLI = true
	}
	if _, err := os.Stat(filepath.Join(extractedRoot, "antigravity-ide")); err == nil {
		hasIdeCLI = true
	}

	if isDesktopHub && !hasIdeCLI {
		return fmt.Errorf("the archive '%s' appears to be the Antigravity Desktop App (2.0), not the Antigravity IDE", filepath.Base(archivePath))
	}

	if !hasIdeCLI {
		return fmt.Errorf("the archive '%s' does not appear to contain a valid Antigravity IDE installation", filepath.Base(archivePath))
	}

	log.Success("Archive payload verified as Antigravity IDE.")

	installDir := filepath.Join(homeDir, ".local", "opt", "antigravity-ide")
	backupDir := installDir + ".bak"

	// Backup existing installation
	if _, err := os.Stat(installDir); err == nil {
		log.Verbose(verbose, "Backing up current installation to %s", backupDir)
		_ = os.RemoveAll(backupDir)
		if err := copyDir(installDir, backupDir); err != nil {
			log.Warn("Failed to backup previous installation: %v", err)
		}
	}

	log.Start("Deploying Antigravity IDE to %s...", installDir)
	_ = os.RemoveAll(installDir)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	if err := copyDir(extractedRoot, installDir); err != nil {
		return fmt.Errorf("failed to deploy files to %s: %w", installDir, err)
	}

	// Ensure permissions
	_ = os.Chmod(filepath.Join(installDir, "antigravity-ide"), 0o755)
	_ = os.Chmod(filepath.Join(installDir, "bin", "antigravity-ide"), 0o755)

	// Ensure ~/.local/bin symlinks
	binDir := filepath.Join(homeDir, ".local", "bin")
	_ = os.MkdirAll(binDir, 0o755)

	targetBin := filepath.Join(installDir, "bin", "antigravity-ide")
	symlinks := []string{"antigravity-ide", "agy-ide", "antigravity"}
	for _, link := range symlinks {
		linkPath := filepath.Join(binDir, link)
		_ = os.Remove(linkPath)
		_ = os.Symlink(targetBin, linkPath)
	}

	// Ensure desktop entry exists
	appsDir := filepath.Join(homeDir, ".local", "share", "applications")
	_ = os.MkdirAll(appsDir, 0o755)
	desktopFile := filepath.Join(appsDir, "antigravity-ide.desktop")
	if _, err := os.Stat(desktopFile); os.IsNotExist(err) {
		desktopContent := fmt.Sprintf(`[Desktop Entry]
Name=Antigravity IDE
Comment=AI-First Integrated Development Environment
Exec=%s --no-sandbox %%F
Icon=%s/.local/share/icons/antigravity-ide.png
Type=Application
StartupNotify=true
StartupWMClass=Antigravity IDE
Categories=Development;IDE;TextEditor;Utility;
MimeType=text/plain;inode/directory;x-scheme-handler/antigravity;
Actions=new-empty-window;

[Desktop Action new-empty-window]
Name=New Empty Window
Exec=%s --no-sandbox --new-window %%F
Icon=%s/.local/share/icons/antigravity-ide.png
`, targetBin, homeDir, targetBin, homeDir)
		_ = os.WriteFile(desktopFile, []byte(desktopContent), 0o644)
	}

	_ = execCommand("update-desktop-database", appsDir).Run()
	_ = execCommand("xdg-mime", "default", "antigravity-ide.desktop", "x-scheme-handler/antigravity").Run()

	// Verify installation
	verCmd := execCommand(targetBin, "--version")
	out, err := verCmd.Output()
	versionStr := "installed"
	if err == nil && len(out) > 0 {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 {
			versionStr = lines[0]
		}
	}

	theme.SuccessMessage(fmt.Sprintf("Antigravity IDE successfully updated! (Version: %s)", versionStr))
	return nil
}

func extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Security check: Zip Slip vulnerability prevention
		cleanPath := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
			continue
		}

		target := filepath.Join(destDir, cleanPath)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			_ = os.Remove(target)
			_ = os.Symlink(header.Linkname, target)
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Handle symlink
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			_ = os.Remove(targetPath)
			return os.Symlink(linkTarget, targetPath)
		}

		// Regular file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}
