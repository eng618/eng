# -----------------------------------------------------------------------------
# Usage

i: install completion

install: 
	go build && go install

# This only works if you have your completions setup this way.
completion:
	rbk completion zsh >~/.local/share/zsh-completions/_rbk

# -----------------------------------------------------------------------------
# Release helpers

changelog:
	git-chglog -o CHANGELOG.md
	git add --update
	git commit -m "chore(CHANGELOG): update [skip-ci]"

release-check: 
	goreleaser check

release-clean: 
	goreleaser release --clean

release: release-clean changelog

# -----------------------------------------------------------------------------
# development

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

test:
	go test ./...

# -----------------------------------------------------------------------------
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy

tidy:
	go mod tidy

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all

# CIO CI/CD
# -----------------------------------------------------------------------------

detectsecrets: detectsecrets_install detectsecrets_update detectsecrets_audit

detectsecrets_install:
	pip install --upgrade "git+https://github.com/ibm/detect-secrets.git@master#egg+detect-secrets"

detectsecrets_audit:
	detect-secrets audit .secrets.baseline

detectsecrets_update:
	detect-secrets scan --update .secrets.baseline --exclude-files 'go.sum'
