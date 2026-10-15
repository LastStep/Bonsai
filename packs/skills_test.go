package packs

// Base's skills' examples read by Bonsai's own reader (design/plan-5.md, 5.5.5's fix round): an agent copies a skill's
// example into a file Bonsai reads, so an example Bonsai cannot read sends it wrong, and in bonsai.yaml it locks the
// agent out (the guard cannot read the protected paths and blocks every call until the person repairs the file: the
// workspace skill's packs[].roles example, a flow mapping, did just that in 5.5.5's S4). Read as committed (LF),
// every skill under base/skills/ holds:
//   - each fenced yaml block reads with reader.ReadYAML, unrefused, but a GitHub workflow's (base:ci's template and its
//     steps: GitHub reads them, never Bonsai);
//   - each fenced markdown block's frontmatter reads with reader.ReadMarkdown, unrefused;
//   - each example cell of a fields table (| Field | Meaning | Allowed values | Example |) that is one code span reads
//     as that field's value, "<key>: <example>", with reader.ReadYAML, unrefused; a cell in words ("as in the
//     template", "under Mappings, below") points to a block the first item reads.
// A template's blanks (<what goes here>) are read as words, and an example with no format: key (a part of a file) is
// read under format 1 too, so a flow mapping, which format 1 refuses, is caught wherever it is.

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/reader"
)

// skillExample is one example a skill shows: where it is, what kind it is, and its text.
type skillExample struct {
	where string // file and line
	kind  string // yaml, markdown or cell
	text  string
}

// githubWorkflow marks a yaml block as a GitHub workflow's (base:ci's template and its steps): GitHub reads it.
var githubWorkflow = regexp.MustCompile(`(?m)^\s*(# \.github/workflows/|- name: |on:\s)`)

// exampleCell is a fields table's row whose last cell is one code span: the field and the span's text.
var exampleCell = regexp.MustCompile("^\\| `([^`]+)` \\|.*\\| `([^`]*)` \\|$")

// skillExamples lists the examples of one skill's text, file being its name for messages.
func skillExamples(file, text string) []skillExample {
	var out []skillExample
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		l := strings.TrimSpace(lines[i])
		if l == "```yaml" || l == "```markdown" {
			kind := strings.TrimPrefix(l, "```")
			start := i + 1
			// A block inside a list item is indented as its fence is: the example is the text without it.
			indent := lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " "))]
			var b strings.Builder
			for i++; i < len(lines) && strings.TrimSpace(lines[i]) != "```"; i++ {
				b.WriteString(strings.TrimPrefix(lines[i], indent) + "\n")
			}
			out = append(out, skillExample{where: file + ":" + strconv.Itoa(start), kind: kind, text: b.String()})
			continue
		}
		if m := exampleCell.FindStringSubmatch(lines[i]); m != nil {
			key := m[1]
			if j := strings.LastIndexAny(key, ".]"); j >= 0 {
				key = key[j+1:]
			}
			if key == "" || strings.ContainsAny(key, "<>") {
				key = "x"
			}
			// A table cell escapes | as \|.
			out = append(out, skillExample{where: file + ":" + strconv.Itoa(i+1), kind: "cell",
				text: key + ": " + strings.ReplaceAll(m[2], `\|`, "|") + "\n"})
		}
	}
	return out
}

// placeholder is a template's blank, <what goes here>, which the agent fills: read as a word.
var placeholder = regexp.MustCompile(`<[^<>\n]*>`)

// exampleProblem says why one example does not read, or "" when it does; an example with no format: key is read
// again with a format: line put first.
func exampleProblem(e skillExample) string {
	if e.kind == "yaml" && githubWorkflow.MatchString(e.text) {
		return ""
	}
	text := placeholder.ReplaceAllString(e.text, "blank")
	read := func(s string) reader.Result {
		if e.kind == "markdown" {
			return reader.ReadMarkdown([]byte(s))
		}
		return reader.ReadYAML([]byte(s))
	}
	r := read(text)
	if r.Outcome == reader.Format0 {
		const f1 = "format: bonsai.example/1\n"
		if e.kind == "markdown" && strings.HasPrefix(text, "---\n") {
			r = read("---\n" + f1 + text[4:])
		} else if e.kind != "markdown" {
			r = read(f1 + text)
		}
	}
	if err := r.Err(); err != nil {
		return e.where + ": its " + e.kind + " example does not read as Bonsai reads it: " + err.Error() +
			" (" + strings.TrimSpace(firstLine(e.text)) + ")"
	}
	return ""
}

// firstLine is a text's first line, for a message.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// baseSkills reads every skill of base as committed, line endings LF, by its path under base/, sorted.
func baseSkills(t *testing.T) map[string]string {
	t.Helper()
	found, err := filepath.Glob(filepath.Join(baseDir, "skills", "*", "SKILL.md"))
	if err != nil || len(found) == 0 {
		t.Fatalf("base's skills cannot be listed: %v (%d found)", err, len(found))
	}
	out := map[string]string{}
	for _, f := range found {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("%s cannot be read: %v", f, err)
		}
		out[path.Join("skills", filepath.Base(filepath.Dir(f)), "SKILL.md")] = strings.ReplaceAll(string(raw), "\r\n", "\n")
	}
	return out
}

// TestBaseSkillExamplesRead: every example in base's skills reads as Bonsai reads it.
func TestBaseSkillExamplesRead(t *testing.T) {
	skills := baseSkills(t)
	names := make([]string, 0, len(skills))
	for n := range skills {
		names = append(names, n)
	}
	sort.Strings(names)
	total := 0
	for _, n := range names {
		for _, e := range skillExamples(n, skills[n]) {
			total++
			if p := exampleProblem(e); p != "" {
				t.Error(p)
			}
		}
	}
	if total < 40 {
		t.Errorf("only %d examples found in base's skills: the reading of blocks or table cells has broken", total)
	}
}

// TestBaseSkillExamplesCatch: a copy of a skill with an example Bonsai cannot read is caught: a flow mapping in a
// table cell (the form an agent copied into bonsai.yaml), in a yaml block and in one inside a list item, and a
// markdown block's broken frontmatter.
func TestBaseSkillExamplesCatch(t *testing.T) {
	cases := map[string]string{
		"a flow mapping in a cell": "| Field | Meaning | Allowed values | Example |\n|---|---|---|---|\n" +
			"| `packs[].roles` | Limits | a mapping, or null | `{\"builder\":{\"minutes\":{\"default\":240,\"max\":480}}}` |\n",
		"a flow mapping in a yaml block":      "```yaml\nroles: {builder: {minutes: {default: 240}}}\n```\n",
		"a tab in a markdown frontmatter":     "```markdown\n---\nformat: bonsai.task/1\nid:\tT-0001\n---\n```\n",
		"a flow mapping in an indented block": "- step\n\n  ```yaml\n  ratchets: {tests: 3}\n  ```\n",
	}
	keys := make([]string, 0, len(cases))
	for k := range cases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		caught := false
		for _, e := range skillExamples("copy.md", cases[k]) {
			if exampleProblem(e) != "" {
				caught = true
			}
		}
		if !caught {
			t.Errorf("%s: not caught", k)
		}
	}
}
