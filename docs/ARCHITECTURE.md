# Eng CLI - Parable Bloom Subcommand Architecture

## Overview

The `parable_bloom` subcommand in the `eng` CLI provides tools for generating, validating, rendering, and repairing levels for the Parable Bloom game.

## Directory Structure

```text
cmd/parable_bloom/
├── parable_bloom.go          # Main command definition
├── common/                   # Shared types and utilities
│   ├── models.go            # Core data structures (Level, Vine, Point, etc.)
│   ├── utils.go             # Common utility functions
│   ├── validators.go        # Validation logic
│   ├── file_io.go           # File I/O operations
│   ├── solver.go            # Level solving algorithms
│   ├── modules.go           # Module definitions and ranges
│   ├── presets.go           # Generation presets and profiles
│   └── validate_fast.go     # Fast validation routines
├── generate/                # Level generation subcommand
│   ├── level_generate.go    # Main generation logic
│   ├── generation.go        # Core generation algorithms
│   ├── tile_gen.go          # Tile-based generation
│   ├── cmd_param_sweep.go   # Parameter sweep functionality
│   └── *_test.go            # Unit and benchmark tests
├── validate/                # Level validation subcommand
│   └── level_validate.go    # Validation implementation
├── render/                  # Level rendering subcommand
│   └── level_render.go      # Rendering to various formats
└── repair/                  # Level repair subcommand
    └── level_repair.go      # Repair and fixing logic
```

## Package Organization

- **`parable_bloom`**: Root package containing the main command and subcommand registration
- **`common`**: Shared package containing types, utilities, and algorithms used across subcommands
- **`generate`**: Package for level generation functionality
- **`validate`**: Package for level validation
- **`render`**: Package for level rendering/output
- **`repair`**: Package for level repair and fixing

## Key Design Decisions

1. **Separated Packages**: Each subcommand is in its own package to maintain clear boundaries and enable focused testing
2. **Shared Common Package**: Common types and utilities are centralized to avoid duplication
3. **Import Prefixing**: Shared types are prefixed with `common.` for clarity
4. **Exported Functions**: Functions needed across packages are properly exported
5. **No Circular Imports**: Package dependencies flow from subcommands to common, not vice versa

## Testing

- Unit tests are co-located with their respective packages
- Benchmark tests ensure performance requirements are met
- Integration tests validate end-to-end functionality

## Dependencies

- Internal utils moved to `internal/utils/` for private utilities
- External dependencies minimized to maintain CLI portability
