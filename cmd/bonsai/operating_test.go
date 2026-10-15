package main

// The operating skill kept in step with this command (design/plan-5.md, note 5.5.3 item 4): base's
// skills/operating-bonsai/SKILL.md, loaded as base:operating-bonsai, teaches every agent in a linked project to run
// Bonsai unattended, so a word, flag or error word it names that this build does not have would send an agent wrong.
// Read as committed (LF), it holds:
//   - every bonsai command line it shows parses against the word registry (word.go): each line of a fenced code block
//     and each inline code span that runs bonsai (bare, as bonsai.exe, or by a path such as /usr/local/bin/bonsai,
//     after any NAME=value assignments) names a word the registry has (a sub-word under its word: bonsai hook guard),
//     built, and every flag on it is that word's, built, with its value where it takes one, given once unless it may
//     repeat; bonsai --help --json and bonsai --version are bonsai's own;
//   - every fixed word it names in backquotes (a code span that is one kebab-case word, such as `needs-yes`) is a
//     word Bonsai has: an error word (format.ErrorWords, error.code's one home) or a rule of the guard and the stop
//     gate (guard.Rules, a guard record's rule: budget-spent, call-refused), each table the word's one home;
//   - every word the registry has is named at least once in such a line, the hook words among them (the skill names
//     them as Claude Code's, never run by hand), so a new word fails here until the skill teaches it;
//   - its size is at most 12 KB, so an agent reads it whole (note 5.5.3 item 3).
// A word, flag or error word changed in this command changes the skill in the same commit.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/guard"
)

// operatingSkillMax is the skill's largest size, in bytes as committed: 12 KB.
const operatingSkillMax = 12 * 1024

// readOperatingSkill reads the skill as committed, its line endings LF.
func readOperatingSkill(t *testing.T) string {
	t.Helper()
	p := filepath.Join(packageDir, "..", "..", "packs", "base", "skills", "operating-bonsai", "SKILL.md")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("base's operating skill cannot be read: %v", err)
	}
	return strings.ReplaceAll(string(raw), "\r\n", "\n")
}

// TestOperatingSkill: the skill as committed has none of the problems the file's comment lists.
func TestOperatingSkill(t *testing.T) {
	for _, p := range operatingProblems(readOperatingSkill(t)) {
		t.Errorf("packs/base/skills/operating-bonsai/SKILL.md: %s", p)
	}
}

// TestOperatingSkillCatchesDrift: copies of the skill, each broken one way, are each caught: a flag no word takes, a
// flag of another word, a flag not built yet, a word Bonsai does not have, an error word it does not have, every
// mention of each word in the registry taken out, and a skill past its size.
func TestOperatingSkillCatchesDrift(t *testing.T) {
	skill := readOperatingSkill(t)
	block := func(line string) string { return skill + "\n```sh\n" + line + "\n```\n" }
	cases := map[string]struct {
		text string
		want string
	}{
		"an unknown flag":                {block("bonsai ladder --task T-0001 --quick --json"), "bonsai ladder takes no --quick"},
		"a flag of another word":         {block("bonsai status --yes --json"), "bonsai status takes no --yes"},
		"a flag not built":               {block("bonsai status --line"), "bonsai status --line is not built yet"},
		"an unknown word":                {block("bonsai lint --json"), "bonsai lint is not a word Bonsai has"},
		"an unknown sub-word":            {block("bonsai hook nudge"), "bonsai hook nudge is not a word Bonsai has"},
		"a value missing":                {block("bonsai ladder --json --task"), "bonsai ladder --task needs a value"},
		"a flag given twice":             {block("bonsai ladder --task T-0001 --task T-0002"), "bonsai ladder takes --task once"},
		"an unknown error word":          {skill + "\nA refusal `budget-gone` means stop.\n", "`budget-gone` is not a word Bonsai has"},
		"a bad flag in a code span":      {skill + "\nRun `bonsai check --fix --json`.\n", "bonsai check takes no --fix"},
		"past its size":                  {skill + strings.Repeat("x", operatingSkillMax), "bytes, past"},
		"a line after a here-doc":        {block("bonsai message --kind report --body - <<'EOF'\nbonsai lint\nEOF\nbonsai ask --soon"), "bonsai ask takes no --soon"},
		"a bad flag on a continued line": {block("bonsai ladder --task T-0001 \\\n  --quick"), "bonsai ladder takes no --quick"},
	}
	for _, w := range helpEntries() {
		// Every mention, a line break inside a code span included: bonsai, then the word's own words.
		mention := regexp.MustCompile(`bonsai\s+` + strings.ReplaceAll(regexp.QuoteMeta(w.Name), " ", `\s+`) + `([^a-z-]|$)`)
		if !mention.MatchString(skill) {
			continue // TestOperatingSkill reports it
		}
		cases["no "+w.Name] = struct {
			text string
			want string
		}{mention.ReplaceAllString(skill, "bonsai --version$1"), "bonsai " + w.Name + " is never named"}
	}
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := cases[name]
		t.Run(name, func(t *testing.T) {
			found := false
			for _, p := range operatingProblems(c.text) {
				found = found || strings.Contains(p, c.want)
			}
			if !found {
				t.Errorf("a copy of the skill with %s is not caught (no problem says %q); it says: %q", name, c.want,
					operatingProblems(c.text))
			}
		})
	}
}

