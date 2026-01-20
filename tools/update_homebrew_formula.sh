#!/bin/bash

set -e

# Environment variables expected:
# VERSION: Version without v prefix (e.g., 0.32.0)
# TAG_NAME: Full tag name with v (e.g., v0.32.0)
# DARWIN_AMD64_FILE, DARWIN_ARM64_FILE, LINUX_AMD64_FILE, LINUX_ARM64_FILE: Archive filenames
# CHECKSUMS_JSON: JSON string with checksums
# PAT_TOKEN: Personal access token for pushing to homebrew-eng

# Change directory into the downloaded dist directory
cd dist

git clone https://github.com/eng618/homebrew-eng.git
cd homebrew-eng/Formula

# Parse checksums
DARWIN_AMD64_SHA=$(echo "$CHECKSUMS_JSON" | jq -r ".\"$DARWIN_AMD64_FILE\"")
DARWIN_ARM64_SHA=$(echo "$CHECKSUMS_JSON" | jq -r ".\"$DARWIN_ARM64_FILE\"")
LINUX_AMD64_SHA=$(echo "$CHECKSUMS_JSON" | jq -r ".\"$LINUX_AMD64_FILE\"")
LINUX_ARM64_SHA=$(echo "$CHECKSUMS_JSON" | jq -r ".\"$LINUX_ARM64_FILE\"")

# Generate Homebrew formula
cat > eng.rb << EOF
class Eng < Formula
  desc "Personal cli to help facilitate my normal workflow"
  homepage "https://github.com/eng618/eng"
  version "${VERSION}"
  # URLs now use TAG_NAME (with v) for path, and FILE variable (without v) for filename
  case
  when OS.mac? && Hardware::CPU.intel?
    url "https://github.com/eng618/eng/releases/download/${TAG_NAME}/${DARWIN_AMD64_FILE}"
    sha256 "${DARWIN_AMD64_SHA}"
  when OS.mac? && Hardware::CPU.arm?
    url "https://github.com/eng618/eng/releases/download/${TAG_NAME}/${DARWIN_ARM64_FILE}"
    sha256 "${DARWIN_ARM64_SHA}"
  when OS.linux?
    if Hardware::CPU.intel?
      url "https://github.com/eng618/eng/releases/download/${TAG_NAME}/${LINUX_AMD64_FILE}"
      sha256 "${LINUX_AMD64_SHA}"
    elsif Hardware::CPU.arm?
      url "https://github.com/eng618/eng/releases/download/${TAG_NAME}/${LINUX_ARM64_FILE}"
      sha256 "${LINUX_ARM64_SHA}"
    end
  end
  license "MIT"

  def install
    puts "bin: #{bin}"
    puts "Installing eng to: #{bin}"
    bin.install "eng"
    puts "eng installed successfully"
    puts "Permissions of eng: #{File.stat("#{bin}/eng").mode.to_s(8)}"
    # Verify the binary is functional before generating completions
    system "#{bin}/eng", "--help"
    generate_completions
  end

  def generate_completions
    puts "PATH: #{ENV['PATH']}"
    puts "Running: #{bin}/eng completion bash"
    (bash_completion/"eng").write Utils.safe_popen_read("#{bin}/eng", "completion", "bash")
    (zsh_completion/"_eng").write Utils.safe_popen_read("#{bin}/eng", "completion", "zsh")
    (fish_completion/"eng.fish").write Utils.safe_popen_read("#{bin}/eng", "completion", "fish")
  end

  test do
    system "#{bin}/eng", "--help"
  end
end
EOF

git config user.email "eng618@garciaericn.com"
git config user.name "Eric N. Garcia"
git add eng.rb
# Commit message still uses the full tag name (with 'v')
git commit -m "Update eng to ${TAG_NAME}"
git push https://oauth2:${PAT_TOKEN}@github.com/eng618/homebrew-eng.git