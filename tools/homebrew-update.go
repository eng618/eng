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

func main() {
	// Get env vars
	version := os.Getenv("VERSION")
	tagName := os.Getenv("TAG_NAME")
	checksumsJSON := os.Getenv("CHECKSUMS_JSON")
	darwinAMD64File := os.Getenv("DARWIN_AMD64_FILE")
	darwinARM64File := os.Getenv("DARWIN_ARM64_FILE")
	linuxAMD64File := os.Getenv("LINUX_AMD64_FILE")
	linuxARM64File := os.Getenv("LINUX_ARM64_FILE")
	patToken := os.Getenv("PAT_TOKEN")

	if version == "" || tagName == "" || checksumsJSON == "" || darwinAMD64File == "" || darwinARM64File == "" ||
		linuxAMD64File == "" ||
		linuxARM64File == "" ||
		patToken == "" {
		log.Fatal("Missing required environment variables")
	}

	// Parse checksums
	var checksums Checksums
	if err := json.Unmarshal([]byte(checksumsJSON), &checksums); err != nil {
		log.Fatalf("Failed to parse CHECKSUMS_JSON: %v", err)
	}

	// Function to get checksum
	getChecksum := func(file string) string {
		if c, ok := checksums[file]; ok {
			return c
		}
		log.Fatalf("Checksum not found for file: %s", file)
		return ""
	}

	// Clone repo
	runCmd("git", "clone", "https://github.com/eng618/homebrew-eng.git")
	if err := os.Chdir("homebrew-eng/Formula"); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	// Generate formula
	formula := fmt.Sprintf(`class Eng < Formula
  desc 'Personal cli to help facilitate my normal workflow'
  homepage 'https://github.com/eng618/eng'
  version '%s'
  # URLs now use TAG_NAME (with v) for path, and FILE variable (without v) for filename
  case
  when OS.mac? && Hardware::CPU.intel?
    url 'https://github.com/eng618/eng/releases/download/%s/%s'
    sha256 '%s'
  when OS.mac? && Hardware::CPU.arm?
    url 'https://github.com/eng618/eng/releases/download/%s/%s'
    sha256 '%s'
  when OS.linux?
    if Hardware::CPU.intel?
      url 'https://github.com/eng618/eng/releases/download/%s/%s'
      sha256 '%s'
    elsif Hardware::CPU.arm?
      url 'https://github.com/eng618/eng/releases/download/%s/%s'
      sha256 '%s'
    end
  end
  license 'MIT'

  def install
    puts "bin: #{bin}"
    puts "Installing eng to: #{bin}"
    bin.install 'eng'
    puts "eng installed successfully"
    puts "Permissions of eng: #{File.stat(\"#{bin}/eng\").mode.to_s(8)}"
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
`, version,
		tagName, darwinAMD64File, getChecksum(darwinAMD64File),
		tagName, darwinARM64File, getChecksum(darwinARM64File),
		tagName, linuxAMD64File, getChecksum(linuxAMD64File),
		tagName, linuxARM64File, getChecksum(linuxARM64File))

	// Write to file
	if err := os.WriteFile("eng.rb", []byte(formula), 0o644); err != nil {
		log.Fatalf("Failed to write eng.rb: %v", err)
	}

	// Git operations
	runCmd("git", "config", "user.email", "eng618@garciaericn.com")
	runCmd("git", "config", "user.name", "Eric N. Garcia")
	runCmd("git", "add", "eng.rb")
	runCmd("git", "commit", "-m", fmt.Sprintf("Update eng to %s", tagName))
	runCmd("git", "push", fmt.Sprintf("https://oauth2:%s@github.com/eng618/homebrew-eng.git", patToken))
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Command failed: %s %s", name, strings.Join(args, " "))
	}
}