// operatingProblems lists what is wrong with the skill's text, each with its line: its size, each bonsai line that
// does not parse against the registry, each fixed word Bonsai does not have, and each word the registry has that it
// never names.
func operatingProblems(text string) []string {
	var out []string
	if n := len(text); n > operatingSkillMax {
		out = append(out, fmt.Sprintf("the skill is %d bytes, past its %d: an agent reads it whole", n, operatingSkillMax))
	}
	named := map[string]bool{}
	for _, l := range skillCommandLines(text) {
		p, ok := parseSkillCommand(l.text, l.fenced, named)
		if ok && p != "" {
			out = append(out, fmt.Sprintf("line %d: %s (in %q)", l.line, p, l.text))
		}
	}
	for _, s := range skillSpans(text) {
		if kebabWord.MatchString(s.text) && !bonsaiHasWord(s.text) {
			out = append(out, fmt.Sprintf("line %d: `%s` is not a word Bonsai has: no error word (format.ErrorWords) "+
				"and no rule of the guard (guard.Rules)", s.line, s.text))
		}
	}
	for _, w := range helpEntries() {
		if w.Later == "" && !named[w.Name] {
			out = append(out, "bonsai "+w.Name+" is never named in a bonsai line: the skill teaches every word the "+
				"registry has (the hook words as Claude Code's, never run by hand)")
		}
	}
	return out
}

