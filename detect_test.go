package pouch

import "testing"

func TestDetect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want Kind
	}{
		{path: "sample", want: KindDir},
		{path: "sample.ts", want: KindFile},
		{path: "sample/temp.ts", want: KindFile},
		{path: ".env", want: KindFile},
		{path: "Dockerfile", want: KindDir},
		{path: "dir.with.dot", want: KindFile},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			if got := Detect(tt.path); got != tt.want {
				t.Fatalf("Detect(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
