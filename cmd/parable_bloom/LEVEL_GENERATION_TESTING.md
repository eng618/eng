# Level Generation Testing & Quality Assurance

## Overview

The Parable Bloom level generator produces **100% solvable levels** through an iterative retry mechanism combined with solver-aware vine placement. This document describes the testing strategy and quality benchmarks.

## Key Guarantees

- **100% Solvability**: Every generated level is validated as solvable before being written to disk
- **Proper Difficulty Progression**: Each module follows Seedling → Sprout → Nurturing → Flourishing → Transcendent
- **No Corrupted Files**: Atomic writes with sanity checks prevent partial/truncated JSON files
- **Deterministic Generation**: Same seed always produces the same level (reproducible)

## Testing Strategy: BET (Benchmark Example Test)

Tests follow the order: **Example → Test → Benchmark** for clarity and progression.

### File Organization

```
level_generate_example_test.go  # Example tests showing usage patterns
level_generate_test.go           # Unit tests for correctness
level_generate_benchmark_test.go # Performance benchmarks
```

### Test Categories

#### 1. Example Tests (`*_example_test.go`)

- Show how to use the generator API
- Demonstrate expected behavior
- Run as part of `go test` but focus on clarity

**Examples included:**

- `ExampleGenerateLevel_Seedling` - Generate a simple Seedling level
- `ExampleGenerateVines_Difficulty` - Generate vines for different difficulties
- `ExampleDifficultyForLevel_Module` - Show difficulty progression across modules

#### 2. Unit Tests (`*_test.go`)

- Verify correctness of core functions
- Test edge cases and error handling
- Assert invariants and properties

**Tests included:**

- `TestDifficultyForLevel_Progression` - Verify difficulty progression
- `TestGenerateVines_SolverValidation` - Ensure all generated vines are solvable
- `TestGenerateVines_GridSize` - Test various grid sizes
- `TestGenerateLevel_Occupancy` - Verify grid occupancy requirements
- `TestGenerateVines_ColorDistribution` - Ensure proper color variation

#### 3. Benchmark Tests (`*_benchmark_test.go`)

- Measure performance characteristics
- Establish baselines for optimization
- Track performance regression

**Benchmarks included:**

- `BenchmarkGenerateVines_Seedling` - Fast easy levels (target: <10ms)
- `BenchmarkGenerateVines_Nurturing` - Medium difficulty (target: <50ms)
- `BenchmarkGenerateVines_Transcendent` - Hard levels (target: <500ms)
- `BenchmarkGenerateLevel_FullPipeline` - End-to-end generation
- `BenchmarkSolver_IsSolvableGreedy` - Solver performance baseline

## Performance Targets

### Generation Time Per Level

| Difficulty | Target | Notes |
|---|---|---|
| **Seedling** | < 10ms | Uses fast algorithm |
| **Sprout** | < 20ms | Still fast |
| **Nurturing** | < 50ms | Uses solver-aware placement |
| **Flourishing** | < 100ms | Solver-aware, requires more retries |
| **Transcendent** | < 500ms | Hardest, most retries expected |

### Batch Generation (10 levels per module)

- **Target**: < 5 seconds per module
- **Maximum**: < 30 seconds for module 8 (largest)
- **Typical**: 2-5 seconds per module

## Running Tests

### Run all tests

```bash
go test ./cmd/parable_bloom -v
```

### Run only examples

```bash
go test ./cmd/parable_bloom -run Example -v
```

### Run only unit tests

```bash
go test ./cmd/parable_bloom -run TestGenerate -v
```

### Run only benchmarks

```bash
go test ./cmd/parable_bloom -bench=. -benchmem -benchtime=3s
```

### Run benchmarks with profiling

```bash
go test ./cmd/parable_bloom -bench=BenchmarkGenerateVines -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Run with coverage

```bash
go test ./cmd/parable_bloom -cover
go test ./cmd/parable_bloom -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Quality Metrics

### Code Quality

- **Cyclomatic Complexity**: All functions < 20 (measured via golangci-lint)
- **Test Coverage**: Target > 80% for `level_generate.go`
- **Linting**: 0 issues reported by golangci-lint

### Generation Quality

- **Solvability Rate**: 100% (all generated levels must be solvable)
- **Occupancy Accuracy**: Full coverage of visible cells (100% required; 99% allowed with mask)
- **Difficulty Distribution**: Correct progression per module
- **Color Variation**: 3-5 distinct colors per level

