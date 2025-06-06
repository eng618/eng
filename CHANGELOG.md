<a name="unreleased"></a>

## [Unreleased]

### Refactor

- **codemod:** simplify command handling for lint dependency installation


<a name="v0.21.3"></a>

## [v0.21.3] - 2025-06-06

### Chore

- clean up code structure and improve readability

### Refactor

- **codemod:** enhance lint dependency installation logic


<a name="v0.21.2"></a>

## [v0.21.2] - 2025-06-06

### Chore

- remove outdated Makefile


<a name="v0.21.1"></a>

## [v0.21.1] - 2025-06-06

### Chore

- update go-git and other dependencies to latest versions


<a name="v0.21.0"></a>

## [v0.21.0] - 2025-06-06

### Build

- **deps:** update all packages
- **deps:** bump golangci/golangci-lint-action from 7 to 8 ([#18](https://github.com/eng618/eng/issues/18))

### Chore

- PR reviewers from dependabot -> CODEOWNERS
- update CODEOWNERS
- move PR reviewers to CODEOWNER

### Refactor

- **codemod:** simplify eslint configuration and update lint scripts


<a name="v0.20.5"></a>

## [v0.20.5] - 2025-05-29

### Feat

- **codemod:** update lint dependencies and refine lint-staged config


<a name="v0.20.4"></a>

## [v0.20.4] - 2025-05-23

### Feat

- enhance lint scripts in package.json with caching and add lint report generation


<a name="v0.20.3"></a>

## [v0.20.3] - 2025-05-23

### Feat

- implement lint setup command with improved error handling and structured package.json updates
- enhance package.json writing with ordered fields and improved error handling
- streamline Husky setup by replacing command execution with direct file writing for pre-commit hook


<a name="v0.20.2"></a>

## [v0.20.2] - 2025-05-23

### Feat

- update lint setup command to use latest package versions and change husky installation command


<a name="v0.20.1"></a>

## [v0.20.1] - 2025-05-23

### Feat

- enhance lint setup command with improved error handling and support for legacy peer dependencies

### Refactor

- improve working directory management in TestLintSetupCmd


<a name="v0.20.0"></a>

## [v0.20.0] - 2025-05-23

### Feat

- add codemod command with lint setup functionality and update Copilot instructions

### Test

- refactor proxy configuration saving logic and enhance test coverage for proxy settings


<a name="v0.19.8"></a>

## [v0.19.8] - 2025-05-05

### Feat

- add support for custom no_proxy settings in proxy configuration


<a name="v0.19.7"></a>

## [v0.19.7] - 2025-05-01

### Chore

- update dependencies in go.mod and go.sum to latest versions

### Feat

- enhance PullRebaseBareRepo with autostash and progress options


<a name="v0.19.6"></a>

## [v0.19.6] - 2025-04-30

### Docs

- add command descriptions for config, dotfiles repo, email, and verbose commands

### Feat

- update version command to suggest automatic updates and provide alternative installation instructions
- implement error handling in down and up commands; add utils functions for process management and verbose flag checks

### Refactor

- rename EnsureOnMaster to EnsureOnMain and update related logic


<a name="v0.19.5"></a>

## [v0.19.5] - 2025-04-30

### Feat

- enhance dotfiles commands with info flag and improved error messages


<a name="v0.19.4"></a>

## [v0.19.4] - 2025-04-30

### Feat

- enhance update instructions for Homebrew and manual installation methods


<a name="v0.19.3"></a>

## [v0.19.3] - 2025-04-30

### Feat

- extend Homebrew installation prefixes to include Linuxbrew support


<a name="v0.19.2"></a>

## [v0.19.2] - 2025-04-30

### Feat

- enhance version command with verbose logging for Homebrew update process and GitHub release fetching


<a name="v0.19.1"></a>

## [v0.19.1] - 2025-04-30

### Feat

- add --update flag to version command for Homebrew updates and enhance logging with debug level


<a name="v0.19.0"></a>

## [v0.19.0] - 2025-04-30

### Feat

- improve version command documentation and enhance error handling for GitHub API requests


<a name="v0.18.4"></a>

## [v0.18.4] - 2025-04-30

### Feat

- enhance findNonMovieFolders command with improved logging and spinner handling


<a name="v0.18.3"></a>

## [v0.18.3] - 2025-04-30

### Feat

- update ldflags in goreleaser.yaml for versioning and improve snapshot version template formatting


<a name="v0.18.2"></a>

## [v0.18.2] - 2025-04-29

### Feat

- refactor version command to use logging instead of fmt for output and improve error handling
- enhance findNonMovieFolders command with improved spinner handling and error logging
- improve version command with spinner for update checks and enhanced error handling


<a name="v0.18.1"></a>

## [v0.18.1] - 2025-04-29

### Feat

- enhance version command to check for latest release on GitHub
- add version command with build information and refactor version handling
- enhance killPort command with improved argument validation and error handling

### Refactor

- remove unnecessary time delay after scan completion


<a name="v0.18.0"></a>

## [v0.18.0] - 2025-04-29

### Feat

- enhance findNonMovieFolders command with improved argument handling and progress feedback


<a name="v0.17.22"></a>

## [v0.17.22] - 2025-04-29

### Feat

- enhance spinner functionality with progress tracking for folder scanning


<a name="v0.17.21"></a>

## [v0.17.21] - 2025-04-29

### Feat

- add progress tracking to findNonMovieFolders command with spinner updates
- add spinner utility for improved command feedback during folder scanning


<a name="v0.17.20"></a>

## [v0.17.20] - 2025-04-29

### Refactor

- streamline findNonMovieFolders logic and improve directory handling


<a name="v0.17.19"></a>

## [v0.17.19] - 2025-04-29

### Refactor

- simplify findNonMovieFolders command and improve error handling


<a name="v0.17.18"></a>

## [v0.17.18] - 2025-04-29

### Fix

- streamline folder deletion logic in findNonMovieFolders command


<a name="v0.17.17"></a>

## [v0.17.17] - 2025-04-29

### Refactor

- improve error handling and logging in findNonMovieFolders command


<a name="v0.17.16"></a>

## [v0.17.16] - 2025-04-29

### Refactor

- enhance error logging for directory search in findNonMovieFolders command


<a name="v0.17.15"></a>

## [v0.17.15] - 2025-04-29

### Feat

- add Brew command to update Homebrew packages


<a name="v0.17.14"></a>

## [v0.17.14] - 2025-04-29

### Refactor

- streamline verbose logging in findNonMovieFolders command


<a name="v0.17.13"></a>

## [v0.17.13] - 2025-04-29

### Refactor

- improve logging for folder checks in findNonMovieFolders command


<a name="v0.17.12"></a>

## [v0.17.12] - 2025-04-22

### Chore

- **CHANGELOG:** update [skip-ci]

### Refactor

- **proxy:** enhance proxy command to manage multiple configurations and improve user interaction ([#16](https://github.com/eng618/eng/issues/16))


<a name="v0.17.11"></a>

## [v0.17.11] - 2025-04-22

### Chore

- **CHANGELOG:** update [skip-ci]

### Fix

- correct typo in fetch command description and enhance environment variable handling
- enhance proxy command output with detailed environment variables


<a name="v0.17.10"></a>

## [v0.17.10] - 2025-04-22

### Chore

- **CHANGELOG:** update [skip-ci]

### Fix

- update log message and command syntax for Ubuntu system updates
- replace log.Message with log.Success for successful update notifications
- correct syntax for CI skip condition in GitHub Actions workflow


<a name="v0.17.9"></a>

## [v0.17.9] - 2025-04-22

### Chore

- **CHANGELOG:** update [skip-ci]

### Fix

- update command execution for Ubuntu system updates to use bash for proper command parsing
- ensure CI is skipped for commits with [skip ci] or [skip-ci] messages


<a name="v0.17.8"></a>

## [v0.17.8] - 2025-04-22

### Chore

- **CHANGELOG:** update [skip-ci]

### Feat

- update: support WSL Linux systems


<a name="v0.17.7"></a>

## [v0.17.7] - 2025-04-22

### Chore

- **CHANGELOG:** update [skip-ci]

### Feat

- add version assignment to Homebrew formula without 'v' prefix


<a name="v0.17.6"></a>

## [v0.17.6] - 2025-04-22

### Feat

- enhance system update command with verbose logging and support for isVerbose flag


<a name="v0.17.5"></a>

## [v0.17.5] - 2025-04-18

### Feat

- add step to calculate version without 'v' prefix for Homebrew publishing


<a name="v0.17.4"></a>

## [v0.17.4] - 2025-04-18

### Refactor

- streamline environment variable definitions for Homebrew publishing


<a name="v0.17.3"></a>

## [v0.17.3] - 2025-04-18

### Fix

- correct format specification for windows archive in goreleaser configuration


<a name="v0.17.2"></a>

## [v0.17.2] - 2025-04-18

### Build

- **deps:** bump golang.org/x/net from 0.37.0 to 0.38.0 ([#15](https://github.com/eng618/eng/issues/15))

### Feat

- implement version command and enhance build metadata injection

### Refactor

- rename UpdateSystemCmd to UpdateCmd for consistency and clarity
- enhance logging for config file reading with verbose output

### Test

- add unit tests for IsVerbose and SyncDirectory functions


<a name="v0.17.1"></a>

## [v0.17.1] - 2025-04-16

### Feat

- add system update logic including brew

### Refactor

- redirect output streams to log writers in StartChildProcess and PullLatestCode
- improve documentation and organization of log package functions


<a name="v0.17.0"></a>

## [v0.17.0] - 2025-04-16

### Feat

- add verbose command for managing verbose output settings

### Fix

- linter issues and minor organizations

### Refactor

- restructure command organization and add proxy management features - Moved config commands to a dedicated config package - Created dotfiles package with commands for managing dotfiles - Added system commands for managing system settings and proxy configuration - Implemented proxy configuration management with user prompts - Removed obsolete commands and files for better clarity and organization


<a name="v0.16.15"></a>

## [v0.16.15] - 2025-04-04

### Fix

- update command in Homebrew formula to use '--help' instead of '--version'


<a name="v0.16.14"></a>

## [v0.16.14] - 2025-04-04

### Fix

- correct string interpolation syntax in Homebrew formula


<a name="v0.16.13"></a>

## [v0.16.13] - 2025-04-04

### Fix

- update checksum retrieval logic and escape characters in Homebrew formula


<a name="v0.16.12"></a>

## [v0.16.12] - 2025-04-04

### Fix

- enhance Homebrew formula for multi-architecture support and improve installation logging


<a name="v0.16.11"></a>

## [v0.16.11] - 2025-04-04

### Fix

- correct syntax for accessing checksums in Homebrew formula


<a name="v0.16.10"></a>

## [v0.16.10] - 2025-04-04

### Fix

- improve checksum retrieval logic in Homebrew publish workflow


<a name="v0.16.9"></a>

## [v0.16.9] - 2025-04-04

### Fix

- update checksum retrieval method in Homebrew publish workflow to use jq


<a name="v0.16.8"></a>

## [v0.16.8] - 2025-04-04

### Fix

- simplify checksum retrieval in Homebrew publish workflow


<a name="v0.16.7"></a>

## [v0.16.7] - 2025-04-04

### Fix

- update Homebrew publish workflow to include checksum generation for artifacts


<a name="v0.16.6"></a>

## [v0.16.6] - 2025-04-04

### Fix

- update sha256 command in Homebrew formula to use awk instead of cut


<a name="v0.16.5"></a>

## [v0.16.5] - 2025-04-04

### Fix

- add publish job to download dist directory in Homebrew workflow


<a name="v0.16.4"></a>

## [v0.16.4] - 2025-04-04

### Fix

- update upload and download artifact actions to version 4 in Homebrew publish workflow


<a name="v0.16.3"></a>

## [v0.16.3] - 2025-04-04

### Fix

- add steps to persist and download dist directory in Homebrew publish workflow


<a name="v0.16.2"></a>

## [v0.16.2] - 2025-04-04

### Fix

- update sha256 command in Homebrew formula to use cut instead of awk


<a name="v0.16.1"></a>

## [v0.16.1] - 2025-04-04

### Fix

- remove unnecessary dependency declaration for Homebrew formula


<a name="v0.16.0"></a>

## [v0.16.0] - 2025-04-04

### Build

- **deps:** bump github.com/spf13/viper from 1.19.0 to 1.20.1 ([#13](https://github.com/eng618/eng/issues/13))
- **deps:** bump golangci/golangci-lint-action from 6 to 7 ([#14](https://github.com/eng618/eng/issues/14))

### Feat

- add updateSystem command to perform system updates for Ubuntu
- add findNonMovieFolders command to identify and manage non-movie directories

### Fix

- update sha256 paths and install command in Homebrew publish workflow

### Refactor

- update comments in repo.go for clarity and detail
- enhance copyFile function to support verbose logging


<a name="v0.15.13"></a>

## [v0.15.13] - 2025-03-12

### Ci

- dynamic directory


<a name="v0.15.12"></a>

## [v0.15.12] - 2025-03-11

### Ci

- update paths


<a name="v0.15.11"></a>

## [v0.15.11] - 2025-03-11

### Ci

- remove debug statements


<a name="v0.15.10"></a>

## [v0.15.10] - 2025-03-11


<a name="v0.15.9"></a>

## [v0.15.9] - 2025-03-11

### Ci

- update paths


<a name="v0.15.8"></a>

## [v0.15.8] - 2025-03-11

### Ci

- update debug statements


<a name="v0.15.7"></a>

## [v0.15.7] - 2025-03-11

### Ci

- add debug statements


<a name="v0.15.6"></a>

## [v0.15.6] - 2025-03-11

### Fix

- path correction


<a name="v0.15.5"></a>

## [v0.15.5] - 2025-03-11

### Feat

- update Homebrew formula for multi-architecture support and set Go version to stable


<a name="v0.15.4"></a>

## [v0.15.4] - 2025-03-11

### Ci

- use PAT


<a name="v0.15.3"></a>

## [v0.15.3] - 2025-03-11

### Ci

- add token deploy


<a name="v0.15.2"></a>

## [v0.15.2] - 2025-03-11

### Ci

- fixes for git commands


<a name="v0.15.1"></a>

## [v0.15.1] - 2025-03-11


<a name="v0.15.0"></a>

## [v0.15.0] - 2025-03-11

### Build

- **deps:** bump github.com/go-git/go-git/v5 from 5.13.2 to 5.14.0 ([#12](https://github.com/eng618/eng/issues/12))

### Chore

- update Go version to 1.24 in workflows and configuration files

### Ci

- add golangci-lint

### Docs

- update README with homebrew formula link; improve config error logging

### Feat

- enhance workflows with task-based commands and improved Homebrew formula


<a name="v0.14.13"></a>

## [v0.14.13] - 2025-02-12

### Feat

- update GoReleaser workflow to combine installation and execution steps


<a name="v0.14.12"></a>

## [v0.14.12] - 2025-02-12

### Feat

- add GITHUB_TOKEN environment variable to GoReleaser workflow


<a name="v0.14.11"></a>

## [v0.14.11] - 2025-02-12

### Feat

- specify GoReleaser version and distribution in Homebrew publish workflow


<a name="v0.14.10"></a>

## [v0.14.10] - 2025-02-12

### Feat

- update GoReleaser action to version 6 in Homebrew publish workflow


<a name="v0.14.9"></a>

## [v0.14.9] - 2025-02-12

### Feat

- replace manual GoReleaser installation with GitHub Action for improved workflow


<a name="v0.14.8"></a>

## [v0.14.8] - 2025-02-12

### Feat

- add GITHUB_TOKEN environment variable for GoReleaser installation


<a name="v0.14.7"></a>

## [v0.14.7] - 2025-02-12

### Feat

- integrate GoReleaser for streamlined Homebrew publishing and update asset paths


<a name="v0.14.6"></a>

## [v0.14.6] - 2025-02-11

### Feat

- add verification step for binary version in Homebrew publish workflow


<a name="v0.14.5"></a>

## [v0.14.5] - 2025-02-11


<a name="v0.14.4"></a>

## [v0.14.4] - 2025-02-11

### Chore

- update GitHub Actions to use latest versions of checkout and setup-go
- **CHANGELOG:** update [skip-ci]

### Feat

- set execute permissions for the build output in Homebrew publish workflow
- update Homebrew publish workflow to use version tags instead of branch name
- add GitHub Actions workflow for publishing to Homebrew
- add caching for Go modules and build to improve CI performance

### Reverts

- feat: add caching for Go modules and build to improve CI performance


<a name="v0.14.3"></a>

## [v0.14.3] - 2025-02-08

### Chore

- **CHANGELOG:** update [skip-ci]

### Docs

- enhance comments for SyncDirectory and copyFile functions to clarify parameters and return values
- improve comments for StartChildProcess function to clarify behavior and parameters
- enhance documentation for config and log packages

### Feat

- enhance dotfiles commands with verbose logging and error handling
- add verbose flag for enhanced output control
- add fetch and sync commands for managing dotfiles repository
- set default workTree to HOME environment variable in dotfiles configuration
- add command to set dotfiles repository path and implement related functionality


<a name="v0.14.2"></a>

## [v0.14.2] - 2025-02-08

### Build

- update all mods

### Chore

- update Go version to 1.23.6 and fix archive formats in goreleaser config
- various improvements
- **CHANGELOG:** update [skip-ci]

### Docs

- update README

### Feat

- **ts:** add new command for tailscale help
- **ts:** add new command for tailscale help


<a name="v0.14.1"></a>

## [v0.14.1] - 2024-10-31

### Build

- bump deps, and go version

### Chore

- **CHANGELOG:** update [skip-ci]

### Ci

- update releaser properties
- test latest go version
- **goreleaser:** update version


<a name="v0.14.0"></a>

## [v0.14.0] - 2024-06-12

### Build

- **deps:** upgrade all
- **deps:** bump github.com/go-git/go-git/v5 from 5.11.0 to 5.12.0
- **deps:** bump golang.org/x/net from 0.22.0 to 0.23.0 ([#6](https://github.com/eng618/eng/issues/6))
- **deps:** bump github.com/fatih/color from 1.16.0 to 1.17.0 ([#7](https://github.com/eng618/eng/issues/7))

### Chore

- **CHANGELOG:** update [skip-ci]

### Docs

- **README:** various updates
- **README:** add badges

### Feat

- **dotfiles:** add sync command
- **killPort:** add new command


<a name="v0.13.0"></a>

## [v0.13.0] - 2024-03-05

### Build

- update all dependencies
- **deps:** update all
- **deps:** bump all
- **go:** upgrade to 1.22
- **task:** add task configuration

### Chore

- add CODEOWNERS
- **CHANGELOG:** update [skip-ci]

### Ci

- **actions:** upgrade
- **dependabot:** add config file

### Docs

- **config:** update descritpions

### Feat

- **system:** add base command, and sub command
- **task:** add alias for install

### Fix

- **lint:** check error


<a name="v0.0.4"></a>

## [v0.0.4] - 2023-12-31

### Build

- **deps:** update all

### Chore

- **CHANGELOG:** update [skip-ci]


<a name="v0.0.3"></a>

## [v0.0.3] - 2023-12-31

### Chore

- **changelog:** add config and initial changelog


<a name="v0.0.2"></a>

## [v0.0.2] - 2023-12-31

### Build

- **releaser:** add config file


<a name="v0.0.1"></a>

## [v0.0.1] - 2023-12-31

### Chore

- personalize

### Ci

- update go version
- create go action

### Feat

- start config example
- **config:** confirm/update email
- **eng:** base cli generated
- **header:** update ascii


<a name="v0.12.1"></a>

## [v0.12.1] - 2023-12-19

### Build

- **makefile:** update depricated flag

### Chore

- update changelog [skip ci]

### Feat

- **db:** return set time

### Test

- **db:** add test coverage


<a name="v0.12.0"></a>

## [v0.12.0] - 2023-12-18

### Build

- remove disabled workflow
- **deps:** bump all deps
- **deps:** bump all

### Chore

- update changelog [skip ci]

### Docs

- **grammar:** updates phonetics

### Feat

- **algo:** adds anagrams package
- **db:** add in memory database example
- **leet:** add a couple answers
- **list:** standard library example
- **vowels:** add vowels algorithm package

### Fix

- **anagrams:** simplified logic

### Refactor

- **fibonacci:** clean up typos and examples
- **list:** standard library example

### Test

- remove expected output
- **slice_queue:** update examples


<a name="v0.11.8"></a>

## [v0.11.8] - 2023-10-02

### Build

- **deps:** bump actions/checkout from 3 to 4 ([#18](https://github.com/eng618/eng/issues/18))
- **deps:** bump goreleaser/goreleaser-action from 4 to 5 ([#17](https://github.com/eng618/eng/issues/17))

### Ci

- disable twitter announce


<a name="v0.11.7"></a>

## [v0.11.7] - 2023-08-18

### Build

- go1.21

### Ci

- fix action version


<a name="v0.11.6"></a>

## [v0.11.6] - 2023-08-18

### Ci

- update releaser config
- update branches


<a name="v0.11.5"></a>

## [v0.11.5] - 2023-08-13


<a name="v0.11.4"></a>

## [v0.11.4] - 2023-08-13

### Build

- **mod:** update mod


<a name="v0.11.3"></a>

## [v0.11.3] - 2023-06-10

### Chore

- update changelog [skip ci]

### Docs

- remove unneeded README

### Feat

- **examples:** add RESTFul API using Gin

### Test

- add simple test


<a name="v0.11.2"></a>

## [v0.11.2] - 2023-06-10

### Fix

- correct generics path name

### Revert

- change package name back to github hosted path ([#16](https://github.com/eng618/eng/issues/16))


<a name="v0.11.1"></a>

## [v0.11.1] - 2023-06-07

### Build

- bump go version

### Chore

- minor updates
- update changelog [skip ci]

### Docs

- add package documentation

### Feat

- add generics example ([#14](https://github.com/eng618/eng/issues/14))


<a name="v0.11.0"></a>

## [v0.11.0] - 2023-05-06

### Chore

- update changelog [skip ci]

### Ci

- update release workflow

### Fix

- fully update module name


<a name="v0.10.0"></a>

## [v0.10.0] - 2023-05-06

### Chore

- add publish command

### Ci

- remove support for go 1.18

### Feat

- update module name

### Fix

- cleanup go.mod


<a name="v0.9.2"></a>

## [v0.9.2] - 2023-04-07

### Ci

- update goreleaser


<a name="v0.9.1"></a>

## [v0.9.1] - 2023-04-07

### Chore

- minor adjustments


<a name="v0.9.0"></a>

## [v0.9.0] - 2023-04-07

### Build

- upgrade go to 1.19 latest
- bump go patch version
- **deps:** bump actions/setup-go from 3 to 4

### Chore

- update changelog [skip ci]
- update changelog [skip ci]
- update changelog template
- update changelog [skip ci]
- update changelog template
- update changelog [skip ci]
- update changelog comand
- update changelog [skip-ci]
- update changelog comand
- update changelog
- update changelog comand

### Ci

- comment out deprications
- update goreleaser
- fix go version
- add dependabot.yml
- update lint tool
- udpate to go1.20
- remove go 1.20
- add go 1.20 and update releaser version
- update actions

### Docs

- update documentation

### Feat

- go 1.20

### Test

- speed up tests un nanosecond vs second


<a name="v0.8.1"></a>

## [v0.8.1] - 2022-09-11

### Build

- add changelog command

### Chore

- update changelog
- update change log


<a name="v0.8.0"></a>

## [v0.8.0] - 2022-09-11

### Build

- upgrade go to 1.18
- update to latest go 1.17.x
- update release command

### Feat

- add package to write to a file

### Style

- apply formatting


<a name="v0.7.0"></a>

## [v0.7.0] - 2022-05-06

### Build

- add release command
- update some Makefile commands
- update remaining 1.16 references
- bump default build version to 1.17

### Chore

- create Makefile

### Ci

- update lint configuration
- remove verbose test logging
- test go 1.18 and 1.17 only
- specify coverage file
- update codecov token
- update golangci config
- only use go version n-2
- add code coverage with Codecov

### Docs

- add codecov badge
- update CHANGELOG

### Feat

- add context with timeout examples
- stub circular package


<a name="v0.6.0"></a>

## [v0.6.0] - 2021-06-17

### Docs

- update changelog [skip-ci]

### Feat

- **fibonacci:** add algorithms to calculate fib
- **queue:** add Peek method, increase test cov


<a name="v0.5.0"></a>

## [v0.5.0] - 2021-06-16

### Chore

- **lint:** correct typos and golint warnings

### Docs

- add CHANGELOG
- updated README, add module doc

### Feat

- **hash:** add hash table data structure
- **queue:** add LinkedList implementation

### Refactor

- integrated golangci-lint with config


<a name="v0.4.3"></a>

## [v0.4.3] - 2021-06-14

### Feat

- **linkedlist:** add Remove method


<a name="v0.4.2"></a>

## [v0.4.2] - 2021-06-14

### Feat

- enhanced merge sort
- made search a single package
- **queue:** add slice based queue Example
- **search:** add linear function

### Refactor

- more changes for readability
- remove Data, and simply define []int
- **linkedlist:** to be more concise

### Test

- **benchmark:** add benchmark tests


<a name="v0.4.1"></a>

## [v0.4.1] - 2021-06-11

### Ci

- fix creds


<a name="v0.4.0"></a>

## [v0.4.0] - 2021-06-11

### Ci

- add twitter creds


<a name="v0.3.0"></a>

## [v0.3.0] - 2021-06-11

### Ci

- only run goreleaser on tags


<a name="v0.2.0"></a>

## [v0.2.0] - 2021-06-11

### Refactor

- organized for ease of use


<a name="v0.1.0"></a>

## v0.1.0 - 2021-06-11

### Build

- add goreleaser and required configuration.

### Ci

- add actions to build, test, and lint

### Docs

- add Big O cheatsheet
- update readme

### Feat

- add merge sort package
- add main
- **binary:** add binary search package
- **linkedlist:** create linkedlist package
- **stack:** add stack package

### Fix

- various fixes...
- correct go mod name, and add go reportcard badge

### Refactor

- replaced how new stacks are created
- use go naming conventions for package

### Test

- add tests for merge sort
- add test cases for list
- fix delete tests
- add delete tests
- fix example output


[Unreleased]: https://github.com/eng618/eng/compare/v0.21.3...HEAD
[v0.21.3]: https://github.com/eng618/eng/compare/v0.21.2...v0.21.3
[v0.21.2]: https://github.com/eng618/eng/compare/v0.21.1...v0.21.2
[v0.21.1]: https://github.com/eng618/eng/compare/v0.21.0...v0.21.1
[v0.21.0]: https://github.com/eng618/eng/compare/v0.20.5...v0.21.0
[v0.20.5]: https://github.com/eng618/eng/compare/v0.20.4...v0.20.5
[v0.20.4]: https://github.com/eng618/eng/compare/v0.20.3...v0.20.4
[v0.20.3]: https://github.com/eng618/eng/compare/v0.20.2...v0.20.3
[v0.20.2]: https://github.com/eng618/eng/compare/v0.20.1...v0.20.2
[v0.20.1]: https://github.com/eng618/eng/compare/v0.20.0...v0.20.1
[v0.20.0]: https://github.com/eng618/eng/compare/v0.19.8...v0.20.0
[v0.19.8]: https://github.com/eng618/eng/compare/v0.19.7...v0.19.8
[v0.19.7]: https://github.com/eng618/eng/compare/v0.19.6...v0.19.7
[v0.19.6]: https://github.com/eng618/eng/compare/v0.19.5...v0.19.6
[v0.19.5]: https://github.com/eng618/eng/compare/v0.19.4...v0.19.5
[v0.19.4]: https://github.com/eng618/eng/compare/v0.19.3...v0.19.4
[v0.19.3]: https://github.com/eng618/eng/compare/v0.19.2...v0.19.3
[v0.19.2]: https://github.com/eng618/eng/compare/v0.19.1...v0.19.2
[v0.19.1]: https://github.com/eng618/eng/compare/v0.19.0...v0.19.1
[v0.19.0]: https://github.com/eng618/eng/compare/v0.18.4...v0.19.0
[v0.18.4]: https://github.com/eng618/eng/compare/v0.18.3...v0.18.4
[v0.18.3]: https://github.com/eng618/eng/compare/v0.18.2...v0.18.3
[v0.18.2]: https://github.com/eng618/eng/compare/v0.18.1...v0.18.2
[v0.18.1]: https://github.com/eng618/eng/compare/v0.18.0...v0.18.1
[v0.18.0]: https://github.com/eng618/eng/compare/v0.17.22...v0.18.0
[v0.17.22]: https://github.com/eng618/eng/compare/v0.17.21...v0.17.22
[v0.17.21]: https://github.com/eng618/eng/compare/v0.17.20...v0.17.21
[v0.17.20]: https://github.com/eng618/eng/compare/v0.17.19...v0.17.20
[v0.17.19]: https://github.com/eng618/eng/compare/v0.17.18...v0.17.19
[v0.17.18]: https://github.com/eng618/eng/compare/v0.17.17...v0.17.18
[v0.17.17]: https://github.com/eng618/eng/compare/v0.17.16...v0.17.17
[v0.17.16]: https://github.com/eng618/eng/compare/v0.17.15...v0.17.16
[v0.17.15]: https://github.com/eng618/eng/compare/v0.17.14...v0.17.15
[v0.17.14]: https://github.com/eng618/eng/compare/v0.17.13...v0.17.14
[v0.17.13]: https://github.com/eng618/eng/compare/v0.17.12...v0.17.13
[v0.17.12]: https://github.com/eng618/eng/compare/v0.17.11...v0.17.12
[v0.17.11]: https://github.com/eng618/eng/compare/v0.17.10...v0.17.11
[v0.17.10]: https://github.com/eng618/eng/compare/v0.17.9...v0.17.10
[v0.17.9]: https://github.com/eng618/eng/compare/v0.17.8...v0.17.9
[v0.17.8]: https://github.com/eng618/eng/compare/v0.17.7...v0.17.8
[v0.17.7]: https://github.com/eng618/eng/compare/v0.17.6...v0.17.7
[v0.17.6]: https://github.com/eng618/eng/compare/v0.17.5...v0.17.6
[v0.17.5]: https://github.com/eng618/eng/compare/v0.17.4...v0.17.5
[v0.17.4]: https://github.com/eng618/eng/compare/v0.17.3...v0.17.4
[v0.17.3]: https://github.com/eng618/eng/compare/v0.17.2...v0.17.3
[v0.17.2]: https://github.com/eng618/eng/compare/v0.17.1...v0.17.2
[v0.17.1]: https://github.com/eng618/eng/compare/v0.17.0...v0.17.1
[v0.17.0]: https://github.com/eng618/eng/compare/v0.16.15...v0.17.0
[v0.16.15]: https://github.com/eng618/eng/compare/v0.16.14...v0.16.15
[v0.16.14]: https://github.com/eng618/eng/compare/v0.16.13...v0.16.14
[v0.16.13]: https://github.com/eng618/eng/compare/v0.16.12...v0.16.13
[v0.16.12]: https://github.com/eng618/eng/compare/v0.16.11...v0.16.12
[v0.16.11]: https://github.com/eng618/eng/compare/v0.16.10...v0.16.11
[v0.16.10]: https://github.com/eng618/eng/compare/v0.16.9...v0.16.10
[v0.16.9]: https://github.com/eng618/eng/compare/v0.16.8...v0.16.9
[v0.16.8]: https://github.com/eng618/eng/compare/v0.16.7...v0.16.8
[v0.16.7]: https://github.com/eng618/eng/compare/v0.16.6...v0.16.7
[v0.16.6]: https://github.com/eng618/eng/compare/v0.16.5...v0.16.6
[v0.16.5]: https://github.com/eng618/eng/compare/v0.16.4...v0.16.5
[v0.16.4]: https://github.com/eng618/eng/compare/v0.16.3...v0.16.4
[v0.16.3]: https://github.com/eng618/eng/compare/v0.16.2...v0.16.3
[v0.16.2]: https://github.com/eng618/eng/compare/v0.16.1...v0.16.2
[v0.16.1]: https://github.com/eng618/eng/compare/v0.16.0...v0.16.1
[v0.16.0]: https://github.com/eng618/eng/compare/v0.15.13...v0.16.0
[v0.15.13]: https://github.com/eng618/eng/compare/v0.15.12...v0.15.13
[v0.15.12]: https://github.com/eng618/eng/compare/v0.15.11...v0.15.12
[v0.15.11]: https://github.com/eng618/eng/compare/v0.15.10...v0.15.11
[v0.15.10]: https://github.com/eng618/eng/compare/v0.15.9...v0.15.10
[v0.15.9]: https://github.com/eng618/eng/compare/v0.15.8...v0.15.9
[v0.15.8]: https://github.com/eng618/eng/compare/v0.15.7...v0.15.8
[v0.15.7]: https://github.com/eng618/eng/compare/v0.15.6...v0.15.7
[v0.15.6]: https://github.com/eng618/eng/compare/v0.15.5...v0.15.6
[v0.15.5]: https://github.com/eng618/eng/compare/v0.15.4...v0.15.5
[v0.15.4]: https://github.com/eng618/eng/compare/v0.15.3...v0.15.4
[v0.15.3]: https://github.com/eng618/eng/compare/v0.15.2...v0.15.3
[v0.15.2]: https://github.com/eng618/eng/compare/v0.15.1...v0.15.2
[v0.15.1]: https://github.com/eng618/eng/compare/v0.15.0...v0.15.1
[v0.15.0]: https://github.com/eng618/eng/compare/v0.14.13...v0.15.0
[v0.14.13]: https://github.com/eng618/eng/compare/v0.14.12...v0.14.13
[v0.14.12]: https://github.com/eng618/eng/compare/v0.14.11...v0.14.12
[v0.14.11]: https://github.com/eng618/eng/compare/v0.14.10...v0.14.11
[v0.14.10]: https://github.com/eng618/eng/compare/v0.14.9...v0.14.10
[v0.14.9]: https://github.com/eng618/eng/compare/v0.14.8...v0.14.9
[v0.14.8]: https://github.com/eng618/eng/compare/v0.14.7...v0.14.8
[v0.14.7]: https://github.com/eng618/eng/compare/v0.14.6...v0.14.7
[v0.14.6]: https://github.com/eng618/eng/compare/v0.14.5...v0.14.6
[v0.14.5]: https://github.com/eng618/eng/compare/v0.14.4...v0.14.5
[v0.14.4]: https://github.com/eng618/eng/compare/v0.14.3...v0.14.4
[v0.14.3]: https://github.com/eng618/eng/compare/v0.14.2...v0.14.3
[v0.14.2]: https://github.com/eng618/eng/compare/v0.14.1...v0.14.2
[v0.14.1]: https://github.com/eng618/eng/compare/v0.14.0...v0.14.1
[v0.14.0]: https://github.com/eng618/eng/compare/v0.13.0...v0.14.0
[v0.13.0]: https://github.com/eng618/eng/compare/v0.0.4...v0.13.0
[v0.0.4]: https://github.com/eng618/eng/compare/v0.0.3...v0.0.4
[v0.0.3]: https://github.com/eng618/eng/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/eng618/eng/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/eng618/eng/compare/v0.12.1...v0.0.1
[v0.12.1]: https://github.com/eng618/eng/compare/v0.12.0...v0.12.1
[v0.12.0]: https://github.com/eng618/eng/compare/v0.11.8...v0.12.0
[v0.11.8]: https://github.com/eng618/eng/compare/v0.11.7...v0.11.8
[v0.11.7]: https://github.com/eng618/eng/compare/v0.11.6...v0.11.7
[v0.11.6]: https://github.com/eng618/eng/compare/v0.11.5...v0.11.6
[v0.11.5]: https://github.com/eng618/eng/compare/v0.11.4...v0.11.5
[v0.11.4]: https://github.com/eng618/eng/compare/v0.11.3...v0.11.4
[v0.11.3]: https://github.com/eng618/eng/compare/v0.11.2...v0.11.3
[v0.11.2]: https://github.com/eng618/eng/compare/v0.11.1...v0.11.2
[v0.11.1]: https://github.com/eng618/eng/compare/v0.11.0...v0.11.1
[v0.11.0]: https://github.com/eng618/eng/compare/v0.10.0...v0.11.0
[v0.10.0]: https://github.com/eng618/eng/compare/v0.9.2...v0.10.0
[v0.9.2]: https://github.com/eng618/eng/compare/v0.9.1...v0.9.2
[v0.9.1]: https://github.com/eng618/eng/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/eng618/eng/compare/v0.8.1...v0.9.0
[v0.8.1]: https://github.com/eng618/eng/compare/v0.8.0...v0.8.1
[v0.8.0]: https://github.com/eng618/eng/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/eng618/eng/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/eng618/eng/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/eng618/eng/compare/v0.4.3...v0.5.0
[v0.4.3]: https://github.com/eng618/eng/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/eng618/eng/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/eng618/eng/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/eng618/eng/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/eng618/eng/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/eng618/eng/compare/v0.1.0...v0.2.0
