package pouch

import "os"

type Mode int

const (
	ModeAuto Mode = iota
	ModeFile
	ModeDir
)

type Kind int

const (
	KindFile Kind = iota
	KindDir
)

type Action int

const (
	ActionNone Action = iota
	ActionCreateFile
	ActionCreateDir
	ActionSkipExisting
)

type Options struct {
	Mode     Mode
	DirPerm  os.FileMode
	FilePerm os.FileMode
	DryRun   bool
}

type Result struct {
	Path   string
	Kind   Kind
	Action Action
}

const (
	DefaultDirPerm  os.FileMode = 0o755
	DefaultFilePerm os.FileMode = 0o644
)
