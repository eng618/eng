/*
DEPRECATED: The 'eng parable-bloom' command is deprecated and will be removed. Use the standalone
`parble-bloom/tools/level-builder` tool instead. Example:

	cd parable-bloom/tools/level-builder && go build && ./level-builder --help

Package parable_bloom provides level management and validation tools for the Parable Bloom game.

This package implements two main commands accessible via the eng CLI:
  - parable-bloom level-validate: Validates level files for structural integrity and solvability
  - parable-bloom level-generate: Procedurally generates solvable game levels

# Level Structure

Levels are defined in JSON format with the following structure:
  - id: Unique integer level identifier
  - name: Human-readable level name
  - difficulty: One of Tutorial, Seedling, Sprout, Nurturing, Flourishing, Transcendent
  - grid_size: [width, height] of the game grid
  - vines: Array of Vine objects (game entities to be cleared)
  - max_moves: Maximum allowed moves to complete the level
  - min_moves: Minimum expected moves for optimal completion
  - grace: Grace period (lives/attempts) for the player
  - complexity: Difficulty indicator (simple, medium, complex, challenging, expert)
  - mask: Optional visual masking for grid cells (mode: show-all, show, hide)

# Validation

The validator performs 8 comprehensive checks:
 1. Basic structure: Required fields, valid types, positive IDs/counts
 2. Vine paths: Contiguity, Manhattan distance validation, directional consistency
 3. Grid occupancy: Require full coverage (100% of visible cells; 99% allowed when mask hides cells) (30% for Tutorial only during early tutorial adjustments)
 4. Colors: Valid color names, count ranges, max 35% per color
 5. Vine lengths: Within difficulty-specific ranges (e.g., 6-8 segments for Seedling)
 6. Blocking relationships: Acyclic, at least one clearable vine
 7. Directional balance: Expected percentage ranges per direction
 8. Difficulty compliance: Vine count and grace value checks

Violations are errors that prevent level use, while warnings are advisories that may
indicate design issues. Solvability is checked using:
  - Greedy solver (default, O(n²)): Fast heuristic for generation
  - BFS solver (--strict flag): Thorough exploration of all removal orderings

# Generation

The level generator creates solvable levels using deterministic seeding:
  - Same levelID always produces the same level (reproducible)
  - Difficulty-based grid sizing and vine parameters
  - Target 95% grid occupancy with procedural vine placement
  - Color distribution across 5-color palette
  - Vine head-direction validation during generation

Batch generation supports module-level concurrency:
  - One goroutine per module range (Tutorial, Seedling, etc.)
  - Semaphore limiting for resource control
  - Parallel validation before file output

# Module System

Levels are organized into modules (Parable chapters):
  - Tutorial: Levels 1-5 (simple introduction)
  - The Mustard Seed: Levels 6-20 (Seedling tier)
  - The Sower: Levels 21-35 (Sprout tier)
  - Wheat and Weeds: Levels 36-50 (Nurturing tier)
  - The Lost Sheep: Levels 51-65 (Flourishing tier)
  - The Prodigal Son: Levels 66-80 (Flourishing tier)
  - The Hidden Treasure: Levels 81-90 (Transcendent tier)
  - The Pearl of Great Price: Levels 91-100 (Transcendent tier)

Modules are loaded from assets/data/modules.json with sensible defaults if not found.

# Package Structure

The package is organized into focused modules:
  - models.go: Data structures for levels, vines, modules, and difficulty specs
  - validators.go: Comprehensive validation logic with detailed error reporting
  - solver.go: Solvability checking using greedy and BFS algorithms
  - modules.go: Module loading and difficulty tier mapping
  - file_io.go: Safe level file reading and writing operations
  - utils.go: Utility functions for calculations and lookups
  - level_validate.go: CLI command for validating levels
  - level_generate.go: CLI command for procedurally generating levels

# Usage Examples

Validate a single level file:

eng parable-bloom level-validate --file assets/levels/level_1.json

Validate all levels in a directory with solvability checks:

eng parable-bloom level-validate --directory assets/levels --check-solvability

Generate a single level with default parameters:

eng parable-bloom level-generate --name "My Level" --stdout

Generate 15 levels for Module 2 (The Mustard Seed, Seedling tier):

eng pb level-generate --module 2 --count 15 --output assets/levels

Generate levels for all modules 2-8 (95 levels total):

eng pb level-generate --module 2 --count 15 --output assets/levels  # Mustard Seed (6-20)
eng pb level-generate --module 3 --count 15 --output assets/levels  # Sower (21-35)
eng pb level-generate --module 4 --count 15 --output assets/levels  # Wheat and Weeds (36-50)
eng pb level-generate --module 5 --count 15 --output assets/levels  # Lost Sheep (51-65)
eng pb level-generate --module 6 --count 15 --output assets/levels  # Prodigal Son (66-80)
eng pb level-generate --module 7 --count 10 --output assets/levels  # Hidden Treasure (81-90)
eng pb level-generate --module 8 --count 10 --output assets/levels  # Pearl (91-100)

Generate levels for all modules 2-8 (95 levels):

for id in {2..8}; do eng pb level-generate --module $id --count 15 --output assets/levels; done

# Error Handling

The package uses Go's standard error wrapping with context:
  - File I/O errors include file path and operation
  - Validation errors return both violations and warnings
  - CLI commands exit with code 1 on validation failure

All JSON parsing uses strict mode (DisallowUnknownFields) to catch schema drift.
*/
package parable_bloom
