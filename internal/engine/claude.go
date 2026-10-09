package engine

// Claude Code's version against the floor (spec §7, "No Claude Code version pin"): warn only, never refuse. The floor
// is the higher of two: Bonsai's own, ClaudeCodeFloor, the oldest Claude Code its tests ran on (the gate measured
// 2.1.294 on both sides: records/gate-skeleton.md §2.10), which each release sets again; and any locked pack's
// needs.claude_code (bonsai/pack.yaml), read from the lock's declares so CI needs no pack. bonsai check warns when
// claude --version is older (claude-code-old), and when it cannot be read: not on the PATH, failing, or an answer with
// no version in it (claude-code-unknown, spec §7: "whether its output form stays stable is unchecked"). No command,
// hook or rung reads the version to decide anything; status --json lists the floor as a needs entry (kind tool, name
// claude-code), never a problem, and status --full shows the version read (status's checks).
//
// claude --version is read defensively: run with no input, a timeout, and its output's first version-shaped word
// (digits.digits.digits) taken, whatever surrounds it ("2.1.294 (Claude Code)" today). The caller gives the binary's
// path: tests give a fake, and no test reaches the real claude.

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// ClaudeCodeFloor is Bonsai's own Claude Code floor (spec §7; gate report §2.10).
const ClaudeCodeFloor = "2.1.294"

// versionTimeout is the longest claude --version may take.
const versionTimeout = 20 * time.Second

var versionWord = regexp.MustCompile(`(^|[^0-9.])([0-9]+)\.([0-9]+)\.([0-9]+)([^0-9.]|$)`)

// ParseVersion finds the first version-shaped word (digits.digits.digits) in a version line; ok false when there is
// none.
func ParseVersion(s string) (v [3]int, text string, ok bool) {
	m := versionWord.FindStringSubmatch(s)
	if m == nil {
		return v, "", false
	}
	for i := 0; i < 3; i++ {
		n, err := strconv.Atoi(m[2+i])
		if err != nil {
			return v, "", false
		}
		v[i] = n
	}
	return v, m[2] + "." + m[3] + "." + m[4], true
}

// older reports whether version a is older than b.
func older(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// Floor is a workspace's Claude Code floor: Bonsai's own, or a locked pack's needs.claude_code when that is higher
// (a version Bonsai does not read is left out); from names where it came from ("bonsai" or the pack's id).
func Floor(lock *workspace.Lock) (version, from string) {
	version, from = ClaudeCodeFloor, "bonsai"
	best, _, _ := ParseVersion(ClaudeCodeFloor)
	if lock == nil {
		return
	}
	for _, lp := range lock.Packs {
		d, err := lp.Declared()
		if err != nil || d.Needs == nil || d.Needs.ClaudeCode == nil {
			continue
		}
		v, text, ok := ParseVersion(*d.Needs.ClaudeCode)
		if ok && older(best, v) {
			best, version, from = v, text, lp.ID
		}
	}
	return
}

// ClaudeVersion runs `<bin> --version` with no input and a timeout, and gives its output's first line.
func ClaudeVersion(bin string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", errors.New("claude --version took longer than " + versionTimeout.String())
	}
	if err != nil {
		return "", errors.New("claude --version failed: " + oneLine(err.Error()))
	}
	return firstLine(stdout.String()), nil
}

// LookClaude finds claude on the PATH: its path, or ErrNoClaude.
func LookClaude() (string, error) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return "", ErrNoClaude
	}
	return bin, nil
}

// ClaudeState is Claude Code's version against the floor: what was read, and how it compares.
type ClaudeState struct {
	Version string // the version read, "" when none was
	Floor   string
	From    string // where the floor came from: bonsai or a pack's id
	State   string // ok, old, or unknown (not on the PATH, failed, or an answer with no version)
	Why     string // for unknown: why, one sentence; "" otherwise
}

// JudgeClaude compares what ask gives (claude --version's first line, or an error: ErrNoClaude when it is not on the
// PATH) with the workspace's floor.
func JudgeClaude(lock *workspace.Lock, ask func() (string, error)) ClaudeState {
	st := ClaudeState{}
	st.Floor, st.From = Floor(lock)
	out, err := ask()
	switch {
	case errors.Is(err, ErrNoClaude):
		st.State, st.Why = "unknown", "Claude Code is not on the PATH"
		return st
	case err != nil:
		st.State, st.Why = "unknown", ascii(oneLine(err.Error()))
		return st
	}
	v, text, ok := ParseVersion(out)
	if !ok {
		shown := out
		if len(shown) > 60 {
			shown = shown[:60] + "..."
		}
		st.State, st.Why = "unknown", "claude --version gave an answer Bonsai does not read: "+strconvQuote(shown)
		return st
	}
	st.Version = text
	floor, _, _ := ParseVersion(st.Floor)
	st.State = "ok"
	if older(v, floor) {
		st.State = "old"
	}
	return st
}

func strconvQuote(s string) string { return strconv.QuoteToASCII(s) }

// CompareClaudeCode adds check's Claude Code warnings (spec §7): older than the floor, or a version it cannot read.
// Neither changes the exit code. A nil ask compares nothing (tests that do not give one).
func CompareClaudeCode(r *CheckResult, ask func() (string, error)) {
	if r == nil || r.Config == nil || ask == nil {
		return
	}
	st := JudgeClaude(r.Lock, ask)
	floor := st.Floor
	if st.From != "bonsai" {
		floor += " (the pack " + st.From + "'s needs.claude_code)"
	} else {
		floor += " (Bonsai's own)"
	}
	switch st.State {
	case "old":
		r.add("claude-code-old", "", "", "Claude Code here is "+st.Version+", older than the floor "+floor+": warn only, nothing refuses for it "+
			"(spec section 7)", "a person updates Claude Code on this machine, run: claude update")
	case "unknown":
		next := "a person checks Claude Code on this machine, run: claude --version"
		if strings.Contains(st.Why, "not on the PATH") {
			next = "where sessions run, a person installs Claude Code " + st.Floor + " or newer; where none run (CI), nothing is needed"
		}
		r.add("claude-code-unknown", "", "", st.Why+", so its version was not compared with the floor "+floor, next)
	}
}
