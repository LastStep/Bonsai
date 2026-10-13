package vergit

// How the rung runs git: one program, a deadline, a fixed environment and fixed settings, so that nothing the user
// or the repository has set can point git at another repository, run a helper, wait for a person or quietly stop
// looking at a change.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DefaultTimeout is the whole judgment's limit when Input gives none: Bonsai's own ver-git rung allows 120 seconds.
const DefaultTimeout = 120 * time.Second

// waitDelay is how long a git call's pipes may stay open after it ends or is killed (a helper it started holding
// them), before the call gives up on them.
const waitDelay = 2 * time.Second

// keepEnv are the only variables of this process's environment a git call keeps: where programs are (PATH; git
// itself needs no more on Linux, and Git for Windows' launcher finds its own folders from it) and what Windows needs to
// start a program at all. Everything else is dropped: HOME and XDG_CONFIG_HOME (so git finds no user configuration,
// global excludes or global attributes by default), every GIT_* variable (GIT_DIR, GIT_WORK_TREE, GIT_INDEX_FILE,
// GIT_OBJECT_DIRECTORY and GIT_ALTERNATE_OBJECT_DIRECTORIES would point git at other files; GIT_CONFIG_* and
// GIT_CONFIG_PARAMETERS would add settings; GIT_EXEC_PATH would swap its helpers; GIT_TRACE* would write into its
// output; GIT_NAMESPACE and the pathspec switches would change what it reads; GIT_CEILING_DIRECTORIES is set anew, in
// newRunner), PAGER and the editors. Windows variable names are matched without case. Every call is one of git's own
// commands, which no alias can replace, with --no-pager.
var keepEnv = []string{"PATH", "PATHEXT", "SYSTEMROOT", "WINDIR", "TEMP", "TMP", "COMSPEC"}

// fixedEnv is added to every call. GIT_CONFIG_NOSYSTEM and GIT_CONFIG_GLOBAL (set to the null device in env) leave
// only the repository's own configuration, which no variable can switch off; the settings in pins override the parts
// of it that could hide a change. GIT_ATTR_NOSYSTEM drops the system's attributes. GIT_NO_REPLACE_OBJECTS stops
// refs/replace from showing one commit as another (a rewritten base made to look whole); GIT_GRAFT_FILE (also the
// null device) does the same for info/grafts. GIT_OPTIONAL_LOCKS=0 keeps these reads from writing the index.
// GIT_TERMINAL_PROMPT=0 and GIT_NO_LAZY_FETCH=1 (newer git; protocol.allow=never below covers older) keep git from
// asking anyone anything or reaching a network. LC_ALL=C keeps its messages in one language.
var fixedEnv = []string{"GIT_CONFIG_NOSYSTEM=1", "GIT_ATTR_NOSYSTEM=1", "GIT_NO_REPLACE_OBJECTS=1",
	"GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "GIT_NO_LAZY_FETCH=1", "LC_ALL=C", "LANGUAGE="}

// pins are settings given on every call's command line, where they outrank the repository's configuration:
//   - core.fsmonitor=false: a file-system monitor (a program the repository names) tells git which files changed,
//     and one that lies hides an edit; core.untrackedCache=false: the index's cache of untracked folders is not
//     trusted either;
//   - core.trustctime=true, core.checkStat=default, core.ignoreStat=false: git looks at every file's full stat
//     information before it believes a file is unchanged (with less, a file edited to the same size and given back its
//     time reads as unchanged); a stricter look only makes git compare more contents, it never reports a change that
//     is not there;
//   - core.attributesFile= (none): no attributes but the repository's own (.gitattributes and info/attributes);
//   - core.commitGraph=false: ancestry is read from the commits themselves, not from a cache file that could be
//     made to say otherwise;
//   - protocol.allow=never: no fetch of a missing object (a partial clone) reaches a network;
//   - core.excludesFile is pinned per run, and the user's core.autocrlf and core.eol kept (see userSettings).
var pins = []string{"core.fsmonitor=false", "core.untrackedCache=false", "core.trustctime=true",
	"core.checkStat=default", "core.ignoreStat=false", "core.attributesFile=", "core.commitGraph=false",
	"protocol.allow=never"}

// gitRunner runs git for one judgment.
type gitRunner struct {
	ctx      context.Context
	prog     string   // the git program
	dir      string   // the folder git runs in
	excludes string   // core.excludesFile for every call: the user's own, or "" for none
	keep     []string // the user's line-end settings, given on every call (userSettings)
	extra    []string // more settings for one kind of read, given last (the contents' read: every filter off)
	env      []string // the fixed environment
}

// gitError is a git call that failed: its arguments, its exit code (-1 when it did not run to an exit), and the first
// line of what it said.
type gitError struct {
	args []string
	code int
	msg  string
	err  error
}

func (e *gitError) Error() string {
	s := "git " + strings.Join(e.args, " ")
	if e.msg != "" {
		return s + ": " + e.msg
	}
	return s + ": " + oneLine(e.err.Error())
}

func (e *gitError) Unwrap() error { return e.err }

// newRunner finds the git program (prog, or git on this process's PATH when prog is "") and builds the fixed
// environment for the checkout whose top folder is dir. GIT_CEILING_DIRECTORIES, its parent, keeps git from taking up
// a repository above it: the checkout judged is the one at dir, or none.
func newRunner(ctx context.Context, prog, dir string) (*gitRunner, error) {
	if prog == "" {
		found, err := exec.LookPath("git")
		if err != nil {
			return nil, err
		}
		prog = found
	}
	env := []string{"GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_GRAFT_FILE=" + os.DevNull,
		"GIT_CEILING_DIRECTORIES=" + filepath.Dir(dir)}
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		for _, keep := range keepEnv {
			if k == keep || (runtime.GOOS == "windows" && strings.EqualFold(k, keep)) {
				env = append(env, kv)
				break
			}
		}
	}
	return &gitRunner{ctx: ctx, prog: prog, dir: dir, env: append(env, fixedEnv...)}, nil
}

