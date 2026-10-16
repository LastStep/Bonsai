package status

// bonsai status --line (spec §4, plan-5 5.6.4): the statusline's workspace half, one line of ASCII that a status-line
// command prints at every refresh. So it is cheap and it never breaks the line it sits in: no git process (the
// checkout is found by walking up to a .git, the main checkout and HEAD read from git's own files), no Claude Code, no
// network, no stdin; it reads bonsai.yaml, the task files, the asks folder and the active task's last ladder result.
// Outside a linked folder it is empty; a bonsai.yaml it cannot read gives LineUnreadable; nothing here fails.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/stopgate"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// LineUnreadable is the line of a folder whose bonsai.yaml Bonsai cannot read.
const LineUnreadable = "bonsai.yaml unreadable"

// Limits of the line: its length, and a task title's.
const (
	lineMax  = 100
	titleMax = 30
)

// Ladder states of bonsai.line/1's ladder.
const (
	LadderGreen = "green"
	LadderRed   = "red"
	LadderOld   = "old"
	LadderNone  = "none"
)

// BuildLine reads the line for the folder dir, with this process's environment (BONSAI_TASK, CLAUDE_PROJECT_DIR).
// It never fails: a folder that is not linked gives the empty line, an unreadable bonsai.yaml its sentence in error.
func BuildLine(dir string) *format.Line {
	l := &format.Line{}
	co, ok := findNoGit(dir)
	if !ok {
		return l
	}
	if _, err := os.Stat(filepath.Join(co.Root, workspace.ConfigFile)); err != nil {
		return l // not linked
	}
	cfg, err := workspace.LoadConfigFull(co.Root)
	if err != nil {
		e := engine.ConfigError(err)
		l.Line = LineUnreadable
		l.Error = &format.ErrorObject{Code: e.Code, Message: ASCIIText(e.What), Next: format.Next{Do: e.Next, Who: "person"}}
		return l
	}
	l.Workspace = &format.LineWorkspace{ID: cfg.ID, Name: cfg.Name}
	taskDir := cfg.Full.Documents.Task
	if co.Main != co.Root {
		if main, err := workspace.LoadConfigFull(co.Main); err == nil {
			taskDir = main.Full.Documents.Task
		}
	}
	counts := true
	if session := os.Getenv(sessionEnv); session != "" {
		s, ok := findNoGit(session)
		counts = ok && samePlace(s.Main, co.Main)
	}
	cache := loadLineCache(co.Main)
	// The task files are read once, for the active task and the verify count both (workspace.Active reads them as
	// well), and from the cache when they have not changed: 300 files are most of a status line's time. activeOf is
	// workspace.Active's answer from the summaries, which TestStatusLineActiveMatches holds to status.Active.
	var a active
	if tasks, err := cache.tasks(co.Main, taskDir); err != nil {
		a = active{why: "unreadable"}
	} else {
		a = activeOf(tasks, os.Getenv(workspace.TaskEnv), counts)
		for _, t := range tasks {
			if !t.Broken && t.Status == "verify" {
				l.Verify++
			}
		}
		if a.why == workspace.ReasonManyRunning {
			l.Running = &a.running
		}
	}
	if a.task != nil {
		l.Task = &format.LineTask{ID: ASCIIText(a.task.ID), Title: ASCIIText(a.task.Title), Status: ASCIIText(a.task.Status)}
		s := ladderState(co, a.task.ID)
		l.Ladder = &s
	}
	l.Asks = cache.openAsks(co.Main)
	cache.save()
	l.Line = lineText(l)
	return l
}

// active is the active task as the line needs it: the task, or why there is none (a workspace.Reason code) and,
// for two or more running, how many.
type active struct {
	task    *taskSum
	why     string
	running int64
}

