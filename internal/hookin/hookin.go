// Package hookin is Bonsai's one hook adapter (spec §7: "One hook adapter reads Claude Code's payload (tool names,
// backslash paths on Windows, unknown fields). A payload it cannot read blocks."; design/plan-5.md, 5.3.0). Every
// `bonsai hook` word reads Claude Code's payload through it, so a field Claude Code adds or renames is read in one
// place, and it holds the frame of the hooks that block.
//
// What it gives:
//   - payload.go: Read and Parse, the one decode of a payload, into Payload: the common fields and each event's own,
//     the tool inputs Bonsai reads (the spawn tools' among them), and the first fault met; ReadJudged and ParseJudged
//     for a reader that judges some tools only, every other tool's input passed over unread;
//   - hash.go: InputHash, a tool input's hash (contract §8.1), beside the one decode of the input it hashes;
//   - frame.go: Frame, the run of a hook that blocks (the guard; the stop gate from step 5.3.4): its own timer, the
//     work in a goroutine, a panic as a block, the answer and its exit code;
//   - this file: where a hook finds its project, and the payload's limits.
//
// What a payload that does not read does is the caller's (plan 5.3.0 note 2): the guard and the stop gate block on an
// error or a fault; `bonsai hook record` writes nothing and exits 0; `bonsai hook start` prints what it can.
package hookin

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// MaxInput caps the payload a hook reads, in bytes. A larger one is not read: the guard blocks it, the recorder writes
// nothing.
const MaxInput = 64 << 20

// ProjectEnv names the variable Claude Code sets for a hook to the project's folder, the one the session started in.
// A hook finds its project from it, never from git (whose answer comes from files an agent may edit); the guard never
// from the payload's cwd either (an agent's cd moves that).
const ProjectEnv = "CLAUDE_PROJECT_DIR"

// Read reads one payload from a hook's stdin, at most MaxInput bytes, and parses it (Parse). terminal is true when
// stdin is a terminal: a hook run by hand with nothing piped in, which is not waited on. Its error, for a payload
// that cannot be read at all, says why in ASCII words that follow "the hook input".
func Read(stdin io.Reader, terminal bool) (*Payload, error) { return ReadJudged(stdin, terminal, nil) }

// ReadJudged is Read for a reader that judges only the tools judged reports true for (ParseJudged): every other tool's
// tool_input is passed over, held only to being valid JSON. judged nil reads every tool's, as Read.
func ReadJudged(stdin io.Reader, terminal bool, judged func(tool string) bool) (*Payload, error) {
	if terminal {
		return nil, errors.New("is a terminal, not a payload (a hook reads Claude Code's payload on stdin)")
	}
	if stdin == nil {
		return nil, errors.New("is empty")
	}
	raw, err := io.ReadAll(io.LimitReader(stdin, MaxInput+1))
	if err != nil {
		return nil, errors.New("cannot be read (" + ascii(err.Error()) + ")")
	}
	if len(raw) > MaxInput {
		return nil, errors.New("is larger than " + strconv.Itoa(MaxInput>>20) + " MiB")
	}
	return ParseJudged(raw, judged)
}

// Drain reads and drops a payload, so Claude Code never writes into a closed pipe: what a hook does in a folder with
// no bonsai.yaml, where it answers at once and records nothing (spec §3, §7).
func Drain(stdin io.Reader, terminal bool) {
	if terminal || stdin == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(stdin, MaxInput))
}

// FindProject finds the project a folder belongs to: the folder itself or the nearest one above it that holds
// bonsai.yaml (FindProjects' first). linked is false when there is none (spec §7: in a project with no bonsai.yaml a
// hook allows and records nothing), and root is then the folder, made absolute; err says why the folder cannot be read.
// Every hook finds its project with it; the recorder, with no CLAUDE_PROJECT_DIR, from the payload's cwd.
func FindProject(dir string) (root string, linked bool, err error) {
	roots, abs, err := FindProjects(dir)
	if err != nil {
		return "", false, err
	}
	if len(roots) == 0 {
		return abs, false, nil
	}
	return roots[0], true, nil
}

// FindProjects finds every project holding a folder (design/plan-5.md, 5.3.1 note 1; contract §13: "the project is
// the one holding the path"): the folder itself and each folder above it, up to the volume's root, that holds
// bonsai.yaml, nearest first. It walks past any .git: a nested repository, or a .git someone placed in a subfolder,
// does not hide the project around it. abs is the folder, absolute and clean; err says why it cannot be read (it must
// be a folder that exists, and whether each folder on the way holds bonsai.yaml must be told). It costs one Lstat per
// folder on the way.
func FindProjects(dir string) (roots []string, abs string, err error) {
	abs, err = filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}
	if fi, err := os.Stat(abs); err != nil {
		return nil, "", err
	} else if !fi.IsDir() {
		return nil, "", errors.New("it is not a folder")
	}
	roots, err = walk(abs, true, nil)
	return roots, abs, err
}

// Holding gives the folders at or above the folder dir (absolute and clean) that hold bonsai.yaml, nearest first, up
// to the volume's root: FindProjects' walk, for a folder that need not exist (a new file's folder). A folder where
// bonsai.yaml cannot be looked for (one that cannot be searched, so nothing can be written in it either) is passed
// over, and the walk goes on above it. seen keeps what each folder held, so the forms of one path, which share most
// of their folders, look each up once; nil keeps nothing.
func Holding(dir string, seen map[string]bool) []string {
	roots, _ := walk(dir, false, seen)
	return roots
}

// walk is FindProjects' and Holding's walk up from dir. strict stops it at a folder where bonsai.yaml cannot be looked
// for (an error other than its absence), with that error.
func walk(dir string, strict bool, seen map[string]bool) ([]string, error) {
	var roots []string
	for d := dir; ; {
		holds, known := seen[d]
		if !known {
			_, err := os.Lstat(filepath.Join(d, workspace.ConfigFile))
			if strict && err != nil && !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
			holds = err == nil
			if seen != nil {
				seen[d] = holds
			}
		}
		if holds {
			roots = append(roots, d)
		}
		parent := filepath.Dir(d)
		if parent == d {
			return roots, nil
		}
		d = parent
	}
}

// ID gives an id as Bonsai keeps it: itself when it can name a log file, by the log file's own rule
// (record.LogFileName: letters, digits, _ and -, at most 128; Claude Code's session ids are UUIDs, its tool-use ids
// toolu_ ids), else "". The session id is held to it as the payload is read; a subagent's and a tool use's are held
// to it by the recorder, which keeps them in its records.
func ID(s string) string {
	if s == "" {
		return ""
	}
	if _, err := record.LogFileName(s, time.Time{}); err != nil {
		return ""
	}
	return s
}

// ascii keeps a message on one ASCII line: printable ASCII as itself, a line break or tab as a space, any other
// character as a Go escape, so a refusal prints unbroken in PowerShell 5.1 (spec §3).
func ascii(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case r >= 0x20 && r < 0x7f:
			b.WriteRune(r)
		default:
			q := strconv.QuoteRuneToASCII(r)
			b.WriteString(q[1 : len(q)-1])
		}
	}
	return b.String()
}

// quoted gives a value as an ASCII Go string, cut to max runes.
func quoted(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		s = string(r[:max]) + "..."
	}
	return strconv.QuoteToASCII(s)
}
