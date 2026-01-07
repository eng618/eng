package generate

import (
	"math/rand"
	"testing"

	"github.com/eng618/eng/cmd/parable_bloom/common"
)

// BenchmarkGenerateVines_Seedling benchmarks fast easy level generation.
// Target: < 10ms.
func BenchmarkGenerateVines_Seedling(b *testing.B) {
	gridSize := common.GridSizeForLevel(7, "Seedling")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := generateVines(gridSize, "Seedling", 1000+i, 0, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateVines_Sprout benchmarks medium level generation.
// Target: < 20ms.
func BenchmarkGenerateVines_Sprout(b *testing.B) {
	gridSize := common.GridSizeForLevel(10, "Sprout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := generateVines(gridSize, "Sprout", 2000+i, 0, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateVines_Nurturing benchmarks harder level generation with solver-aware placement.
// Target: < 50ms.
func BenchmarkGenerateVines_Nurturing(b *testing.B) {
	gridSize := common.GridSizeForLevel(15, "Nurturing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := generateVines(gridSize, "Nurturing", 3000+i, 0, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateVines_Flourishing benchmarks hard level generation.
// Target: < 100ms.
func BenchmarkGenerateVines_Flourishing(b *testing.B) {
	gridSize := common.GridSizeForLevel(18, "Flourishing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := generateVines(gridSize, "Flourishing", 4000+i, 0, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateVines_Transcendent benchmarks hardest level generation.
// Target: < 500ms (may require many retries).
func BenchmarkGenerateVines_Transcendent(b *testing.B) {
	gridSize := common.GridSizeForLevel(20, "Transcendent")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := generateVines(gridSize, "Transcendent", 5000+i, 0, false)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateLevel_FullPipeline benchmarks end-to-end level generation including all validation.
// Target: < 100ms per level.
func BenchmarkGenerateLevel_FullPipeline(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levelID := 15 + (i % 10)
		GenerateLevel(levelID, "Bench", "Nurturing", 0, 0, false, 0, false)
	}
}

// BenchmarkSolver_IsSolvableGreedy benchmarks the greedy solvability checker.
// Target: < 5ms per level (used frequently during generation).
func BenchmarkSolver_IsSolvableGreedy(b *testing.B) {
	level := GenerateLevel(42, "Bench", "Nurturing", 0, 0, false, 0, false)
	solver := common.NewSolver(level)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.IsSolvableGreedy()
	}
}

// BenchmarkSolver_IsSolvableBFS benchmarks the thorough BFS solvability checker.
// Target: < 50ms per level (used for validation, not generation).
func BenchmarkSolver_IsSolvableBFS(b *testing.B) {
	level := GenerateLevel(42, "Bench", "Nurturing", 0, 0, false, 0, false)
	solver := common.NewSolver(level)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		solver.IsSolvableBFS()
	}
}

// BenchmarkCalculateBlocking benchmarks the blocking relationship calculation.
// Target: < 2ms per level.
func BenchmarkCalculateBlocking(b *testing.B) {
	gridSize := common.GridSizeForLevel(15, "Nurturing")
	vines, _, err := generateVines(gridSize, "Nurturing", 12345, 0, false)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateBlocking(vines, gridSize)
	}
}

// BenchmarkBuildVines_Fast benchmarks the fast vine building algorithm (Seedling/Sprout).
// Target: < 5ms.
func BenchmarkBuildVines_Fast(b *testing.B) {
	gridSize := [2]int{8, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildVinesFast(gridSize, "Seedling", int64(i))
	}
}

// BenchmarkTileGridIntoVines_Seedling compares tiling-first generator vs legacy fast generator.
func BenchmarkTileGridVsLegacy_Seedling(b *testing.B) {
	gridSize := common.GridSizeForLevel(7, "Seedling")
	spec := common.DifficultySpecs["Seedling"]
	cfg := common.GeneratorConfig{MaxSeedRetries: 20, LocalRepairRadius: 2, RepairRetries: 3}

	b.Run("tiling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rng := rand.New(rand.NewSource(int64(i)))
			_, _ = TileGridIntoVines(gridSize, spec, common.GetPresetProfile("Seedling"), cfg, rng)
		}
	})
	b.Run("legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = buildVinesFast(gridSize, "Seedling", int64(i))
		}
	})
}

// BenchmarkBuildVines_SolverAware benchmarks the solver-aware vine building (Nurturing+).
// Target: < 50ms (includes solver validation).
func BenchmarkBuildVines_SolverAware(b *testing.B) {
	gridSize := [2]int{12, 16}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildVinesSolverAware(gridSize, "Nurturing", int64(i))
	}
}

