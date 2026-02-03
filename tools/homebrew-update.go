package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Checksums map[string]string

type Config struct {
	Version         string
	TagName         string
	Checksums       Checksums
	DarwinAMD64File string
	DarwinARM64File string
	LinuxAMD64File  string
	LinuxARM64File  string
	PatToken        string
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	formula, err := generateFormula(config)
	if err != nil {
		log.Fatal(err)
	}

	// Clone repo
	runCmd("git", "clone", "https://github.com/eng618/homebrew-eng.git")
	if err := os.Chdir("homebrew-eng/Formula"); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	// Write to file
	if err := os.WriteFile("eng.rb", []byte(formula), 0o644); err != nil {
		log.Fatalf("Failed to write eng.rb: %v", err)
	}

	// Git operations
	runCmd("git", "config", "user.email", "eng618@garciaericn.com")
	runCmd("git", "config", "user.name", "Eric N. Garcia")
	runCmd("git", "config", "--global", "credential.helper", "store")
	runCmd("sh", "-c", fmt.Sprintf("echo 'https://oauth2:%s@github.com' > ~/.git-credentials", config.PatToken))
	runCmd("git", "add", "eng.rb")
	runCmd("git", "commit", "-m", fmt.Sprintf("Update eng to %s", config.TagName))
	runCmd("git", "push", "https://github.com/eng618/homebrew-eng.git")
}

func loadConfig() (*Config, error) {
	// Get env vars
	version := os.Getenv("VERSION")
	tagName := os.Getenv("TAG_NAME")
	checksumsJSON := os.Getenv("CHECKSUMS_JSON")
	darwinAMD64File := os.Getenv("DARWIN_AMD64_FILE")
	darwinARM64File := os.Getenv("DARWIN_ARM64_FILE")
	linuxAMD64File := os.Getenv("LINUX_AMD64_FILE")
	linuxARM64File := os.Getenv("LINUX_ARM64_FILE")
	patToken := os.Getenv("PAT_TOKEN")

	requiredVars := map[string]string{
		"VERSION":           version,
		"TAG_NAME":          tagName,
		"CHECKSUMS_JSON":    checksumsJSON,
		"DARWIN_AMD64_FILE": darwinAMD64File,
		"DARWIN_ARM64_FILE": darwinARM64File,
		"LINUX_AMD64_FILE":  linuxAMD64File,
		"LINUX_ARM64_FILE":  linuxARM64File,
		"PAT_TOKEN":         patToken,
	}

	var missingVars []string
	for varName, value := range requiredVars {
		if value == "" {
			missingVars = append(missingVars, varName)
		}
	}
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	var checksums Checksums
	if err := json.Unmarshal([]byte(checksumsJSON), &checksums); err != nil {
		return nil, fmt.Errorf("failed to parse CHECKSUMS_JSON: %v", err)
	}

	return &Config{
		Version:         version,
		TagName:         tagName,
		Checksums:       checksums,
		DarwinAMD64File: darwinAMD64File,
		DarwinARM64File: darwinARM64File,
		LinuxAMD64File:  linuxAMD64File,
		LinuxARM64File:  linuxARM64File,
		PatToken:        patToken,
	}, nil
}

func generateFormula(c *Config) (string, error) {
	getChecksum := func(file string) (string, error) {
		if val, ok := c.Checksums[file]; ok {
			return val, nil
		}
		return "", fmt.Errorf("checksum not found for file: %s", file)
	}

	dAMD64, err := getChecksum(c.DarwinAMD64File)
	if err != nil {
		return "", err
	}
	dARM64, err := getChecksum(c.DarwinARM64File)
	if err != nil {
		return "", err
	}
	lAMD64, err := getChecksum(c.LinuxAMD64File)
	if err != nil {
		return "", err
	}
	lARM64, err := getChecksum(c.LinuxARM64File)
	if err != nil {
		return "", err
	}

	formula := fmt.Sprintf(`class Eng < Formula
  desc 'Personal cli to help facilitate my normal workflow'
  homepage 'https://github.com/eng618/eng'
  version '%[1]s'
  # URLs now use TAG_NAME (with v) for path, and FILE variable (without v) for filename
  case
  when OS.mac? && Hardware::CPU.intel?
    url 'https://github.com/eng618/eng/releases/download/%[2]s/%[3]s'
    sha256 '%[4]s'
  when OS.mac? && Hardware::CPU.arm?
    url 'https://github.com/eng618/eng/releases/download/%[2]s/%[5]s'
    sha256 '%[6]s'
  when OS.linux?
    if Hardware::CPU.intel?
      url 'https://github.com/eng618/eng/releases/download/%[2]s/%[7]s'
      sha256 '%[8]s'
    elsif Hardware::CPU.arm?
      url 'https://github.com/eng618/eng/releases/download/%[2]s/%[9]s'
      sha256 '%[10]s'
    end
  end
  license 'MIT'

  def install
    puts "bin: #{bin}"
    puts "Installing eng to: #{bin}"
    bin.install 'eng'
    puts "eng installed successfully"
    puts "Permissions of eng: #{File.stat("#{bin}/eng").mode.to_s(8)}"
    # Verify the binary is functional before generating completions
    system "#{bin}/eng", '--help'
    generate_completions
  end

  def generate_completions
    puts "PATH: #{ENV['PATH']}"
    puts "Running: #{bin}/eng completion bash"
    (bash_completion/'eng').write Utils.safe_popen_read("#{bin}/eng", 'completion', 'bash')
    (zsh_completion/'_eng').write Utils.safe_popen_read("#{bin}/eng", 'completion', 'zsh')
    (fish_completion/'eng.fish').write Utils.safe_popen_read("#{bin}/eng", 'completion', 'fish')
  end

  test do
    system "#{bin}/eng", '--help'
  end
end
`, c.Version, c.TagName,
		c.DarwinAMD64File, dAMD64,
		c.DarwinARM64File, dARM64,
		c.LinuxAMD64File, lAMD64,
		c.LinuxARM64File, lARM64)

	return formula, nil
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Command failed: %s %s", name, strings.Join(args, " "))
	}
}
