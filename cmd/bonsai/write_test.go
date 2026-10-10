package main

// bonsai check --write from the command line (step 5.1.8): the exit code says only whether it wrote, findings are
// still listed, a stale table is a warning only and never a status problem, and a worktree is refused (exit 4).

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

func TestCheckWrite(t *testing.T) {
	fakeClaude(t, "2.1.294 (Claude Code)", nil)
	c := newCLI(t)
	c.link()
	c.commit("link")
	// Every test project has one finding (its pack source is a local path): plain check exits 1, --write 0.
	c.task("T-0901", "running", "", "")
	code, out, _ := c.run("", "check", "--json")
	doc := fits(t, out, "check")
	if code != 1 || codesIn(doc, "findings") != "absolute-path" || codesIn(doc, "warnings") != "tables" {
		t.Fatalf("a stale table: %d\n%s", code, out)
	}
	code, out, _ = c.run("", "check", "--write")
	if code != 0 || !strings.Contains(out, "bonsai check: 1 finding.\n") || !strings.Contains(out, "note: wrote .bonsai/tasks.md\n") ||
		strings.Contains(out, "warning:") {
		t.Errorf("check --write: %d\n%s", code, out)
	}
	table := c.read(".bonsai/tasks.md")
	if !strings.Contains(table, "Active task when none is named: T-0901\n") || !strings.Contains(table, "| T-0901 | A task | running | | | |\n") {
		t.Errorf("the table:\n%s", table)
	}
	sum := sha256.Sum256([]byte(table))
	for i := 0; i < 2; i++ {
		code, out, _ = c.run("", "check", "--write", "--json")
		doc = fits(t, out, "check")
		if code != 0 || codesIn(doc, "findings") != "absolute-path" || codesIn(doc, "warnings") != "" || sha256.Sum256([]byte(c.read(".bonsai/tasks.md"))) != sum {
			t.Errorf("check --write --json, run %d: %d\n%s", i, code, out)
		}
	}
	if code, out, _ = c.run("", "check", "--write"); code != 0 || !strings.Contains(out, "note: .bonsai/tasks.md is already as a rebuild gives it\n") {
		t.Errorf("a second check --write: %d\n%s", code, out)
	}
	// A stale table is never a status problem.
	c.task("T-0902", "todo", "", "")
	_, out, _ = c.run("", "status", "--json")
	problems, _ := fits(t, out, "status").Get("problems")
	for _, p := range problems.([]any) {
		if strings.Contains(p.(string), "tasks.md") || strings.Contains(p.(string), "check --write") {
			t.Errorf("status lists the stale table as a problem: %s", p)
		}
	}
	if _, out, _ := c.run("", "check", "--json"); codesIn(fits(t, out, "check"), "warnings") != "tables" {
		t.Errorf("not stale:\n%s", out)
	}

	// Could not write: the table's place is a folder. Findings are listed, the error object is filled, exit 3.
	c.commit("tasks")
	if err := os.Remove(filepath.Join(c.root, ".bonsai", "tasks.md")); err != nil {
		t.Fatal(err)
	}
	c.write(".bonsai/tasks.md/keep", "x")
	code, out, errOut := c.run("", "check", "--write")
	if code != 3 || !strings.Contains(out, "bonsai check: 1 finding.\n") || !strings.Contains(errOut, "bonsai check: .bonsai/tasks.md cannot be written") ||
		!strings.Contains(errOut, "\nnext: ") {
		t.Errorf("a write that fails: %d\n%s%s", code, out, errOut)
	}
	code, out, _ = c.run("", "check", "--write", "--json")
	doc = fits(t, out, "check")
	if e := errorIn(doc, "check"); code != 3 || e.String("code") != "write-failed" || codesIn(doc, "findings") != "absolute-path" {
		t.Errorf("a write that fails, --json: %d\n%s", code, out)
	}
}

func TestCheckWriteInAWorktree(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	wt := filepath.Join(c.tmp, "branch")
	testpack.Git(t, c.root, "worktree", "add", "-q", "-b", "branch", wt)
	t.Chdir(wt)
	before, _ := os.ReadFile(filepath.Join(wt, ".bonsai", "tasks.md"))
	code, out, errOut := c.run("", "check", "--write")
	if code != 4 || out != "" || !strings.Contains(errOut, "this is a worktree of "+filepath.ToSlash(c.root)) ||
		!strings.Contains(errOut, "next: run it in the main checkout, "+filepath.ToSlash(c.root)+": bonsai check --write\n") {
		t.Errorf("check --write in a worktree: %d\n%s%s", code, out, errOut)
	}
	code, out, _ = c.run("", "check", "--write", "--json")
	doc := fits(t, out, "check")
	if e := errorIn(doc, "check"); code != 4 || e.String("code") != "not-main-checkout" || whoOf(e) != "agent" {
		t.Errorf("check --write --json in a worktree: %d\n%s", code, out)
	}
	if after, _ := os.ReadFile(filepath.Join(wt, ".bonsai", "tasks.md")); string(after) != string(before) {
		t.Errorf("the worktree's table changed")
	}
	// Plain check in the worktree runs, and the table is current there.
	if code, out, _ := c.run("", "check", "--json"); code != 1 || strings.Contains(codesIn(fits(t, out, "check"), "warnings"), "tables") {
		t.Errorf("check in a worktree: %d\n%s", code, out)
	}
}