// BenchmarkTileGridVsLegacy_Nurturing compares tiling-first vs solver-aware legacy generator.
func BenchmarkTileGridVsLegacy_Nurturing(b *testing.B) {
	gridSize := common.GridSizeForLevel(15, "Nurturing")
	spec := common.DifficultySpecs["Nurturing"]
	cfg := common.GeneratorConfig{MaxSeedRetries: 30, LocalRepairRadius: 3, RepairRetries: 4}

	b.Run("tiling", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rng := rand.New(rand.NewSource(int64(i)))
			_, _ = TileGridIntoVines(gridSize, spec, common.GetPresetProfile("Nurturing"), cfg, rng)
		}
	})
	b.Run("legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = buildVinesSolverAware(gridSize, "Nurturing", int64(i))
		}
	})
}

// BenchmarkDifficultyForLevel benchmarks difficulty determination.
// Target: < 1μs (very fast, called frequently).
func BenchmarkDifficultyForLevel(b *testing.B) {
	modules, err := common.LoadModules("")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levelID := 6 + (i % 95)
		common.DifficultyForLevel(levelID, modules)
	}
}

// BenchmarkGridSizeForLevel benchmarks grid size determination.
// Target: < 5μs (very fast, called once per level).
func BenchmarkGridSizeForLevel(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		levelID := 6 + (i % 95)
		common.GridSizeForLevel(levelID, "Nurturing")
	}
}

// BenchmarkValidateLevel benchmarks the full validation pipeline.
// Target: < 10ms per level.
func BenchmarkValidateLevel(b *testing.B) {
	level := GenerateLevel(42, "Bench", "Nurturing", 0, 0, false, 0, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		level.Validate()
	}
}

// BenchmarkCompleteModule simulates generating an entire module (15 levels).
// Target: < 2 seconds for Seedling module, < 30 seconds for Transcendent.
func BenchmarkCompleteModule(b *testing.B) {
	modules, err := common.LoadModules("")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate module 2 (Seedling/Sprout difficulty range)
		for levelID := 6; levelID <= 20; levelID++ {
			GenerateLevel(levelID, "Bench", common.DifficultyForLevel(levelID, modules), 0, 0, false, 0, false)
		}
	}
}

// BenchmarkMemory_GenerateLevel benchmarks memory allocation during level generation.
// Shows peak memory usage during generation process.
func BenchmarkMemory_GenerateLevel(b *testing.B) {
	b.Run("Seedling", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenerateLevel(7, "Bench", "Seedling", 0, 0, false, 0, false)
		}
	})

	b.Run("Nurturing", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenerateLevel(15, "Bench", "Nurturing", 0, 0, false, 0, false)
		}
	})

	b.Run("Transcendent", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			GenerateLevel(20, "Bench", "Transcendent", 0, 0, false, 0, false)
		}
	})
}

// BenchmarkParallel_MultipleModules simulates parallel module generation.
// Shows performance under concurrent generation (like the actual CLI).
func BenchmarkParallel_MultipleModules(b *testing.B) {
	modules, err := common.LoadModules("")
	if err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			levelID := 6 + (i % 95)
			GenerateLevel(levelID, "Bench", common.DifficultyForLevel(levelID, modules), 0, 0, false, 0, false)
			i++
		}
	})
}
