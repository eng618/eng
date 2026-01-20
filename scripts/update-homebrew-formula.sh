#!/bin/bash

# Script to update Homebrew formula for eng
# Usage: ./scripts/update-homebrew-formula.sh
# Requires environment variables:
# - VERSION: version without 'v' prefix (e.g., 0.17.4)
# - TAG_NAME: full tag name with 'v' (e.g., v0.17.4)
# - CHECKSUMS_JSON: JSON object with checksums
# - DARWIN_AMD64_FILE, DARWIN_ARM64_FILE, LINUX_AMD64_FILE, LINUX_ARM64_FILE: filenames
# - PAT_TOKEN: GitHub PAT for pushing

set -e

# Check required env vars
if [[ -z "$VERSION" || -z "$TAG_NAME" || -z "$CHECKSUMS_JSON" || -z "$DARWIN_AMD64_FILE" || -z "$DARWIN_ARM64_FILE" || -z "$LINUX_AMD64_FILE" || -z "$LINUX_ARM64_FILE" || -z "$PAT_TOKEN" ]]; then
  echo "Error: Missing required environment variables"
  exit 1
fi

# Function to get checksum from JSON
get_checksum() {
  local file="$1"
  echo "$CHECKSUMS_JSON" | jq -r ".\"$file\""
}

# Clone the homebrew repo
git clone https://github.com/eng618/homebrew-eng.git
cd homebrew-eng/Formula

# Generate the formula
cat > eng.rb << EOF
class Eng < Formula
  desc 'Personal cli to help facilitate my normal workflow'
  homepage 'https://github.com/eng618/eng'
  version '${VERSION}'
  # URLs now use TAG_NAME (with v) for path, and FILE variable (without v) for filename
  case
  when OS.mac? && Hardware::CPU.intel?
    url 'https://github.com/eng618/eng/releases/download/${TAG_NAME}/${DARWIN_AMD64_FILE}'
    sha256 '$(get_checksum "$DARWIN_AMD64_FILE")'
  when OS.mac? && Hardware::CPU.arm?
    url 'https://github.com/eng618/eng/releases/download/${TAG_NAME}/${DARWIN_ARM64_FILE}'
    sha256 '$(get_checksum "$DARWIN_ARM64_FILE")'
  when OS.linux?
    if Hardware::CPU.intel?
      url 'https://github.com/eng618/eng/releases/download/${TAG_NAME}/${LINUX_AMD64_FILE}'
      sha256 '$(get_checksum "$LINUX_AMD64_FILE")'
    elsif Hardware::CPU.arm?
      url 'https://github.com/eng618/eng/releases/download/${TAG_NAME}/${LINUX_ARM64_FILE}'
      sha256 '$(get_checksum "$LINUX_ARM64_FILE")'
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
EOF

# Git operations
git config user.email "eng618@garciaericn.com"
git config user.name "Eric N. Garcia"
git add eng.rb
git commit -m "Update eng to ${TAG_NAME}"
git push "https://oauth2:${PAT_TOKEN}@github.com/eng618/homebrew-eng.git"