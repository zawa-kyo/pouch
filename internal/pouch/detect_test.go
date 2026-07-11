package pouch

import "testing"

func TestDetectClassifiesPathsByDocumentedAutoModeRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want Kind
	}{
		{
			name: "bare final segment is a directory",
			path: "sample",
			want: KindDir,
		},
		{
			name: "dotted final segment is a file",
			path: "sample.ts",
			want: KindFile,
		},
		{
			name: "only the final segment controls file detection",
			path: "sample/temp.ts",
			want: KindFile,
		},
		{
			name: "dotfile is a file",
			path: ".env",
			want: KindFile,
		},
		{
			name: "trailing slash is an explicit directory hint",
			path: "dir.with.dot/",
			want: KindDir,
		},
		{
			name: "well-known extensionless filename remains a directory in auto mode",
			path: "Dockerfile",
			want: KindDir,
		},
		{
			name: "dotted directory name needs an explicit override without trailing slash",
			path: "dir.with.dot",
			want: KindFile,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Detect(tt.path); got != tt.want {
				t.Fatalf("Detect(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