### Validation Results

```
Levels: 95/95 valid (100%)
Violations: 0
Warnings: 298 (mostly direction distribution advisories - acceptable)
```

## Continuous Integration

### Pre-commit

```bash
task lint-fix  # Fix formatting and linting issues
go test ./cmd/parable_bloom -v  # Run all tests
```

### CI Pipeline

1. Build binary (`go build`)
2. Run linter (`golangci-lint run`)
3. Run unit tests (`go test`)
4. Run benchmarks (record baseline)
5. Generate sample levels (validate with eng validator)
6. Generate coverage report

## Performance Regression Detection

Benchmarks are recorded in CI to detect performance regressions:

```bash
# Baseline run (commit reference)
go test ./cmd/parable_bloom -bench=. -benchmem > baseline.txt

# Current run
go test ./cmd/parable_bloom -bench=. -benchmem > current.txt

# Compare
benchstat baseline.txt current.txt
```

**Acceptable regression**: ±10% variance acceptable day-to-day
**Warning threshold**: >15% slower than baseline
**Fail threshold**: >25% slower than baseline

## Solver-Aware Placement Strategy

For **Nurturing, Flourishing, and Transcendent** difficulties:

1. Place vine candidate
2. Validate level with `Solver.IsSolvableGreedy()`
3. Accept only if solvable
4. Retry with different seed if unsolvable

This guarantees all generated vines create solvable levels.

## Tiling-first Generator (Tile the grid, then solver-gate)

The generator uses a "tiling-first, solvability-second" approach: create a complete tiling of the visible grid into self-avoiding vines, then pass the result to the solver/validator as a gating step. This keeps generator complexity low while producing varied, solvable configurations.

High-level pseudocode (see documentation for details):

```pseudo
function GenerateLevel(tier, seed):
  rng <- RNG(seed)

  for attempt in 1..MAX_ATTEMPTS:
    (w,h) <- pickGridSize(tier, rng)
    profile <- pickVarietyProfile(tier, rng) # lengthMix, turnMix, regionBias, dirBalance

    (grid, vines) <- TileGridIntoVines(w, h, tier.vineCountRange, profile, rng)
    if FAIL: continue

    if not GeometryValid(vines, w, h): continue
    if not FullCoverageNoOverlap(grid): continue

    if not MeetsVariety(vines, profile): continue

    result <- SolveAndValidate(w, h, vines, tier.maxMoves)
    if result.UNSOLVABLE: continue

    return BuildLevelJson(w,h,vines,result)

  return FAIL
```

### Variation & Quality Guidelines

- Avg Length Targets (GDD): Seedling 6–8, Sprout 5–7, Nurturing 4–6, Flourishing 3–5, Transcendent 2–4.
- Length mix: enforce that each level contains at least 3 buckets (short/medium/long) to avoid monotony.
- Turniness: track `turnDensity = turns / (len-1)` and ensure a mix (some near-straight, some bendy).
- Coverage distribution: vary levels between Edge-heavy, Center-heavy, and Balanced profiles.
- Directional balance: cap any head direction to <= 40% of vines.

### Anti-Repeat (Rolling Window)

Maintain a rolling window of the last N generated levels and reject (or re-roll) if a candidate is too similar by comparing:

- Length histogram
- TurnDensity histogram
- Head-direction histogram
- Coarse 3×3 region heatmap of vine length placement

This reduces perceived repetition across level batches and keeps modules feeling fresh.

## Retry Statistics

Typical retry counts when searching for solvable configurations:

| Module | Seedling | Nurturing | Transcendent |
|---|---|---|---|
| **Module 2-3** | 1-5 | 10-80 | 50-100 |
| **Module 4-5** | 1-10 | 30-250 | 100-500 |
| **Module 6-8** | 1-20 | 50-1000 | 200-1000+ |

## Future Improvements

- [ ] Parallel generation within solver-aware placement
- [ ] ML-based vine placement heuristics
- [ ] Caching of known-good configurations
- [ ] Interactive level editor with real-time validation
- [ ] Analytics on level difficulty metrics

## References

- [Level System Reference](../parable-bloom/docs/LEVEL_SYSTEM.md)
- [Generator Implementation](./level_generate.go)
- [Solver Algorithm](./solver.go)
