package cleanup

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

var execCommand = exec.Command
var lookPath = exec.LookPath

// IsDockerAvailable checks if the docker CLI is present and the daemon responds.
func IsDockerAvailable() bool {
	if _, err := lookPath("docker"); err != nil {
		return false
	}
	cmd := execCommand("docker", "info", "--format", "{{.ServerVersion}}")
	return cmd.Run() == nil
}

// GetDockerDiskUsage returns the raw output of docker system df.
func GetDockerDiskUsage() (string, error) {
	cmd := execCommand("docker", "system", "df")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker system df failed: %w", err)
	}
	return string(out), nil
}

// RunDockerCleanup performs granular Docker engine pruning.
func RunDockerCleanup(opts DockerCleanOptions) (*Report, error) {
	report := &Report{}

	if !IsDockerAvailable() {
		report.Add(ItemResult{
			Name:     "Docker Daemon & Images",
			Category: "docker",
			Status:   StatusSkipped,
			Message:  "Docker is not installed or the Docker daemon is not running.",
		})
		return report, nil
	}

	if opts.DryRun {
		dfOut, err := GetDockerDiskUsage()
		if err != nil {
			report.Add(ItemResult{
				Name:     "Docker System Analysis (Dry-Run)",
				Category: "docker",
				Status:   StatusFailed,
				Message:  err.Error(),
			})
			return report, nil
		}

		log.Message("\n--- Docker Disk Usage (Dry-Run Preview) ---")
		fmt.Println(dfOut)

		report.Add(ItemResult{
			Name:        "Docker Disk Analysis (Dry-Run)",
			Category:    "docker",
			Status:      StatusSuccess,
			ReclaimText: "Simulated",
		})
		return report, nil
	}

	var multiSpinner *ui.MultiSpinner
	if !ui.DisableProgress {
		multiSpinner, _ = ui.NewMultiSpinner()
		defer func() {
			if multiSpinner != nil {
				multiSpinner.Stop()
			}
		}()
	}

	// 1. Prune Dangling Image Layers
	var spDangling ui.ProgressSpinner
	if multiSpinner != nil {
		spDangling = multiSpinner.AddSpinner("Pruning dangling Docker images...")
	}

	danglingBytes, danglingOut, err := runPruneCmd("docker", "image", "prune", "-f")
	if err != nil {
		if spDangling != nil {
			spDangling.Fail(fmt.Sprintf("Failed to prune dangling images: %v", err))
		}
		report.Add(ItemResult{
			Name:     "Docker Dangling Layers",
			Category: "docker",
			Status:   StatusFailed,
			Message:  err.Error(),
		})
	} else {
		freedStr := humanize.Bytes(uint64(danglingBytes))
		if spDangling != nil {
			spDangling.Success(fmt.Sprintf("Pruned dangling images (freed %s)", freedStr))
		}
		log.Verbose(opts.Verbose, "Dangling prune output:\n%s", danglingOut)
		report.Add(ItemResult{
			Name:       "Docker Dangling Layers",
			Category:   "docker",
			Status:     StatusSuccess,
			FreedBytes: danglingBytes,
		})
	}

	// 2. Prune Build Cache
	if opts.BuildCache {
		var spBuild ui.ProgressSpinner
		if multiSpinner != nil {
			spBuild = multiSpinner.AddSpinner("Pruning Docker build cache...")
		}

		buildBytes, buildOut, err := runPruneCmd("docker", "builder", "prune", "-f")
		if err != nil {
			if spBuild != nil {
				spBuild.Fail(fmt.Sprintf("Failed to prune build cache: %v", err))
			}
			report.Add(ItemResult{
				Name:     "Docker Build Cache",
				Category: "docker",
				Status:   StatusFailed,
				Message:  err.Error(),
			})
		} else {
			freedStr := humanize.Bytes(uint64(buildBytes))
			if spBuild != nil {
				spBuild.Success(fmt.Sprintf("Pruned build cache (freed %s)", freedStr))
			}
			log.Verbose(opts.Verbose, "Build cache prune output:\n%s", buildOut)
			report.Add(ItemResult{
				Name:       "Docker Build Cache",
				Category:   "docker",
				Status:     StatusSuccess,
				FreedBytes: buildBytes,
			})
		}
	}

	// 3. Prune Unused Images (Filtered by Age or All)
	duration := opts.OlderThan
	if duration == "" {
		duration = "168h"
	}

	var spUnused ui.ProgressSpinner
	unusedLabel := fmt.Sprintf("Pruning unused Docker images older than %s...", duration)
	if opts.All {
		unusedLabel = "Pruning all unused Docker images..."
	}
	if multiSpinner != nil {
		spUnused = multiSpinner.AddSpinner(unusedLabel)
	}

	pruneArgs := []string{"image", "prune", "-a", "-f"}
	if !opts.All {
		pruneArgs = append(pruneArgs, "--filter", fmt.Sprintf("until=%s", duration))
	}

	unusedBytes, unusedOut, err := runPruneCmd("docker", pruneArgs...)
	if err != nil {
		if spUnused != nil {
			spUnused.Fail(fmt.Sprintf("Failed to prune unused images: %v", err))
		}
		report.Add(ItemResult{
			Name:     fmt.Sprintf("Docker Unused Images (filter: %s)", duration),
			Category: "docker",
			Status:   StatusFailed,
			Message:  err.Error(),
		})
	} else {
		freedStr := humanize.Bytes(uint64(unusedBytes))
		if spUnused != nil {
			spUnused.Success(fmt.Sprintf("Pruned unused images (freed %s)", freedStr))
		}
		log.Verbose(opts.Verbose, "Unused image prune output:\n%s", unusedOut)
		report.Add(ItemResult{
			Name:       fmt.Sprintf("Docker Unused Images (filter: %s)", duration),
			Category:   "docker",
			Status:     StatusSuccess,
			FreedBytes: unusedBytes,
		})
	}

	// 4. Prune Dangling Volumes if requested
	if opts.Volumes {
		var spVol ui.ProgressSpinner
		if multiSpinner != nil {
			spVol = multiSpinner.AddSpinner("Pruning unused Docker volumes...")
		}

		volBytes, volOut, err := runPruneCmd("docker", "volume", "prune", "-f")
		if err != nil {
			if spVol != nil {
				spVol.Fail(fmt.Sprintf("Failed to prune volumes: %v", err))
			}
			report.Add(ItemResult{
				Name:     "Docker Unused Volumes",
				Category: "docker",
				Status:   StatusFailed,
				Message:  err.Error(),
			})
		} else {
			freedStr := humanize.Bytes(uint64(volBytes))
			if spVol != nil {
				spVol.Success(fmt.Sprintf("Pruned unused volumes (freed %s)", freedStr))
			}
			log.Verbose(opts.Verbose, "Volume prune output:\n%s", volOut)
			report.Add(ItemResult{
				Name:       "Docker Unused Volumes",
				Category:   "docker",
				Status:     StatusSuccess,
				FreedBytes: volBytes,
			})
		}
	}

	return report, nil
}

