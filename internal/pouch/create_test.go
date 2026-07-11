package pouch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCreatesMissingTargetsFromPathSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		path         func(string) string
		options      Options
		wantKind     Kind
		wantAction   Action
		assertTarget func(*testing.T, string)
	}{
		{
			name:         "auto mode creates a directory for an extensionless final segment",
			path:         func(root string) string { return filepath.Join(root, "sample") },
			wantKind:     KindDir,
			wantAction:   ActionCreateDir,
			assertTarget: assertDir,
		},
		{
			name:         "auto mode creates a file and its parent directories for a dotted final segment",
			path:         func(root string) string { return filepath.Join(root, "src", "main.go") },
			wantKind:     KindFile,
			wantAction:   ActionCreateFile,
			assertTarget: assertFile,
		},
		{
			name: "auto mode treats a trailing separator as a directory hint",
			path: func(root string) string {
				return filepath.Join(root, "dir.with.dot") + string(filepath.Separator)
			},
			wantKind:     KindDir,
			wantAction:   ActionCreateDir,
			assertTarget: assertDir,
		},
		{
			name:         "file mode creates an extensionless path as a file",
			path:         func(root string) string { return filepath.Join(root, "Dockerfile") },
			options:      Options{Mode: ModeFile},
			wantKind:     KindFile,
			wantAction:   ActionCreateFile,
			assertTarget: assertFile,
		},
		{
			name:         "directory mode creates a dotted path as a directory",
			path:         func(root string) string { return filepath.Join(root, "dir.with.dot") },
			options:      Options{Mode: ModeDir},
			wantKind:     KindDir,
			wantAction:   ActionCreateDir,
			assertTarget: assertDir,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.path(t.TempDir())

			result, err := Create(path, tt.options)
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			assertResult(t, result, path, tt.wantKind, tt.wantAction)
			tt.assertTarget(t, filepath.Clean(path))
		})
	}
}

func TestCreateLeavesExistingTargetsUnchangedByDefault(t *testing.T) {
	t.Parallel()

	t.Run("existing file is skipped without changing its content", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "sample.go")
		writeFile(t, path, "original")

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		assertResult(t, result, path, KindFile, ActionSkipExisting)
		assertFileContent(t, path, "original")
	})

	t.Run("existing directory is skipped", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "sample")
		mkdir(t, path)

		result, err := Create(path, Options{})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		assertResult(t, result, path, KindDir, ActionSkipExisting)
		assertDir(t, path)
	})
}

func TestCreateDryRunReportsPlannedWorkWithoutWriting(t *testing.T) {
	t.Parallel()

	t.Run("file target does not create the file or its parent directory", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "src", "main.go")

		result, err := Create(path, Options{DryRun: true})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		assertResult(t, result, path, KindFile, ActionCreateFile)
		assertMissing(t, filepath.Dir(path))
		assertMissing(t, path)
	})

	t.Run("directory target is reported but not created", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "sample")

		result, err := Create(path, Options{DryRun: true})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		assertResult(t, result, path, KindDir, ActionCreateDir)
		assertMissing(t, path)
	})
}

func TestCreateRejectsConflictingTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        func(string) string
		prepare     func(*testing.T, string)
		options     Options
		wantErr     string
		assertAfter func(*testing.T, string)
	}{
		{
			name:        "strict mode rejects an existing file without changing its content",
			path:        func(root string) string { return filepath.Join(root, "sample.go") },
			prepare:     func(t *testing.T, path string) { writeFile(t, path, "original") },
			options:     Options{Strict: true},
			wantErr:     "path already exists",
			assertAfter: func(t *testing.T, path string) { assertFileContent(t, path, "original") },
		},
		{
			name:        "strict mode rejects an existing directory without replacing it",
			path:        func(root string) string { return filepath.Join(root, "sample") },
			prepare:     func(t *testing.T, path string) { mkdir(t, path) },
			options:     Options{Strict: true},
			wantErr:     "path already exists",
			assertAfter: assertDir,
		},
		{
			name:        "file mode rejects an existing directory without replacing it",
			path:        func(root string) string { return filepath.Join(root, "target") },
			prepare:     func(t *testing.T, path string) { mkdir(t, path) },
			options:     Options{Mode: ModeFile},
			wantErr:     "path exists as directory",
			assertAfter: assertDir,
		},
		{
			name:        "directory mode rejects an existing file without changing its content",
			path:        func(root string) string { return filepath.Join(root, "target") },
			prepare:     func(t *testing.T, path string) { writeFile(t, path, "original") },
			options:     Options{Mode: ModeDir},
			wantErr:     "path exists as file",
			assertAfter: func(t *testing.T, path string) { assertFileContent(t, path, "original") },
		},
		{
			name: "file mode rejects a trailing separator before writing anything",
			path: func(root string) string {
				return filepath.Join(root, "target") + string(filepath.Separator)
			},
			options:     Options{Mode: ModeFile},
			wantErr:     "trailing slash is only valid for directories",
			assertAfter: func(t *testing.T, path string) { assertMissing(t, filepath.Clean(path)) },
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.path(t.TempDir())
			if tt.prepare != nil {
				tt.prepare(t, filepath.Clean(path))
			}

			_, err := Create(path, tt.options)

			assertErrorContains(t, err, tt.wantErr)
			tt.assertAfter(t, path)
		})
	}
}

func TestCreateManyProcessesPathsInInputOrder(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	paths := []string{
		filepath.Join(root, "first"),
		filepath.Join(root, "second.go"),
		filepath.Join(root, "third"),
	}

	results, err := CreateMany(paths, Options{})
	if err != nil {
		t.Fatalf("CreateMany() error = %v", err)
	}

	assertResults(t, results, []Result{
		{Path: paths[0], Kind: KindDir, Action: ActionCreateDir},
		{Path: paths[1], Kind: KindFile, Action: ActionCreateFile},
		{Path: paths[2], Kind: KindDir, Action: ActionCreateDir},
	})
	assertDir(t, paths[0])
	assertFile(t, paths[1])
	assertDir(t, paths[2])
}

func TestCreateManyStopsAtFirstErrorAndReturnsCompletedResults(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "first.go")
	conflict := filepath.Join(root, "conflict")
	later := filepath.Join(root, "later.go")
	mkdir(t, conflict)

	results, err := CreateMany([]string{first, conflict, later}, Options{Mode: ModeFile})

	assertErrorContains(t, err, "path exists as directory")
	assertResults(t, results, []Result{
		{Path: first, Kind: KindFile, Action: ActionCreateFile},
	})
	assertFile(t, first)
	assertMissing(t, later)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), DefaultDirPerm); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), DefaultFilePerm); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, DefaultDirPerm); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", path, err)
	}
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%q is a directory, want file", path)
	}
}

func assertDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is a file, want directory", path)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("os.Stat(%q) error = %v, want missing path", path, err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("file content = %q, want %q", string(got), want)
	}
}

func assertResult(t *testing.T, got Result, wantPath string, wantKind Kind, wantAction Action) {
	t.Helper()
	assertResults(t, []Result{got}, []Result{{Path: wantPath, Kind: wantKind, Action: wantAction}})
}

func assertResults(t *testing.T, got, want []Result) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(results) = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("results[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want message containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want message containing %q", err.Error(), want)
	}
}
