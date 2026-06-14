package pouch

// Create creates a file or directory for a single path.
func Create(path string, opts Options) (Result, error) {
	return create(path, withDefaults(opts))
}

// CreateMany processes paths in order and stops at the first error.
func CreateMany(paths []string, opts Options) ([]Result, error) {
	results := make([]Result, 0, len(paths))
	opts = withDefaults(opts)
	for _, path := range paths {
		result, err := create(path, opts)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func withDefaults(opts Options) Options {
	if opts.DirPerm == 0 {
		opts.DirPerm = DefaultDirPerm
	}
	if opts.FilePerm == 0 {
		opts.FilePerm = DefaultFilePerm
	}
	return opts
}
