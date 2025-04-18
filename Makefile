# --- Versioning ---
# Get the latest git tag, or commit hash if no tags
VERSION ?= $(shell git describe --tags --always --dirty --abbrev=0)
# Get the short commit hash
COMMIT ?= $(shell git rev-parse --short HEAD)
# Get the build date in ISO 8601 format
DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
# The Go import path for the package where version variables are defined
PACKAGE_PATH=github.com/eng618/eng/cmd
# Construct the ldflags string
LDFLAGS_STRING := -X '$(PACKAGE_PATH).version=$(VERSION)' -X '$(PACKAGE_PATH).commit=$(COMMIT)' -X '$(PACKAGE_PATH).date=$(DATE)'
# Add ldflags to Go build/install commands
GO_BUILD_FLAGS := -ldflags="$(LDFLAGS_STRING)"

# -----------------------------------------------------------------------------
# Usage

i: install completion

# Updated install command to include ldflags
install: build
	@echo "Installing $(PACKAGE_PATH) version $(VERSION)..."
	go install $(GO_BUILD_FLAGS) .

# Added a dedicated build target
build:
	@echo "Building $(PACKAGE_PATH) version $(VERSION)..."
	go build $(GO_BUILD_FLAGS) -o eng .

# This only works if you have your completions setup this way.
completion:
	@echo "Generating Zsh completion..."
	./eng completion zsh > ~/.local/share/zsh-completions/_eng

# -----------------------------------------------------------------------------
# Release helpers

changelog:
	@echo "Generating changelog..."
	git-chglog -o CHANGELOG.md
	git add --update
	git commit -m "chore(CHANGELOG): update [skip-ci]"

release-check:
	@echo "Checking goreleaser configuration..."
	goreleaser check

release-clean:
	@echo "Running goreleaser release..."
	# Goreleaser typically handles its own ldflags via .goreleaser.yaml
	# Ensure your .goreleaser.yaml is configured if you use goreleaser for releases
	goreleaser release --clean

release: release-clean changelog

# -----------------------------------------------------------------------------
# development

lint:
	@echo "Running linter..."
	golangci-lint run

lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

test:
	@echo "Running tests..."
	go test ./...

validate: lint test

# -----------------------------------------------------------------------------
# Modules support

deps-reset:
	@echo "Resetting go.mod..."
	git checkout -- go.mod
	go mod tidy

tidy:
	@echo "Tidying go modules..."
	go mod tidy

deps-list:
	@echo "Listing module updates..."
	go list -m -u -mod=readonly all

deps-upgrade:
	@echo "Upgrading dependencies..."
	go get -u -v ./...
	go mod tidy

deps-cleancache:
	@echo "Cleaning module cache..."
	go clean -modcache

list:
	@echo "Listing all modules..."
	go list -mod=mod all

.PHONY: i install build completion changelog release-check release-clean release lint lint-fix test validate deps-reset tidy deps-list deps-upgrade deps-cleancache list