// activeOf is the active task of contract §13 from the task files as summarized (workspace.Active's steps, with
// nothing named by --task): the task BONSAI_TASK names when the environment counts, else the one task reading running.
func activeOf(files []*taskSum, env string, counts bool) active {
	if counts && env != "" {
		var hits []*taskSum
		for _, f := range files {
			if f.ID == env {
				hits = append(hits, f)
			}
		}
		switch {
		case len(hits) == 0:
			return active{why: workspace.ReasonNamedMissing}
		case len(hits) > 1:
			return active{why: workspace.ReasonNamedDuplicate}
		case hits[0].Broken:
			return active{why: workspace.ReasonNamedBroken}
		}
		return active{task: hits[0]}
	}
	var running []*taskSum
	for _, f := range files {
		if f.Broken {
			return active{why: workspace.ReasonBrokenFile}
		}
		if f.Status == "running" {
			running = append(running, f)
		}
	}
	switch len(running) {
	case 0:
		return active{why: workspace.ReasonNoneRunning}
	case 1:
		return active{task: running[0]}
	}
	return active{why: workspace.ReasonManyRunning, running: int64(len(running))}
}

// lineText lays the parts out, in order, each left out when it says nothing, within lineMax characters.
func lineText(l *format.Line) string {
	var parts []string
	switch {
	case l.Task != nil:
		title := l.Task.Title
		if len(title) > titleMax {
			title = strings.TrimRight(title[:titleMax-3], " ") + "..."
		}
		parts = append(parts, strings.TrimSpace(l.Task.ID+" "+title))
		if l.Ladder != nil {
			parts = append(parts, "ladder:"+*l.Ladder)
		}
	case l.Running != nil:
		parts = append(parts, strconv.FormatInt(*l.Running, 10)+" running")
	default:
		parts = append(parts, "no task")
	}
	if l.Verify > 0 {
		parts = append(parts, "verify:"+strconv.FormatInt(l.Verify, 10))
	}
	if l.Asks > 0 {
		parts = append(parts, "asks:"+strconv.FormatInt(l.Asks, 10))
	}
	if l.Release != nil {
		parts = append(parts, "bonsai:v"+ASCIIText(*l.Release))
	}
	s := strings.Join(parts, " ")
	if len(s) > lineMax {
		s = s[:lineMax]
	}
	return s
}

// ladderState is where the task's last local result stands: green or red at this checkout's HEAD, old when it is for
// another commit (or HEAD cannot be read), none when there is no result for it.
func ladderState(co *workspace.Checkout, task string) string {
	raw, err := os.ReadFile(filepath.Join(co.Main, filepath.FromSlash(workspace.LocalDir+"/ladder/"+task+".json")))
	if err != nil {
		return LadderNone
	}
	r, err := format.ReadLadder(raw)
	if err != nil || r.Task == nil || *r.Task != task {
		return LadderNone
	}
	head, err := stopgate.Head(workspace.ReadGitLayout(co.Root))
	if err != nil || !strings.EqualFold(r.Git.SHA, head) {
		return LadderOld
	}
	if r.Green {
		return LadderGreen
	}
	return LadderRed
}

// findNoGit finds the checkout holding dir by walking up to the first .git (a folder, or a worktree's file), with no
// git process; its main checkout is the git files' (workspace.ReadGitLayout).
func findNoGit(dir string) (*workspace.Checkout, bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, false
	}
	for d := abs; ; {
		if _, err := os.Lstat(filepath.Join(d, ".git")); err == nil {
			return &workspace.Checkout{Root: d, Main: workspace.ReadGitLayout(d).Main}, true
		}
		up := filepath.Dir(d)
		if up == d {
			return nil, false
		}
		d = up
	}
}

// ASCIIText keeps the printable ASCII of text a file holds (a task's title), dropping control characters and anything
// else, so a title cannot send escape codes to a terminal; runs of spaces become one.
func ASCIIText(s string) string {
	var b strings.Builder
	space := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
			space = b.Len() > 0
		case c > ' ' && c < 0x7f:
			if space {
				b.WriteByte(' ')
				space = false
			}
			b.WriteByte(c)
		}
	}
	return b.String()
}

