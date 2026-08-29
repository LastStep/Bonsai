package cmd

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatBackupHint(t *testing.T) {
	tests := []struct {
		name      string
		paths     []string
		wantParts []string
		notParts  []string
	}{
		{
			name: "zero",
		},
		{
			name:      "one",
			paths:     []string{"station/agent/Skills/design-guide.md.bak"},
			wantParts: []string{"1 file was backed up as .bak:", "station/agent/Skills/design-guide.md.bak", "Ask your agent to reconcile the backup", "into station/agent/Skills/design-guide.md"},
		},
		{
			name:      "multiple and sorted",
			paths:     []string{"b.md.bak", "a.md.bak"},
			wantParts: []string{"2 files were backed up as .bak:", "  a.md.bak\n  b.md.bak", "Ask your agent to reconcile the backups"},
		},
		{
			name:      "exactly ten",
			paths:     backupHintPaths(10),
			wantParts: []string{"10 files were backed up as .bak:", "file-09.md.bak"},
			notParts:  []string{"... and"},
		},
		{
			name:      "truncates after ten",
			paths:     backupHintPaths(13),
			wantParts: []string{"13 files were backed up as .bak:", "file-09.md.bak", "... and 3 more"},
			notParts:  []string{"file-10.md.bak", "file-12.md.bak"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBackupHint(tt.paths)
			if len(tt.wantParts) == 0 {
				if got != "" {
					t.Fatalf("formatBackupHint() = %q, want empty", got)
				}
				return
			}
			for _, want := range tt.wantParts {
				if !strings.Contains(got, want) {
					t.Errorf("formatBackupHint() missing %q in %q", want, got)
				}
			}
			for _, unwanted := range tt.notParts {
				if strings.Contains(got, unwanted) {
					t.Errorf("formatBackupHint() unexpectedly contains %q in %q", unwanted, got)
				}
			}
		})
	}
}

func backupHintPaths(count int) []string {
	paths := make([]string, count)
	for i := range paths {
		paths[i] = fmt.Sprintf("file-%02d.md.bak", i)
	}
	return paths
}
