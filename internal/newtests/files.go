package newtests

// The branch's changed test files (note 4): by the tests form's convention (format.TestFileConventions, read through
// vergit.TestFile, the one reading git integrity's test-file-changed finding makes too), or for junit the files the
// reports' testcase elements name; changed when git's diff of the base against the climbed working tree, or its
// untracked files that git status would show, name them.

import (
	"runtime"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/testout"
	"github.com/LastStep/Bonsai/internal/vergit"
)

// change is one test file the branch changed: its project-relative path with forward slashes, and whether the branch
// deleted it.
type change struct {
	path    string
	deleted bool
}

// changedFiles lists the branch's changed test files, in git's order (the diff's, then the untracked files'). why is
// set when none can be found, so the test files cannot be told. first is the base's own report, which for junit names
// the test files the branch deleted; nil before that run.
func (c *check) changedFiles(first *testout.Report) (files []change, why string, err error) {
	isTest := func(p string) bool { return vergit.TestFile([]string{c.form}, p) }
	if c.form == format.TestsJUnit {
		named := map[string]bool{}
		for _, rep := range []*testout.Report{&c.in.Head, first} {
			if rep == nil {
				continue
			}
			for _, t := range rep.Tests {
				if p := cleanRel(t.File); p != "" {
					named[fileKey(p)] = true
				}
			}
		}
		if len(named) == 0 {
			return nil, "the JUnit report names no file for its tests (no file attribute on its testcase elements)", nil
		}
		isTest = func(p string) bool { return named[fileKey(p)] }
	}
	base := c.res.NewTests.Base
	out, err := c.git.Run("diff", "--name-status", "--no-renames", "--no-ext-diff", "--no-textconv",
		"--ignore-submodules=all", "--no-relative", "-z", base, "--")
	if err != nil {
		return nil, "", err
	}
	d := splitZ(out)
	for i := 0; i+1 < len(d); i += 2 {
		p := cleanRel(d[i+1])
		if p == "" || !isTest(p) {
			continue
		}
		files = append(files, change{path: p, deleted: strings.HasPrefix(d[i], "D")})
	}
	out, err = c.git.Run("ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return nil, "", err
	}
	for _, p := range splitZ(out) {
		if p = cleanRel(p); p != "" && isTest(p) {
			files = append(files, change{path: p})
		}
	}
	if len(files) == 0 {
		if c.form == format.TestsJUnit {
			return nil, "the branch changed no file the JUnit report names", nil
		}
		return nil, "the branch changed no test file by the " + c.form + " convention (bonsai check --schema bonsai.workspace)", nil
	}
	return files, "", nil
}

// fileKey is a project-relative path as a lookup key: without case on Windows, as its file system compares.
func fileKey(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(p)
	}
	return p
}

// splitZ splits NUL-terminated output.
func splitZ(out []byte) []string {
	s := strings.TrimSuffix(string(out), "\x00")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\x00")
}