// The line's reads are cached in the home, cache/line/<machine key>.json, because a project keeps a task file for
// every task and a day file of asks for every day it has been used, and reading and checking all of them takes far
// longer than a status line may. The cache holds, for each task file and day file, its size and time and what was
// read from it; a file whose size and time are the same is not read again, so a refresh reads only what changed. It is
// refillable (a missing, unreadable or older cache is read again from the files), never a write in the project, and a
// write that fails (a read-only home) never fails the line.
const (
	lineCacheFolder = "cache/line"
	lineCacheFormat = 1
)

// settleAfter is how old a file's time must be, when the file is read, for the cache to trust a later look at it that
// finds the same size and time: file times are coarse (a few milliseconds on some disks), so a file written again
// within the same tick at the same size would otherwise read as unchanged. A newer file is read at every refresh.
const settleAfter = 2 * time.Second

// settled reports whether a file with this time may be trusted unchanged by the cache.
func settled(mod time.Time) bool { return time.Since(mod) > settleAfter }

// taskSum is what the line needs of one task file: its id (the file's, else its name's), status and title, and
// whether it does not parse; Path is its project-relative path, and Size and ModTime stamp the file as read.
type taskSum struct {
	Path    string
	Size    int64
	ModTime int64
	ID      string
	Status  string
	Title   string
	Broken  bool
	Settled bool // the file was older than settleAfter when it was read: a same-size, same-time file is the same file
}

// askDay is one ask day file as the cache holds it.
type askDay struct {
	Size    int64
	ModTime int64
	Settled bool
	Records [][2]string // each record: "f" (files the key, opens it) or "c" (resolves or answers it), then the key
}

type lineCache struct {
	Format  int
	Tasks   map[string]*taskSum
	Days    map[string]askDay
	path    string
	changed bool
}

// loadLineCache reads the cache of the project whose main checkout is main; an empty one when there is none to read.
func loadLineCache(main string) *lineCache {
	c := &lineCache{Format: lineCacheFormat, Tasks: map[string]*taskSum{}, Days: map[string]askDay{}}
	home, err := workspace.Home()
	if err != nil {
		return c
	}
	key, err := workspace.MachineKey(main)
	if err != nil {
		return c
	}
	c.path = filepath.Join(home, filepath.FromSlash(lineCacheFolder), key+".json")
	raw, err := os.ReadFile(c.path)
	if err != nil {
		c.changed = true
		return c
	}
	var old lineCache
	if json.Unmarshal(raw, &old) != nil || old.Format != lineCacheFormat || old.Tasks == nil || old.Days == nil {
		c.changed = true
		return c
	}
	c.Tasks, c.Days = old.Tasks, old.Days
	return c
}

// save writes the cache when it changed. A failed write is no failure.
func (c *lineCache) save() {
	if !c.changed || c.path == "" {
		return
	}
	if raw, err := json.Marshal(c); err == nil {
		_ = workspace.WriteFileAtomic(c.path, raw)
	}
}

var taskIDRe = sync.OnceValue(func() *regexp.Regexp { return regexp.MustCompile(workspace.TaskIDPattern) })

