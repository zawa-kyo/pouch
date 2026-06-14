package pouch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func create(path string, opts Options) (Result, error) {
	kind := detectWithMode(path, opts.Mode)
	result := Result{Path: path, Kind: kind}

	info, err := os.Lstat(path)
	if err == nil {
		return handleExisting(path, info, kind, result)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Result{}, fmt.Errorf("inspect %q: %w", path, err)
	}

	switch kind {
	case KindFile:
		return createFile(path, opts, result)
	case KindDir:
		return createDir(path, opts, result)
	default:
		return Result{}, fmt.Errorf("detect %q: unknown kind", path)
	}
}

func detectWithMode(path string, mode Mode) Kind {
	switch mode {
	case ModeFile:
		return KindFile
	case ModeDir:
		return KindDir
	default:
		return Detect(path)
	}
}

func handleExisting(path string, info os.FileInfo, kind Kind, result Result) (Result, error) {
	switch kind {
	case KindFile:
		if info.IsDir() {
			return Result{}, fmt.Errorf("create file %q: path exists as directory", path)
		}
	case KindDir:
		if !info.IsDir() {
			return Result{}, fmt.Errorf("create directory %q: path exists as file", path)
		}
	}
	result.Action = ActionSkipExisting
	return result, nil
}

func createFile(path string, opts Options, result Result) (Result, error) {
	parent := filepath.Dir(path)
	if parent != "." && parent != path {
		if opts.DryRun {
			result.Action = ActionCreateFile
			return result, nil
		}
		if err := os.MkdirAll(parent, opts.DirPerm); err != nil {
			return Result{}, fmt.Errorf("create parent directory for %q: %w", path, err)
		}
	}

	if opts.DryRun {
		result.Action = ActionCreateFile
		return result, nil
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, opts.FilePerm)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			info, statErr := os.Lstat(path)
			if statErr != nil {
				return Result{}, fmt.Errorf("inspect %q after create race: %w", path, statErr)
			}
			return handleExisting(path, info, result.Kind, result)
		}
		return Result{}, fmt.Errorf("create file %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return Result{}, fmt.Errorf("close file %q: %w", path, err)
	}

	result.Action = ActionCreateFile
	return result, nil
}

func createDir(path string, opts Options, result Result) (Result, error) {
	if opts.DryRun {
		result.Action = ActionCreateDir
		return result, nil
	}
	if err := os.MkdirAll(path, opts.DirPerm); err != nil {
		return Result{}, fmt.Errorf("create directory %q: %w", path, err)
	}
	result.Action = ActionCreateDir
	return result, nil
}
