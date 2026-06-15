package nonint

import (
	"bytes"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/generate"
)

// goldenVersion is the FIXED catalog version string the golden fixtures were
// captured with (testdata/{init,add}_golden.jsonl). The byte-identity test
// must re-run RunInit/RunAdd with this exact value so the
// WriteCatalogSnapshot line reproduces byte-for-byte.
const goldenVersion = "v0.0.0-test"

// TestByteIdentity_Init is the back-compat oracle (Plan 41 Verification B1):
// EmitJSONL over the *Result that RunInit produces for the pinned init input
// fixture must equal testdata/init_golden.jsonl byte-for-byte. The golden was
// captured from main PRE-refactor — if this diff is non-empty the Result
// reshape altered the JSONL stream and the refactor is WRONG.
func TestByteIdentity_Init(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	input := readFixture(t, "testdata/init_input.yaml")
	cfgPath := writeYAML(t, tmp, "init_input.yaml", input)
	cfg, err := LoadConfig(cfgPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	result, code, err := RunInit(tmp, filepath.Join(tmp, ".bonsai.yaml"), cfg, cat, goldenVersion)
	if err != nil || code != ExitOK {
		t.Fatalf("RunInit: code=%d err=%v", code, err)
	}

	var buf bytes.Buffer
	if err := EmitJSONL(&buf, result); err != nil {
		t.Fatalf("EmitJSONL: %v", err)
	}
	assertGolden(t, "testdata/init_golden.jsonl", buf.Bytes())
}

// TestByteIdentity_Add is the add-path back-compat oracle (B1). RunInit first
// (discarded) to materialise the project, then RunAdd with the backend
// overlay fixture; EmitJSONL must equal testdata/add_golden.jsonl.
func TestByteIdentity_Add(t *testing.T) {
	cat := loadTestCatalog(t)
	tmp := t.TempDir()

	initCfg := minimalInitCfg(t, tmp)
	configPath := filepath.Join(tmp, ".bonsai.yaml")
	if _, code, err := RunInit(tmp, configPath, initCfg, cat, goldenVersion); err != nil || code != ExitOK {
		t.Fatalf("RunInit (setup): code=%d err=%v", code, err)
	}

	input := readFixture(t, "testdata/add_input.yaml")
	overlayPath := writeYAML(t, tmp, "add_input.yaml", input)
	overlay, err := LoadConfig(overlayPath, tmp, cat)
	if err != nil {
		t.Fatalf("LoadConfig overlay: %v", err)
	}

	result, code, err := RunAdd(tmp, configPath, overlay, cat, goldenVersion)
	if err != nil || code != ExitOK {
		t.Fatalf("RunAdd: code=%d err=%v", code, err)
	}

	var buf bytes.Buffer
	if err := EmitJSONL(&buf, result); err != nil {
		t.Fatalf("EmitJSONL: %v", err)
	}
	assertGolden(t, "testdata/add_golden.jsonl", buf.Bytes())
}

// TestEmitJSONL_NoWarningOnStdout is the targeted regression guard for the
// dropped `warning` event + deleted EmitWarning: a Result with non-empty
// Warnings must emit ZERO warning-event lines and ZERO occurrences of the
// warning text to the JSONL writer. Warnings ride in Result.Warnings only.
func TestEmitJSONL_NoWarningOnStdout(t *testing.T) {
	res := &Result{
		Write: &generate.WriteResult{
			Files: []generate.FileResult{
				{RelPath: "a.md", Action: generate.ActionCreated, Source: "src"},
			},
		},
		Warnings: []string{"could not save lock file: disk full", "invalid discovery: bad-frontmatter.md"},
	}

	var buf bytes.Buffer
	if err := EmitJSONL(&buf, res); err != nil {
		t.Fatalf("EmitJSONL: %v", err)
	}
	out := buf.String()

	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("non-JSON line on stdout: %q (%v)", line, err)
		}
		ev, _ := rec["event"].(string)
		if ev != "file" && ev != "summary" {
			t.Errorf("unexpected event %q on stdout; only file/summary allowed: %q", ev, line)
		}
		if ev == "warning" {
			t.Errorf("warning event must NEVER appear on stdout: %q", line)
		}
	}
	if strings.Contains(out, "could not save lock file") || strings.Contains(out, "invalid discovery") {
		t.Errorf("warning text leaked onto the JSONL stream:\n%s", out)
	}
}