// kebabWord is a code span that is one fixed word: lower-case letters and digits, with a dash at least once.
var kebabWord = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)+$`)

// bonsaiHasWord reports a fixed word Bonsai has: an error word, or a rule of the guard and the stop gate.
func bonsaiHasWord(w string) bool {
	if _, ok := format.ErrorWord(w); ok {
		return true
	}
	for _, r := range guard.Rules {
		if r.Name == w {
			return true
		}
	}
	return false
}

// skillLine is one command line of the skill: its text, its line in the file, and whether a fenced block holds it.
type skillLine struct {
	text   string
	line   int
	fenced bool
}

// skillCommandLines gives each line of the skill's fenced code blocks (a line ending in a backslash joined with the
// next; a here-document's body passed over) and each of its inline code spans.
func skillCommandLines(text string) []skillLine {
	var out []skillLine
	lines := strings.Split(text, "\n")
	fence, heredoc := "", ""
	for i := 0; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		switch {
		case fence == "" && (strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")):
			fence = t[:3]
		case fence != "" && heredoc != "":
			if t == heredoc {
				heredoc = ""
			}
		case fence != "" && strings.HasPrefix(t, fence):
			fence = ""
		case fence != "":
			first := i
			for strings.HasSuffix(t, `\`) && i+1 < len(lines) {
				i++
				t = strings.TrimSuffix(t, `\`) + " " + strings.TrimSpace(lines[i])
			}
			out = append(out, skillLine{text: t, line: first + 1, fenced: true})
			_, heredoc = skillTokens(t)
		}
	}
	for _, s := range skillSpans(text) {
		out = append(out, skillLine{text: s.text, line: s.line})
	}
	return out
}

// skillSpan is one inline code span outside the fenced blocks: its text (a line break in it read as a space) and the
// line it starts on.
type skillSpan struct {
	text string
	line int
}

// skillSpans gives the inline code spans of the text outside its fenced blocks.
func skillSpans(text string) []skillSpan {
	lines := strings.Split(text, "\n")
	fence := ""
	for i, l := range lines {
		t := strings.TrimSpace(l)
		switch {
		case fence == "" && (strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")):
			fence = t[:3]
			lines[i] = ""
		case fence != "":
			if strings.HasPrefix(t, fence) {
				fence = ""
			}
			lines[i] = ""
		}
	}
	prose := strings.Join(lines, "\n")
	var out []skillSpan
	for _, m := range codeSpan.FindAllStringSubmatchIndex(prose, -1) {
		out = append(out, skillSpan{
			text: strings.TrimSpace(strings.ReplaceAll(prose[m[2]:m[3]], "\n", " ")),
			line: strings.Count(prose[:m[0]], "\n") + 1,
		})
	}
	return out
}

// codeSpan is an inline code span: text between single backquotes.
var codeSpan = regexp.MustCompile("`([^`]+)`")

// assignment is a shell's NAME=value before a command.
var assignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// isBonsaiProgram reports a program word that runs bonsai: bonsai or bonsai.exe, bare or by a path.
func isBonsaiProgram(w string) bool {
	w = strings.ReplaceAll(w, `\`, "/")
	base := strings.ToLower(w[strings.LastIndex(w, "/")+1:])
	return base == "bonsai" || base == "bonsai.exe"
}

// parseSkillCommand reads one command line against the registry. ok is false for a line that does not run bonsai,
// and for an inline span naming the program alone (`bonsai`, a mention); else the problem is "" when the line parses,
// and each word it names (with its parent word, for a sub-word) is marked in named.
func parseSkillCommand(line string, fenced bool, named map[string]bool) (problem string, ok bool) {
	toks, _ := skillTokens(line)
	i := 0
	for i < len(toks) && assignment.MatchString(toks[i]) {
		i++
	}
	if i >= len(toks) || !isBonsaiProgram(toks[i]) {
		return "", false
	}
	args := toks[i+1:]
	if len(args) == 0 {
		if fenced {
			return "bonsai with no word runs nothing", true
		}
		return "", false
	}
	if strings.HasPrefix(args[0], "-") {
		if helpRequest(args) || (len(args) == 1 && (args[0] == "--version" || args[0] == "--help" || args[0] == "-h")) {
			return "", true
		}
		return "bonsai " + strings.Join(args, " ") + " is not one of bonsai's own lines (--version, --help, --help --json)", true
	}
	w, found := words[args[0]]
	if !found {
		return "bonsai " + args[0] + " is not a word Bonsai has", true
	}
	named[w.Name] = true
	args = args[1:]
	if len(w.Subs) > 0 && len(args) > 0 && args[0] != "--help" && args[0] != "-h" {
		var sub *Word
		for _, s := range w.Subs {
			if s.Name == w.Name+" "+args[0] {
				sub = s
			}
		}
		if sub == nil {
			return "bonsai " + w.Name + " " + args[0] + " is not a word Bonsai has", true
		}
		w, args = sub, args[1:]
		named[w.Name] = true
	}
	if w.Later != "" {
		return "bonsai " + w.Name + " is not built yet (" + w.Later + ")", true
	}
	given := map[string]bool{}
	for j := 0; j < len(args); j++ {
		a := strings.TrimSuffix(strings.TrimSuffix(strings.TrimLeft(args[j], "[("), "..."), "]")
		a = strings.TrimSuffix(a, ")")
		switch {
		case a == "--help" || a == "-h":
			continue
		case strings.HasPrefix(a, "--"):
		case strings.HasPrefix(a, "-") && a != "-":
			return "bonsai " + w.Name + " takes no " + a, true
		case strings.Contains(w.Args, "<"):
			continue // a word's own argument: answer's <key>
		default:
			return "bonsai " + w.Name + " takes no argument " + fmt.Sprintf("%q", a), true
		}
		name, _, hasValue := strings.Cut(a, "=")
		f := w.flag(name)
		switch {
		case f == nil:
			return "bonsai " + w.Name + " takes no " + name, true
		case f.Later != "":
			return "bonsai " + w.Name + " " + name + " is not built yet (" + f.Later + ")", true
		case f.Value == "" && hasValue:
			return "bonsai " + w.Name + " " + name + " takes no value", true
		case given[name] && !f.Many:
			return "bonsai " + w.Name + " takes " + name + " once", true
		case f.Value != "" && !hasValue:
			if j+1 >= len(args) || (strings.HasPrefix(args[j+1], "-") && args[j+1] != "-") {
				return "bonsai " + w.Name + " " + name + " needs a value", true
			}
			j++
		}
		given[name] = true
	}
	return "", true
}

// skillTokens splits a command line into its words, as a shell would, up to its first operator (a pipe, a redirect,
// &&, ||, ;, a comment): single and double quotes kept together; a placeholder in angle brackets (<the brief's from>)
// one word, whatever it holds. heredoc is the delimiter of a here-document the line starts (<<'EOF'), "" for none.
func skillTokens(s string) (toks []string, heredoc string) {
	var cur strings.Builder
	in := false
	flush := func() {
		if in {
			toks = append(toks, cur.String())
		}
		cur.Reset()
		in = false
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		start := !in
		switch {
		case ch == ' ' || ch == '\t':
			flush()
		case ch == '\'' || ch == '"':
			end := strings.IndexByte(s[i+1:], ch)
			if end < 0 {
				end = len(s) - i - 1
			}
			cur.WriteString(s[i+1 : i+1+end])
			in = true
			i += end + 1
		case ch == '<' && start && i+1 < len(s) && s[i+1] == '<':
			flush()
			rest := strings.Fields(s[i+2:])
			if len(rest) > 0 {
				heredoc = strings.Trim(rest[0], `'"-`)
			}
			return toks, heredoc
		case ch == '<' && i+1 < len(s) && s[i+1] != ' ' && strings.IndexByte(s[i:], '>') > 0:
			end := strings.IndexByte(s[i:], '>')
			cur.WriteString(s[i : i+end+1])
			in = true
			i += end
		case ch == '<' || ch == '>' || ch == ';' || (ch == '#' && start) ||
			(start && (ch == '|' || ch == '&')):
			if ch == '>' && in && strings.Trim(cur.String(), "0123456789") == "" {
				cur.Reset() // 2>: the number is the redirect's
				in = false
			}
			flush()
			return toks, ""
		default:
			cur.WriteByte(ch)
			in = true
		}
	}
	flush()
	return toks, ""
}
