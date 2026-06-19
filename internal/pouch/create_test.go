package pouch

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAuto(t *testing.T) {
	t.Parallel()

	t.Run("creates directory in auto mode", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample")

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Kind != KindDir || result.Action != ActionCreateDir {
			t.Fatalf("unexpected result: %+v", result)
		}
		assertDirExists(t, path)
	})

	t.Run("creates file in auto mode", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample.ts")

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Kind != KindFile || result.Action != ActionCreateFile {
			t.Fatalf("unexpected result: %+v", result)
		}
		assertFileExists(t, path)
	})

	t.Run("creates parent directory for file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample", "temp.ts")

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Action != ActionCreateFile {
			t.Fatalf("unexpected action: %+v", result)
		}
		assertDirExists(t, filepath.Dir(path))
		assertFileExists(t, path)
	})

	t.Run("trailing slash creates directory in auto mode", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "dir.with.dot") + string(filepath.Separator)

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Kind != KindDir || result.Action != ActionCreateDir {
			t.Fatalf("unexpected result: %+v", result)
		}
		assertDirExists(t, filepath.Clean(path))
	})

	t.Run("skips existing file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample.ts")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Action != ActionSkipExisting || result.Kind != KindFile {
			t.Fatalf("unexpected result: %+v", result)
		}
	})

	t.Run("skips existing directory", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Action != ActionSkipExisting || result.Kind != KindDir {
			t.Fatalf("unexpected result: %+v", result)
		}
	})

	t.Run("strict errors for existing file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample.ts")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := Create(path, Options{Strict: true})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("strict errors for existing directory", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}

		_, err := Create(path, Options{Strict: true})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCreateModeOverrides(t *testing.T) {
	t.Parallel()

	t.Run("mode file overrides Dockerfile", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "Dockerfile")

		result, err := Create(path, Options{Mode: ModeFile})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Kind != KindFile || result.Action != ActionCreateFile {
			t.Fatalf("unexpected result: %+v", result)
		}
		assertFileExists(t, path)
	})

	t.Run("mode dir overrides dir.with.dot", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "dir.with.dot")

		result, err := Create(path, Options{Mode: ModeDir})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Kind != KindDir || result.Action != ActionCreateDir {
			t.Fatalf("unexpected result: %+v", result)
		}
		assertDirExists(t, path)
	})
}

func TestCreateDryRun(t *testing.T) {
	t.Parallel()

	t.Run("dry run reports intended file creation", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample", "temp.ts")

		result, err := Create(path, Options{DryRun: true})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Action != ActionCreateFile || result.Kind != KindFile {
			t.Fatalf("unexpected result: %+v", result)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected file not to exist, got err=%v", err)
		}
	})

	t.Run("dry run reports intended directory creation", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "sample")

		result, err := Create(path, Options{DryRun: true})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if result.Action != ActionCreateDir || result.Kind != KindDir {
			t.Fatalf("unexpected result: %+v", result)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected directory not to exist, got err=%v", err)
		}
	})
}

func TestCreateErrors(t *testing.T) {
	t.Parallel()

	t.Run("file mode errors for existing directory", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "target")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}

		_, err := Create(path, Options{Mode: ModeFile})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("dir mode errors for existing file", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "target")
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := Create(path, Options{Mode: ModeDir})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("file mode rejects trailing slash", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		path := filepath.Join(root, "target") + string(filepath.Separator)

		_, err := Create(path, Options{Mode: ModeFile})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCreateMany(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	paths := []string{
		filepath.Join(root, "a"),
		filepath.Join(root, "b.ts"),
	}

	results, err := CreateMany(paths, Options{})
	if err != nil {
		t.Fatalf("CreateMany() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Action != ActionCreateDir || results[1].Action != ActionCreateFile {
		t.Fatalf("unexpected results: %+v", results)
	}

	t.Run("returns partial results before failure", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		conflict := filepath.Join(root, "conflict")
		if err := os.Mkdir(conflict, 0o755); err != nil {
			t.Fatal(err)
		}

		paths := []string{
			filepath.Join(root, "ok.ts"),
			conflict,
			filepath.Join(root, "later"),
		}
		results, err := CreateMany(paths, Options{Mode: ModeFile})
		if err == nil {
			t.Fatal("expected error")
		}
		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Action != ActionCreateFile {
			t.Fatalf("unexpected partial result: %+v", results[0])
		}
	})
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%q is directory, want file", path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is file, want directory", path)
	}
}