// TestEmitJSONL_EmptyWriteRendersZeroSummary asserts the all-installed
// short-circuit seam: a Result whose Write is empty (or nil) renders as a
// lone zero-count summary line — byte-identical to the old
// EmitSummary(w,0,0,0,0,0) call the short-circuit used to make directly.
func TestEmitJSONL_EmptyWriteRendersZeroSummary(t *testing.T) {
	cases := map[string]*Result{
		"empty-write": {Write: &generate.WriteResult{}},
		"nil-write":   {Write: nil},
		"nil-result":  nil,
	}
	want := `{"event":"summary","created":0,"updated":0,"unchanged":0,"skipped":0,"conflicts":0}` + "\n"
	for name, res := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := EmitJSONL(&buf, res); err != nil {
				t.Fatalf("EmitJSONL: %v", err)
			}
			if buf.String() != want {
				t.Errorf("zero-summary mismatch\n want: %q\n got:  %q", want, buf.String())
			}
		})
	}
}

// TestResultCounts_DelegatesToSummary checks Counts() delegates to
// Write.Summary() and is nil-tolerant.
func TestResultCounts_DelegatesToSummary(t *testing.T) {
	res := &Result{Write: &generate.WriteResult{
		Files: []generate.FileResult{
			{Action: generate.ActionCreated},
			{Action: generate.ActionCreated},
			{Action: generate.ActionConflict},
		},
	}}
	c, u, un, s, cf := res.Counts()
	if c != 2 || u != 0 || un != 0 || s != 0 || cf != 1 {
		t.Errorf("Counts mismatch: got (%d,%d,%d,%d,%d)", c, u, un, s, cf)
	}

	// nil-tolerant
	var nilRes *Result
	if c, u, un, s, cf := nilRes.Counts(); c|u|un|s|cf != 0 {
		t.Errorf("nil Result Counts must be all-zero; got (%d,%d,%d,%d,%d)", c, u, un, s, cf)
	}
}

// TestMCPReadiness_NoTUIImports asserts the headless cores stay TUI-free:
// zero huh / bubbletea / lipgloss / glamour / charm imports anywhere in the
// nonint package OR internal/generate (production AND test files). The CLI
// adapter serialises the Result; these packages must never pull in chrome — a
// hard prerequisite for the Plan 42 stdio MCP server (stdout must be pure
// protocol).
//
// Plan 41 Phase 4 widens the scan to internal/generate because the
// `list --json` serializer (list_snapshot.go / SerializeJSON) now lives there
// beside SerializeCatalog — the scan proves the list serializer can't pull in
// chrome undetected. Phase 1 originally guarded nonint only.
func TestMCPReadiness_NoTUIImports(t *testing.T) {
	banned := []string{"huh", "bubbletea", "lipgloss", "glamour", "charm"}

	// "." is the nonint package dir (test cwd). "../generate" is the sibling
	// internal/generate package — both must stay TUI-free.
	dirs := []string{".", "../generate"}

	for _, dir := range dirs {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseDir(%s): %v", dir, err)
		}
		if len(pkgs) == 0 {
			t.Fatalf("ParseDir(%s): no packages found — scan would silently pass", dir)
		}
		for _, pkg := range pkgs {
			for fname, file := range pkg.Files {
				for _, imp := range file.Imports {
					path := strings.Trim(imp.Path.Value, `"`)
					for _, b := range banned {
						if strings.Contains(path, b) {
							t.Errorf("%s imports banned TUI dependency %q (contains %q) — %s must stay TUI-free", fname, path, b, dir)
						}
					}
				}
			}
		}
	}
}

// readFixture reads a committed testdata fixture; fatal on error.
func readFixture(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return string(b)
}

// assertGolden compares got against the committed golden file byte-for-byte
// and prints a line-level diff on mismatch so a refactor regression is
// readable.
func assertGolden(t *testing.T, goldenPath string, got []byte) {
	t.Helper()
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenPath, err)
	}
	if bytes.Equal(want, got) {
		return
	}
	gotLines := strings.Split(strings.TrimRight(string(got), "\n"), "\n")
	wantLines := strings.Split(strings.TrimRight(string(want), "\n"), "\n")
	t.Errorf("BYTE-IDENTITY FAILURE for %s — EmitJSONL output drifted from the pre-refactor golden.\n"+
		"The Result reshape MUST NOT alter the JSONL stream. Do not weaken this test.\n"+
		"want %d lines, got %d lines", goldenPath, len(wantLines), len(gotLines))
	max := len(gotLines)
	if len(wantLines) > max {
		max = len(wantLines)
	}
	for i := 0; i < max; i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w != g {
			t.Errorf("line %d:\n  want: %s\n  got:  %s", i+1, w, g)
		}
	}
}
