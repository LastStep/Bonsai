package main

// The generator's arguments: only a folder that is there is taken as the repository's root. A flag (-h, --help) or a
// path that is not a folder is refused with the usage, and nothing is written (before, `go run ./gen -h` wrote every
// generated file under a folder named -h).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnlyAFolderIsTheRoot(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	file := filepath.Join(dir, "a-file")
	if err := os.WriteFile(file, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-h"}, {"--help"}, {"-"}, {}, {"a", "b"}, {""}, {file}, {filepath.Join(dir, "not-there")}} {
		var stderr bytes.Buffer
		if code := run(args, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage: go run ./gen") {
			t.Errorf("%q: exit %d, stderr %q; want 2 with the usage", args, code, stderr.String())
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "a-file" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("a refused run wrote in the working folder: %q", names)
	}
	if b, _ := os.ReadFile(file); string(b) != "x\n" {
		t.Errorf("the file given as the root was changed: %q", b)
	}
}

func TestAFolderIsWritten(t *testing.T) {
	root := t.TempDir()
	var stderr bytes.Buffer
	if code := run([]string{root}, &stderr); code != 0 {
		t.Fatalf("exit %d, stderr %q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(root, "docs", "reference", "lists.md")); err != nil {
		t.Errorf("the reference page was not written: %v", err)
	}
}
