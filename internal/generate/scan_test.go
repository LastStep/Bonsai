package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LastStep/Bonsai/internal/config"
)

func TestScanCustomFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up workspace structure
	skillsDir := filepath.Join(tmpDir, "ws", "agent", "Skills")
	_ = os.MkdirAll(skillsDir, 0755)

	// Create a tracked skill (simulating a catalog item already in config)
	_ = os.WriteFile(filepath.Join(skillsDir, "planning-template.md"), []byte("catalog skill"), 0644)

	// Create an untracked custom skill with valid frontmatter
	_ = os.WriteFile(filepath.Join(skillsDir, "my-custom.md"), []byte(`---
description: My custom skill
display_name: My Custom
---

# My Custom Skill
`), 0644)

	// Create an untracked custom skill WITHOUT frontmatter
	_ = os.WriteFile(filepath.Join(skillsDir, "no-frontmatter.md"), []byte(`# No Frontmatter
Just content.
`), 0644)

	// Create a subdirectory (should be skipped)
	subDir := filepath.Join(skillsDir, "bubbletea")
	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "ref.md"), []byte("reference file"), 0644)

	installed := &config.InstalledAgent{
		AgentType: "tech-lead",
		Workspace: "ws/",
		Skills:    []string{"planning-template"},
	}

	lock := config.NewLockFile()
	// Track the catalog skill in lock
	lock.Track("ws/agent/Skills/planning-template.md", []byte("catalog skill"), "catalog:skills/planning-template")

	discovered, err := ScanCustomFiles(tmpDir, installed, lock)
	if err != nil {
		t.Fatalf("ScanCustomFiles: %v", err)
	}

	if len(discovered) != 2 {
		t.Fatalf("expected 2 discovered files, got %d", len(discovered))
	}

	// Check the valid custom skill
	var validFile, invalidFile *DiscoveredFile
	for i := range discovered {
		if discovered[i].Name == "my-custom" {
			validFile = &discovered[i]
		}
		if discovered[i].Name == "no-frontmatter" {
			invalidFile = &discovered[i]
		}
	}

	if validFile == nil {
		t.Fatal("my-custom not found in discovered files")
	}
	if validFile.Type != "skill" {
		t.Errorf("my-custom type = %q, want skill", validFile.Type)
	}
	if validFile.Error != "" {
		t.Errorf("my-custom should have no error, got %q", validFile.Error)
	}
	if validFile.Meta.Description != "My custom skill" {
		t.Errorf("my-custom description = %q", validFile.Meta.Description)
	}

	if invalidFile == nil {
		t.Fatal("no-frontmatter not found in discovered files")
	}
	if invalidFile.Error == "" {
		t.Error("no-frontmatter should have an error")
	}
}

func TestScanCustomSensorMissingEvent(t *testing.T) {
	tmpDir := t.TempDir()

	sensorsDir := filepath.Join(tmpDir, "ws", "agent", "Sensors")
	_ = os.MkdirAll(sensorsDir, 0755)

	// Custom sensor WITHOUT event field
	_ = os.WriteFile(filepath.Join(sensorsDir, "bad-sensor.sh"), []byte(`---
description: A sensor without an event
---

#!/usr/bin/env bash
echo "hello"
`), 0644)

	installed := &config.InstalledAgent{
		AgentType: "tech-lead",
		Workspace: "ws/",
	}

	lock := config.NewLockFile()

	discovered, err := ScanCustomFiles(tmpDir, installed, lock)
	if err != nil {
		t.Fatalf("ScanCustomFiles: %v", err)
	}

	if len(discovered) != 1 {
		t.Fatalf("expected 1 discovered file, got %d", len(discovered))
	}

	if discovered[0].Error == "" {
		t.Error("sensor without event should have a validation error")
	}
}
