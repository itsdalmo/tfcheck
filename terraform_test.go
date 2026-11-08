package tfcheck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindBinary(t *testing.T) {
	tests := []struct {
		name     string
		binaries []string
		want     string
		wantErr  bool
	}{
		{name: "terraform", binaries: []string{"terraform"}, want: "terraform"},
		{name: "tofu", binaries: []string{"tofu"}, want: "tofu"},
		{name: "both prefers terraform", binaries: []string{"terraform", "tofu"}, want: "terraform"},
		{name: "neither", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, b := range tt.binaries {
				if err := os.WriteFile(filepath.Join(dir, b), []byte("#!/bin/sh\n"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", dir)

			got, err := FindBinary()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
