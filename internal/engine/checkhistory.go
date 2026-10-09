package engine

// approve_first from git history (contract §6): "a move to running, verify or done needs the task to have read
// approved since it last read todo or plan", and bonsai check checks it in every workspace from git history; "in a
// checkout without full history (a shallow CI clone) bonsai check says 'history not available' for that rule, never
// 'passed'".
//
// Which tasks: those in this checkout's task folder that read running, verify or done and are in a lane with
// approve_first, or in a lane no pack defines (contract §6: an unknown lane counts as the strictest). A workspace whose
// packs define no lane has no rule to check, so git is not asked at all. The rule is judged on each task's present
// status: after the last time its file read todo or plan (or since it was first committed), the first status of
// running, verify or done must come after one of approved. A task moved on without approval stays found until a
// person moves it back to plan and approves it, since history is never rewritten.
//
// How the history is read: one git log over the task folder, oldest first, each commit's patch with no context
// lines (-U0) and renames followed (-M), so every status: line a commit added to a task file is seen in order; a
// change in the working tree not yet committed is the newest status. A shallow clone, or a git that fails, gives one
// approve-first-unchecked warning naming the tasks it could not judge.

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// approveStatuses are the statuses a move to which needs approval first (contract §6).
var approveStatuses = []string{"running", "verify", "done"}

// judgedTask is a task the approve_first rule applies to.
type judgedTask struct {
	path, id, status, lane string
}

// checkHistory checks approve_first from git history.
func (r *CheckResult) checkHistory() {
	lanes, err := workspace.Lanes(r.Lock)
	if err != nil || len(lanes) == 0 || r.Config.Full == nil {
		return
	}
	approve := map[string]bool{}
	for _, l := range lanes {
		approve[l.Name] = l.ApproveFirst
	}
	dir := strings.TrimSuffix(r.Config.Full.Documents.Task, "/")
	files, err := workspace.DocFiles(r.Root, workspace.DocKind{Kind: "task", Path: dir, ID: workspace.TaskIDPattern})
	if err != nil || dir == "" {
		return
	}
	var judged []judgedTask
	for _, p := range files {
		raw, exists, _ := readFile(r.Root, p)
		if !exists {
			continue
		}
		t, err := format.ReadTask(raw)
		if err != nil || t.Lane == nil || *t.Lane == "" || !contains(approveStatuses, t.Status) {
			continue
		}
		if af, known := approve[*t.Lane]; known && !af {
			continue
		}
		judged = append(judged, judgedTask{path: p, id: t.ID, status: t.Status, lane: *t.Lane})
	}
	if len(judged) == 0 {
		return
	}
	var ids []string
	for _, j := range judged {
		ids = append(ids, j.id)
	}
	unchecked := func(why, next string) {
		r.add("approve-first-unchecked", "", "", "history not available: approve_first was not checked for "+strings.Join(ids, ", ")+
			" ("+why+"); this is not a pass", next)
	}
	if shallow, err := gitOut(r.Root, "rev-parse", "--is-shallow-repository"); err != nil {
		unchecked("git rev-parse failed", "check that git works in this checkout, run: git status")
		return
	} else if strings.TrimSpace(shallow) == "true" {
		unchecked("a shallow clone", "fetch the whole history, run: git fetch --unshallow; then run: bonsai check")
		return
	}
	hist, err := statusHistory(r.Root, dir)
	if err != nil {
		unchecked("git log failed: "+oneLine(err.Error()), "check that git works in this checkout, run: git status")
		return
	}
	for _, j := range judged {
		seq := hist[j.path]
		if len(seq) == 0 || seq[len(seq)-1] != j.status {
			seq = append(seq, j.status)
		}
		if ok, since := approvedFirst(seq); !ok {
			lane := "the lane " + j.lane + ", whose approve_first needs it"
			if _, known := approve[j.lane]; !known {
				lane = "the lane " + j.lane + ", which no pack defines (the strictest: approve_first)"
			}
			r.add("approve-first", j.path, "", j.id+" ("+j.path+") reads "+j.status+" in "+lane+", but git history shows no approved "+since,
				"a person approves the task: move it back to plan, then to approved, then on, each move a commit of its own; "+
					"or, if it needs no approval, a person gives it a lane without approve_first")
		}
	}
}

// approvedFirst reports whether, after the last todo or plan in a task's statuses (oldest first), approved comes
// before the first running, verify or done; since says from when, for a person.
func approvedFirst(seq []string) (bool, string) {
	start, since := 0, "since it was first committed"
	for i, s := range seq {
		if s == "todo" || s == "plan" {
			start, since = i+1, "since it last read "+s
		}
	}
	for _, s := range seq[start:] {
		switch {
		case s == "approved":
			return true, since
		case contains(approveStatuses, s):
			return false, since
		}
	}
	return true, since
}

// statusHistory reads, from one git log over the task folder (oldest first), each task file's status: values in the
// order its commits wrote them, following renames.
func statusHistory(root, dir string) (map[string][]string, error) {
	out, err := gitOut(root, "-c", "core.quotepath=off", "log", "--reverse", "--format=%x01%H", "-p", "-U0", "-M",
		"--no-color", "--no-ext-diff", "--", dir)
	if err != nil {
		return nil, err
	}
	hist := map[string][]string{}
	cur := ""
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "\x01"), strings.HasPrefix(line, "diff --git "):
			cur = ""
		case strings.HasPrefix(line, "rename from "):
			cur = strings.TrimPrefix(line, "rename from ")
		case strings.HasPrefix(line, "rename to "):
			to := strings.TrimPrefix(line, "rename to ")
			if cur != "" {
				hist[to] = append(hist[to], hist[cur]...)
				delete(hist, cur)
			}
			cur = to
		case strings.HasPrefix(line, "+++ "):
			cur = ""
			if p := strings.TrimPrefix(line, "+++ "); strings.HasPrefix(p, "b/") {
				cur = strings.TrimPrefix(p, "b/")
			}
		case strings.HasPrefix(line, "+status:") && cur != "":
			if s := statusValue(strings.TrimPrefix(line, "+status:")); s != "" {
				hist[cur] = append(hist[cur], s)
			}
		}
	}
	return hist, sc.Err()
}

// statusValue reads a status: line's value: a word, quoted or not, with any # comment after it.
func statusValue(v string) string {
	v = strings.TrimSpace(v)
	if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return strings.Trim(v, `"'`)
}

// gitOut runs git in a checkout, with Bonsai's git environment, and gives its output.
func gitOut(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = gitEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", &gitError{args: args, stderr: firstLine(stderr.String()), err: err}
	}
	return strings.ReplaceAll(stdout.String(), "\r\n", "\n"), nil
}
