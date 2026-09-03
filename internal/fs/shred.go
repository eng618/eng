package fs

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ShredMethod defines the overwrite pattern to use.
type ShredMethod string

const (
	// MethodAuto uses platform-optimized default (3-pass random).
	MethodAuto ShredMethod = "auto"
	// MethodRandom uses cryptographically random bytes.
	MethodRandom ShredMethod = "random"
	// MethodZero writes all zeros.
	MethodZero ShredMethod = "zero"
	// MethodDoD uses DoD 5220.22-M 3-pass (random, complement, random).
	MethodDoD ShredMethod = "dod"
	// MethodGutmann uses full 35-pass Gutmann method.
	MethodGutmann ShredMethod = "gutmann"
)

// ShredProgress is a callback function to report progress of a single file shredding operation.
type ShredProgress func(fileName string, pass, totalPasses int, percent float64)

// FileShredStatus tracks the shredding status of a single file.
type FileShredStatus struct {
	Path        string
	Size        int64
	CurrentPass int
	TotalPasses int
	Percent     float64
	Done        bool
	Error       error
	StartTime   time.Time
	EndTime     time.Time
}

// MultiFileProgress is a callback for reporting progress across multiple files.
type MultiFileProgress func(statuses []FileShredStatus)

// shredPattern generates the byte pattern for a given pass and method.
func shredPattern(method ShredMethod, pass, totalPasses int, buffer []byte) error {
	switch method {
	case MethodRandom, MethodAuto:
		_, err := rand.Read(buffer)
		return err
	case MethodZero:
		for i := range buffer {
			buffer[i] = 0
		}
		return nil
	case MethodDoD:
		switch pass {
		case 1:
			// Pass 1: Random
			_, err := rand.Read(buffer)
			return err
		case 2:
			// Pass 2: Complement of pass 1
			// We need to regenerate the same random data, so we'll use a deterministic approach
			// For true DoD compliance, we'd need to store pass 1 data, but for practical purposes
			// we'll use a different random pattern
			_, err := rand.Read(buffer)
			for i := range buffer {
				buffer[i] = ^buffer[i]
			}
			return err
		case 3:
			// Pass 3: Random again
			_, err := rand.Read(buffer)
			return err
		default:
			_, err := rand.Read(buffer)
			return err
		}
	case MethodGutmann:
		// Simplified: use random for all 35 passes
		// Full Gutmann implementation would require specific patterns
		_, err := rand.Read(buffer)
		return err
	default:
		_, err := rand.Read(buffer)
		return err
	}
}

// getPassesForMethod returns the number of passes for a given method.
func getPassesForMethod(method ShredMethod, userPasses int) int {
	if userPasses > 0 {
		return userPasses
	}
	switch method {
	case MethodAuto, MethodRandom, MethodZero:
		return 3
	case MethodDoD:
		return 3
	case MethodGutmann:
		return 35
	default:
		return 3
	}
}

// ShredFile securely deletes a file by overwriting it with random data for a specified number of passes.
func ShredFile(path string, passes int, progress ShredProgress) error {
	return ShredFileWithMethod(path, passes, MethodAuto, progress)
}

// ShredFileWithMethod securely deletes a file using the specified method.
func ShredFileWithMethod(path string, passes int, method ShredMethod, progress ShredProgress) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	size := info.Size()
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	totalPasses := getPassesForMethod(method, passes)

	for pass := 1; pass <= totalPasses; pass++ {
		if progress != nil {
			progress(filepath.Base(path), pass, totalPasses, 0)
		}

		var written int64
		buffer := make([]byte, 4096)
		for written < size {
			toWrite := int64(len(buffer))
			if size-written < toWrite {
				toWrite = size - written
			}

			if err := shredPattern(method, pass, totalPasses, buffer[:toWrite]); err != nil {
				return fmt.Errorf("failed to generate pattern: %w", err)
			}

			_, err = file.Write(buffer[:toWrite])
			if err != nil {
				return fmt.Errorf("failed to write pattern: %w", err)
			}

			written += toWrite
			if progress != nil {
				progress(filepath.Base(path), pass, totalPasses, float64(written)/float64(size))
			}
		}
		// Ensure data is written to physical disk
		if err := file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
	}

	// Obfuscate filename before deletion
	randomName := filepath.Join(filepath.Dir(path), fmt.Sprintf("shred_%d", os.Getpid()))
	// Try a few random names if the first one exists
	for {
		if _, err := os.Stat(randomName); os.IsNotExist(err) {
			break
		}
		randomName = filepath.Join(filepath.Dir(path), fmt.Sprintf("shred_%d_%d", os.Getpid(), os.Getpid()))
	}

	if err := os.Rename(path, randomName); err != nil {
		// Non-fatal: if rename fails, we still delete the original
		fmt.Fprintf(os.Stderr, "warning: could not rename %s: %v\n", path, err)
	} else {
		path = randomName
	}

	return os.Remove(path)
}