// tasks summarizes the task files of the main checkout's task folder, in path order, as workspace.ReadTaskFiles reads
// them (the one reader of the task folder): its error is a folder or file that cannot be read at all.
func (c *lineCache) tasks(main, dir string) ([]*taskSum, error) {
	paths, err := workspace.DocFiles(main, workspace.DocKind{Kind: "task", Path: dir, ID: workspace.TaskIDPattern})
	if err != nil {
		return nil, err
	}
	out := make([]*taskSum, 0, len(paths))
	seen := make(map[string]*taskSum, len(paths))
	for _, p := range paths {
		full := filepath.Join(main, filepath.FromSlash(p))
		info, err := os.Stat(full)
		if err != nil {
			return nil, err
		}
		t := c.Tasks[p]
		if t == nil || !t.Settled || t.Size != info.Size() || t.ModTime != info.ModTime().UnixNano() {
			raw, err := os.ReadFile(full)
			if err != nil {
				return nil, err
			}
			t = &taskSum{Path: p, Size: info.Size(), ModTime: info.ModTime().UnixNano(), Settled: settled(info.ModTime()), ID: workspace.NameID(filepath.Base(p), taskIDRe())}
			if task, err := workspace.ReadTask(raw); err != nil {
				t.Broken = true
			} else {
				t.Status, t.Title = task.Status, task.Title
				if task.ID != "" {
					t.ID = task.ID
				}
			}
			c.changed = true
		}
		seen[p] = t
		out = append(out, t)
	}
	if len(seen) != len(c.Tasks) {
		c.changed = true
	}
	c.Tasks = seen
	return out, nil
}

// openAsks counts the keys whose latest record opens them (bonsai.asks/1's open ones): the same count as
// asks.Read's Entries(false), which TestStatusLineAsksMatch holds it to. A folder that cannot be read counts none.
func (c *lineCache) openAsks(main string) int64 {
	dir := asks.Folder(main)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	days := make(map[string]askDay, len(entries))
	state := map[string]bool{}   // key -> open
	for _, de := range entries { // sorted by name: the days in order
		name := de.Name()
		if !de.Type().IsRegular() || !isDayFile(name) {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		day, ok := c.Days[name]
		if !ok || !day.Settled || day.Size != info.Size() || day.ModTime != info.ModTime().UnixNano() {
			recs, err := readAskDay(filepath.Join(dir, name))
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			if err != nil {
				return 0
			}
			day = askDay{Size: info.Size(), ModTime: info.ModTime().UnixNano(), Settled: settled(info.ModTime()), Records: recs}
			c.changed = true
		}
		days[name] = day
		for _, r := range day.Records {
			if r[0] == "f" {
				state[r[1]] = true
			} else if state[r[1]] {
				state[r[1]] = false // the first close stands
			}
		}
	}
	if len(days) != len(c.Days) {
		c.changed = true
	}
	c.Days = days
	n := int64(0)
	for _, open := range state {
		if open {
			n++
		}
	}
	return n
}

// isDayFile is how an asks file is named: its UTC day, YYYY-MM-DD.ndjson (contract §9.1).
func isDayFile(name string) bool {
	if len(name) != len("2006-01-02.ndjson") || !strings.HasSuffix(name, ".ndjson") {
		return false
	}
	for i := 0; i < 10; i++ {
		c := name[i]
		if i == 4 || i == 7 {
			if c != '-' {
				return false
			}
		} else if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

var askRecordStart = []byte(`{"format":"bonsai.ask/`)

// readAskDay reads one day file's records as asks.Read does: a line that is a bonsai.ask/1 record counts, one with
// a torn start counts from where a record begins, anything else is passed over.
func readAskDay(path string) ([][2]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var out [][2]string
	r := bufio.NewReaderSize(f, 64<<10)
	for {
		line, err := r.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			line = bytes.TrimSuffix(bytes.TrimSuffix(line, []byte("\n")), []byte("\r"))
			rec := readAskRecord(line)
			if rec == nil {
				if i := bytes.LastIndex(line, askRecordStart); i > 0 {
					rec = readAskRecord(line[i:])
				}
			}
			if rec != nil {
				switch rec.Op {
				case asks.OpFile:
					out = append(out, [2]string{"f", rec.Key})
				case asks.OpResolve, asks.OpAnswer:
					out = append(out, [2]string{"c", rec.Key})
				}
			}
		}
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func readAskRecord(line []byte) *format.Ask {
	f := format.MustLookup("ask")
	doc, err := f.ReadJSON(line)
	if err != nil {
		return nil
	}
	a := &format.Ask{}
	if err := f.Bind(doc, a); err != nil {
		return nil
	}
	return a
}
