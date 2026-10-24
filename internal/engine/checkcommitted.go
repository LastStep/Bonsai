package engine

// config-uncommitted (BUG-106, T-5736): the main checkout's bonsai.yaml as the ladder and the stop gate read it, at
// its HEAD (stopgate.CommittedConfig, the one reading, which cmd/bonsai hands in: stopgate imports engine, so engine
// cannot call it). While the repository holds no commit, or HEAD holds no bonsai.yaml, or the file differs from HEAD's,
// a climb refuses (config-uncommitted) and so does the stop gate (stop-config-uncommitted); check says so as a
// warning, never a finding, since a fresh link is uncommitted by nature. Like ComparePlugins it runs after Check, in
// bonsai check only: rung 0 judges a commit, and status lists findings alone.

import "github.com/LastStep/Bonsai/internal/workspace"

// Uncommitted is what the reading at HEAD found: Why, in words that end a sentence, "" when bonsai.yaml is committed
// as it stands (or the reading failed some other way, which the ladder reports itself); Never when HEAD holds no
// bonsai.yaml at all; Next the person's step for a file that differs (stopgate.UncommittedNext).
type Uncommitted struct {
	Why, Next string
	Never     bool
}

// UncommittedInitNext is the person's step while HEAD holds no bonsai.yaml: the first commit of what init wrote.
const UncommittedInitNext = "a person commits what bonsai init wrote, in the main checkout (git status lists it; then git add " +
	"and git commit)"

// CompareCommitted adds the config-uncommitted warning for u, when it found one.
func CompareCommitted(r *CheckResult, u Uncommitted) {
	if r == nil || r.Config == nil || u.Why == "" {
		return
	}
	next := u.Next
	if u.Never {
		next = UncommittedInitNext
	}
	r.add("config-uncommitted", workspace.ConfigFile, "", u.Why+": a climb and the stop gate refuse it (config-uncommitted) until a "+
		"person commits it", next)
}
