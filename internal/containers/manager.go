package containers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var execCommand = exec.Command

// Stack represents a Docker Compose stack definition.
type Stack struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	File       string   `json:"file"`
	Services   []string `json:"services"`
	Status     string   `json:"status"`
	Containers int      `json:"containers"`
}

// Manager handles Docker Compose stack operations.
type Manager struct {
	BasePath string
}

// NewManager creates a new containers Manager targeting the specified base path.
func NewManager(basePath string) *Manager {
	if basePath == "" {
		home, _ := os.UserHomeDir()
		basePath = filepath.Join(home, "bin", "containers")
	}
	return &Manager{BasePath: basePath}
}

// DiscoverStacks scans the containers path for valid Compose files.
func (m *Manager) DiscoverStacks() ([]Stack, error) {
	var stacks []Stack

	stacksDir := filepath.Join(m.BasePath, "stacks")
	entries, err := os.ReadDir(stacksDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				stackName := entry.Name()
				composeFile := filepath.Join(stacksDir, stackName, "docker-compose.yml")
				if _, err := os.Stat(composeFile); os.IsNotExist(err) {
					composeFile = filepath.Join(stacksDir, stackName, "docker-compose.yaml")
				}
				if _, err := os.Stat(composeFile); err == nil {
					services, _ := parseServices(composeFile)
					stacks = append(stacks, Stack{
						Name:     stackName,
						Path:     filepath.Dir(composeFile),
						File:     composeFile,
						Services: services,
						Status:   "Unknown",
					})
				}
			}
		}
	}

	// Fallback to top-level compose if no stacks folder present
	if len(stacks) == 0 {
		topCompose := filepath.Join(m.BasePath, "docker-compose.yml")
		if _, err := os.Stat(topCompose); err == nil {
			services, _ := parseServices(topCompose)
			stacks = append(stacks, Stack{
				Name:     "default",
				Path:     m.BasePath,
				File:     topCompose,
				Services: services,
				Status:   "Unknown",
			})
		}
	}

	return stacks, nil
}

// EnsureSharedNetwork checks if eng-shared-net exists, creating it if necessary.
func (m *Manager) EnsureSharedNetwork() error {
	cmd := execCommand("docker", "network", "inspect", "eng-shared-net")
	if err := cmd.Run(); err != nil {
		createCmd := execCommand("docker", "network", "create", "eng-shared-net")
		if out, createErr := createCmd.CombinedOutput(); createErr != nil {
			return fmt.Errorf("failed to create eng-shared-net: %s", string(out))
		}
	}
	return nil
}

// Up starts target stack(s) using docker compose up.
func (m *Manager) Up(stackNames []string, envName string, detach bool, build bool) error {
	if err := m.EnsureSharedNetwork(); err != nil {
		// Log warning or continue
	}

	targetStacks, err := m.resolveStacks(stackNames)
	if err != nil {
		return err
	}

	for _, s := range targetStacks {
		args := []string{"compose", "-f", s.File}

		envFile := filepath.Join(s.Path, fmt.Sprintf(".env.%s", envName))
		if _, err := os.Stat(envFile); err == nil {
			args = append(args, "--env-file", envFile)
		} else {
			defaultEnv := filepath.Join(s.Path, ".env")
			if _, err := os.Stat(defaultEnv); err == nil {
				args = append(args, "--env-file", defaultEnv)
			}
		}

		args = append(args, "up")
		if detach {
			args = append(args, "-d")
		}
		if build {
			args = append(args, "--build")
		}

		cmd := execCommand("docker", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error starting stack %s: %w", s.Name, err)
		}
	}
	return nil
}

// Down stops target stack(s) using docker compose down.
func (m *Manager) Down(stackNames []string, removeVolumes bool) error {
	targetStacks, err := m.resolveStacks(stackNames)
	if err != nil {
		return err
	}

	for _, s := range targetStacks {
		args := []string{"compose", "-f", s.File, "down"}
		if removeVolumes {
			args = append(args, "-v")
		}

		cmd := execCommand("docker", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error stopping stack %s: %w", s.Name, err)
		}
	}
	return nil
}

// Pull pulls latest images for target stack(s).
func (m *Manager) Pull(stackNames []string) error {
	targetStacks, err := m.resolveStacks(stackNames)
	if err != nil {
		return err
	}

	for _, s := range targetStacks {
		cmd := execCommand("docker", "compose", "-f", s.File, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error pulling stack %s: %w", s.Name, err)
		}
	}
	return nil
}

// Status inspects the status of target stack(s).
func (m *Manager) Status(stackNames []string) ([]Stack, error) {
	targetStacks, err := m.resolveStacks(stackNames)
	if err != nil {
		return nil, err
	}

	var results []Stack
	for _, s := range targetStacks {
		cmd := execCommand("docker", "compose", "-f", s.File, "ps", "--format", "json")
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			s.Status = "Stopped"
			s.Containers = 0
		} else {
			s.Containers, s.Status = parseDockerPsJSON(outBuf.Bytes())
		}
		results = append(results, s)
	}

	return results, nil
}

// Logs streams logs for a single stack.
func (m *Manager) Logs(stackName string, follow bool, tail string) error {
	targetStacks, err := m.resolveStacks([]string{stackName})
	if err != nil || len(targetStacks) == 0 {
		return fmt.Errorf("stack %q not found", stackName)
	}

	s := targetStacks[0]
	args := []string{"compose", "-f", s.File, "logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}

	cmd := execCommand("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) resolveStacks(names []string) ([]Stack, error) {
	all, err := m.DiscoverStacks()
	if err != nil {
		return nil, err
	}

	if len(names) == 0 || (len(names) == 1 && names[0] == "all") {
		return all, nil
	}

	nameMap := make(map[string]Stack)
	for _, s := range all {
		nameMap[strings.ToLower(s.Name)] = s
	}

	var matched []Stack
	for _, name := range names {
		if s, ok := nameMap[strings.ToLower(name)]; ok {
			matched = append(matched, s)
		} else {
			return nil, fmt.Errorf("unknown stack: %s", name)
		}
	}
	return matched, nil
}

func parseServices(composeFile string) ([]string, error) {
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var services []string
	if svcs, ok := raw["services"].(map[string]interface{}); ok {
		for name := range svcs {
			services = append(services, name)
		}
	}
	return services, nil
}

type psContainer struct {
	State string `json:"State"`
}

func parseDockerPsJSON(data []byte) (int, string) {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return 0, "Stopped"
	}

	running := 0
	total := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var container psContainer
		if err := json.Unmarshal([]byte(line), &container); err == nil {
			total++
			if strings.EqualFold(container.State, "running") {
				running++
			}
		}
	}

	if total == 0 {
		return 0, "Stopped"
	}
	if running == total {
		return total, "Running"
	}
	if running > 0 {
		return total, fmt.Sprintf("Partial (%d/%d)", running, total)
	}
	return total, "Stopped"
}
