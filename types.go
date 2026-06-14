package pouch

import "os"

// Mode controls how a path is interpreted.
type Mode int

const (
	// ModeAuto detects the target kind from the final path segment.
	ModeAuto Mode = iota
	// ModeFile forces file creation semantics.
	ModeFile
	// ModeDir forces directory creation semantics.
	ModeDir
)

// Kind is the detected or requested target type.
type Kind int

const (
	// KindFile represents a file path.
	KindFile Kind = iota
	// KindDir represents a directory path.
	KindDir
)

// Action describes what Create or CreateMany did or would do.
type Action int

const (
	// ActionNone means no action was recorded.
	ActionNone Action = iota
	// ActionCreateFile means a file was created or would be created.
	ActionCreateFile
	// ActionCreateDir means a directory was created or would be created.
	ActionCreateDir
	// ActionSkipExisting means the target already existed with the expected kind.
	ActionSkipExisting
)

// Options configures path creation behavior.
type Options struct {
	Mode     Mode
	DirPerm  os.FileMode
	FilePerm os.FileMode
	DryRun   bool
}

// Result reports the observed or planned outcome for one path.
type Result struct {
	Path   string
	Kind   Kind
	Action Action
}

const (
	// DefaultDirPerm is the default permission used for new directories.
	DefaultDirPerm os.FileMode = 0o755
	// DefaultFilePerm is the default permission used for new files.
	DefaultFilePerm os.FileMode = 0o644
)