func runPruneCmd(cmdName string, args ...string) (int64, string, error) {
	cmd := execCommand(cmdName, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	output := outBuf.String()
	if err != nil {
		return 0, errBuf.String(), fmt.Errorf("%s: %s", err, errBuf.String())
	}

	freedBytes := ParseReclaimedBytes(output)
	return freedBytes, output, nil
}

var (
	reclaimedRe = regexp.MustCompile(`(?i)(?:Total reclaimed space|Total):\s*([\d\.]+)\s*([KMGT]?B)`)
)

// ParseReclaimedBytes extracts the byte count from docker prune output.
func ParseReclaimedBytes(output string) int64 {
	matches := reclaimedRe.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return 0
	}

	var total int64
	for _, m := range matches {
		if len(m) >= 3 {
			val, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				continue
			}
			unit := strings.ToUpper(m[2])
			var multiplier float64
			switch unit {
			case "B":
				multiplier = 1
			case "KB", "KIB":
				multiplier = 1000
			case "MB", "MIB":
				multiplier = 1000 * 1000
			case "GB", "GIB":
				multiplier = 1000 * 1000 * 1000
			case "TB", "TIB":
				multiplier = 1000 * 1000 * 1000 * 1000
			default:
				multiplier = 1
			}
			total += int64(val * multiplier)
		}
	}
	return total
}
