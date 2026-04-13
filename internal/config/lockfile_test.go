package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLockFile(t *testing.T) {
	lf := NewLockFile()
	if lf.Version != 1 {
		t.Errorf("version = %d, want 1", lf.Version)
	}
	if lf.Files == nil {
		t.Error("Files map is nil")
	}
}

func TestContentHash(t *testing.T) {
	h1 := ContentHash([]byte("hello"))
	h2 := ContentHash([]byte("hello"))
	h3 := ContentHash([]byte("world"))
	if h1 != h2 {
		t.Error("same content should produce same hash")
	}
	if h1 == h3 {
		t.Error("different content should produce different hash")
	}
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64", len(h1))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	lf := NewLockFile()
	lf.Track("agent/Core/identity.md", []byte("content"), "catalog:core/identity.md")
	lf.Track("CLAUDE.md", []byte("routing"), "generated:root-claude-md")

	if err := lf.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := LoadLockFile(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.Version != 1 {
		t.Errorf("version = %d, want 1", loaded.Version)
	}
	if len(loaded.Files) != 2 {
		t.Errorf("files count = %d, want 2", len(loaded.Files))
	}
	entry := loaded.Files["agent/Core/identity.md"]
	if entry == nil {
		t.Fatal("missing entry for agent/Core/identity.md")
	}
	if entry.Source != "catalog:core/identity.md" {
		t.Errorf("source = %q, want %q", entry.Source, "catalog:core/identity.md")
	}
	if entry.Hash != ContentHash([]byte("content")) {
		t.Error("hash mismatch after round-trip")
	}
}

func TestLoadLockFileMissing(t *testing.T) {
	dir := t.TempDir()
	lf, err := LoadLockFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lf.Version != 1 {
		t.Errorf("version = %d, want 1", lf.Version)
	}
	if len(lf.Files) != 0 {
		t.Errorf("expected empty files map, got %d entries", len(lf.Files))
	}
}

func TestIsModifiedFileNotExist(t *testing.T) {
	dir := t.TempDir()
	lf := NewLockFile()
	exists, modified := lf.IsModified(dir, "nonexistent.md")
	if exists || modified {
		t.Error("nonexistent file should be !exists, !modified")
	}
}

func TestIsModifiedUntracked(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "mystery.md"), []byte("user content"), 0644)
	lf := NewLockFile()
	exists, modified := lf.IsModified(dir, "mystery.md")
	if !exists {
		t.Error("file should exist")
	}
	if !modified {
		t.Error("untracked existing file should be treated as modified")
	}
}

func TestIsModifiedUnchanged(t *testing.T) {
	dir := t.TempDir()
	content := []byte("original content")
	_ = os.WriteFile(filepath.Join(dir, "file.md"), content, 0644)
	lf := NewLockFile()
	lf.Track("file.md", content, "test")
	exists, modified := lf.IsModified(dir, "file.md")
	if !exists {
		t.Error("file should exist")
	}
	if modified {
		t.Error("unchanged file should not be modified")
	}
}

func TestIsModifiedChanged(t *testing.T) {
	dir := t.TempDir()
	original := []byte("original")
	_ = os.WriteFile(filepath.Join(dir, "file.md"), []byte("edited by user"), 0644)
	lf := NewLockFile()
	lf.Track("file.md", original, "test")
	exists, modified := lf.IsModified(dir, "file.md")
	if !exists {
		t.Error("file should exist")
	}
	if !modified {
		t.Error("changed file should be modified")
	}
}

func TestUntrack(t *testing.T) {
	lf := NewLockFile()
	lf.Track("a.md", []byte("a"), "test")
	lf.Track("b.md", []byte("b"), "test")
	lf.Untrack("a.md")
	if _, ok := lf.Files["a.md"]; ok {
		t.Error("a.md should be untracked")
	}
	if _, ok := lf.Files["b.md"]; !ok {
		t.Error("b.md should still be tracked")
	}
}