// run runs git with the pins and returns its standard output. stdin may be nil.
func (g *gitRunner) run(stdin []byte, args ...string) ([]byte, error) {
	full := []string{"--no-pager"}
	for _, p := range pins {
		full = append(full, "-c", p)
	}
	full = append(full, "-c", "core.excludesFile="+g.excludes)
	for _, k := range g.keep {
		full = append(full, "-c", k)
	}
	for _, k := range g.extra {
		full = append(full, "-c", k)
	}
	return g.exec(g.env, stdin, append(full, args...), args)
}

// exec runs git with exactly these arguments and this environment, in the runner's folder, within its deadline.
func (g *gitRunner) exec(env []string, stdin []byte, full, shown []string) ([]byte, error) {
	cmd := exec.CommandContext(g.ctx, g.prog, full...)
	cmd.Dir = g.dir
	cmd.Env = env
	cmd.WaitDelay = waitDelay
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.Bytes(), nil
	}
	e := &gitError{args: shown, code: -1, msg: firstLine(stderr.String()), err: err}
	if g.ctx.Err() != nil {
		e.msg, e.err = "git did not answer in time", g.ctx.Err()
		return nil, e
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		e.code = exit.ExitCode()
	}
	return stdout.Bytes(), e
}

// exitCode is a failed call's exit code, or -1.
func exitCode(err error) int {
	var e *gitError
	if errors.As(err, &e) {
		return e.code
	}
	return -1
}

// userSettings reads, once, the user's own settings that every later call keeps, with the user's environment less
// the variables that point git at another repository or run anything (every GIT_* but GIT_CONFIG*):
//   - core.excludesFile, as the user's git in this checkout reads it (the repository's own value wins, as in git), or
//     git's default, $XDG_CONFIG_HOME/git/ignore or ~/.config/git/ignore. That file is what hides a file from the
//     user's git status, so every call ignores by it; a value can only add a file to look through, never take one
//     away. "" when there is no such file.
//   - core.autocrlf and core.eol, so a file checked out with CRLF line ends (Git for Windows' own default) and merely
//     touched reads as unchanged, as in the user's git: line ends alone hide no change worth a finding.
//
// It returns a note when the settings could not be read; git's defaults then stand for them.
func (g *gitRunner) userSettings() (note string) {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if runtime.GOOS == "windows" {
			k = strings.ToUpper(k)
		}
		if strings.HasPrefix(k, "GIT_") && !strings.HasPrefix(k, "GIT_CONFIG") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env, "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0", "LC_ALL=C")
	args := []string{"config", "-z", "--includes", "--path", "--get-regexp", `^core\.(excludesfile|autocrlf|eol)$`}
	out, err := g.exec(env, nil, append([]string{"--no-pager"}, args...), args)
	set := map[string]string{}
	switch {
	case err == nil:
		for _, e := range splitZ(out) {
			k, v, _ := strings.Cut(e, "\n")
			set[k] = v // the last value of a key is the one git uses
		}
	case exitCode(err) == 1: // none set: git's defaults
	default:
		note = "the user's git settings could not be read (" + err.Error() + "), so git's defaults stand for them"
	}
	if v := strings.ToLower(set["core.autocrlf"]); v != "" {
		if v == "input" || boolWord(v) {
			g.keep = append(g.keep, "core.autocrlf="+v)
		}
	}
	if v := strings.ToLower(set["core.eol"]); v == "lf" || v == "crlf" || v == "native" {
		g.keep = append(g.keep, "core.eol="+v)
	}
	file := set["core.excludesfile"]
	if file == "" {
		file = defaultExcludes()
	} else if !filepath.IsAbs(filepath.FromSlash(file)) {
		file = filepath.Join(g.dir, filepath.FromSlash(file))
	}
	if file != "" {
		if fi, err := os.Stat(filepath.FromSlash(file)); err == nil && !fi.IsDir() {
			g.excludes = filepath.ToSlash(file)
		}
	}
	return note
}

// boolWord reports whether git reads a value as a boolean.
func boolWord(v string) bool {
	switch v {
	case "true", "false", "yes", "no", "on", "off", "1", "0":
		return true
	}
	return false
}

// defaultExcludes is git's default global excludes file, as git finds its home folder (Git for Windows: HOME, else
// HOMEDRIVE and HOMEPATH when that folder exists, else USERPROFILE).
func defaultExcludes() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "git", "ignore")
	}
	home := os.Getenv("HOME")
	if home == "" && runtime.GOOS == "windows" {
		if d, p := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); d != "" && p != "" {
			if fi, err := os.Stat(d + p); err == nil && fi.IsDir() {
				home = d + p
			}
		}
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
	}
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "git", "ignore")
}

// splitZ splits NUL-terminated output.
func splitZ(out []byte) []string {
	s := string(out)
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\x00"), "\x00")
}

// firstLine is a text's first non-empty line, trimmed and made ASCII.
func firstLine(s string) string {
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			return asciiOnly(l)
		}
	}
	return ""
}

func oneLine(s string) string { return firstLine(s) }

// asciiOnly replaces every byte outside printable ASCII with ?.
func asciiOnly(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] < 0x20 || b[i] > 0x7e {
			b[i] = '?'
		}
	}
	return string(b)
}

// shown is a name as the reason prints it: as it is when it is printable ASCII, else quoted with Go's escapes, so
// the reason stays one line of ASCII whatever a file is called.
func shown(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] < 0x20 || name[i] > 0x7e {
			return fmt.Sprintf("%+q", name)
		}
	}
	return name
}