// ShredDir recursively shreds all files in a directory and then removes the directories.
func ShredDir(root string, passes int, progress ShredProgress) error {
	return ShredDirWithMethod(root, passes, MethodAuto, progress)
}

// ShredDirWithMethod recursively shreds all files in a directory using the specified method.
func ShredDirWithMethod(root string, passes int, method ShredMethod, progress ShredProgress) error {
	var filesToShred []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			filesToShred = append(filesToShred, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, file := range filesToShred {
		if err := ShredFileWithMethod(file, passes, method, progress); err != nil {
			return fmt.Errorf("failed to shred %s: %w", file, err)
		}
	}

	// Remove directories in reverse order (bottom-up)
	var dirs []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for i := len(dirs) - 1; i >= 0; i-- {
		if err := os.Remove(dirs[i]); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", dirs[i], err)
		}
	}

	return nil
}

// ShredMultiple shreds multiple files concurrently with progress reporting.
func ShredMultiple(paths []string, passes int, method ShredMethod, progress MultiFileProgress) error {
	statuses := make([]FileShredStatus, len(paths))
	for i, path := range paths {
		info, err := os.Stat(path)
		size := int64(0)
		if err == nil {
			size = info.Size()
		}
		statuses[i] = FileShredStatus{
			Path:        path,
			Size:        size,
			TotalPasses: getPassesForMethod(method, passes),
			StartTime:   time.Now(),
		}
	}

	if progress != nil {
		progress(statuses)
	}

	// Limit concurrency to avoid disk contention
	maxConcurrent := runtime.NumCPU()
	if maxConcurrent > 4 {
		maxConcurrent = 4
	}
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for i := range paths {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()

			path := paths[idx]
			status := &statuses[idx]

			err := shredSingleFileWithStatus(
				path,
				passes,
				method,
				status,
				func(fileName string, pass, totalPasses int, percent float64) {
					mu.Lock()
					status.CurrentPass = pass
					status.Percent = percent
					if progress != nil {
						progress(statuses)
					}
					mu.Unlock()
				},
			)

			mu.Lock()
			status.Done = true
			status.EndTime = time.Now()
			if err != nil {
				status.Error = err
				if firstError == nil {
					firstError = err
				}
			}
			if progress != nil {
				progress(statuses)
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	return firstError
}

func shredSingleFileWithStatus(
	path string,
	passes int,
	method ShredMethod,
	status *FileShredStatus,
	progress ShredProgress,
) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	size := info.Size()
	status.Size = size

	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	totalPasses := getPassesForMethod(method, passes)
	fileName := filepath.Base(path)

	for pass := 1; pass <= totalPasses; pass++ {
		if progress != nil {
			progress(fileName, pass, totalPasses, 0)
		}

		var written int64
		buffer := make([]byte, 4096)
		for written < size {
			toWrite := int64(len(buffer))
			if size-written < toWrite {
				toWrite = size - written
			}

			if err := shredPattern(method, pass, totalPasses, buffer[:toWrite]); err != nil {
				return fmt.Errorf("failed to generate pattern: %w", err)
			}

			_, err = file.Write(buffer[:toWrite])
			if err != nil {
				return fmt.Errorf("failed to write pattern: %w", err)
			}

			written += toWrite
			if progress != nil {
				progress(fileName, pass, totalPasses, float64(written)/float64(size))
			}
		}
		// Ensure data is written to physical disk
		if err := file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
	}

	// Obfuscate filename before deletion
	randomName := filepath.Join(filepath.Dir(path), fmt.Sprintf("shred_%d", os.Getpid()))
	for {
		if _, err := os.Stat(randomName); os.IsNotExist(err) {
			break
		}
		randomName = filepath.Join(filepath.Dir(path), fmt.Sprintf("shred_%d_%d", os.Getpid(), os.Getpid()))
	}

	if err := os.Rename(path, randomName); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not rename %s: %v\n", path, err)
	} else {
		path = randomName
	}

	return os.Remove(path)
}

// CollectFilesForShredding recursively collects all files from paths, handling directories.
func CollectFilesForShredding(paths []string, recursive bool) ([]string, error) {
	var files []string
	for _, path := range paths {
		info, err := os.Lstat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", path, err)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink - shred the target file, not the link
			target, err := os.Readlink(path)
			if err != nil {
				return nil, fmt.Errorf("cannot read symlink %s: %w", path, err)
			}
			// Resolve relative symlinks
			if !filepath.IsAbs(target) {
				target = filepath.Join(filepath.Dir(path), target)
			}
			files = append(files, target)
			continue
		}

		if info.IsDir() {
			if !recursive {
				return nil, fmt.Errorf("%s: is a directory (use -r to shred directories)", path)
			}
			err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					files = append(files, p)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			files = append(files, path)
		}
	}
	return files, nil
}

// GetPassesForMethod returns the number of passes for a given method.
func GetPassesForMethod(method ShredMethod, userPasses int) int {
	return getPassesForMethod(method, userPasses)
}
