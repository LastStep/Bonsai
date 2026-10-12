package ratchet

// The Bless offer (design/plan-5.md, 5.4.4 note 2), walked from a table with each condition false in turn: not the
// task's result, another workspace's, not local, red, dirty, no base branch, not the base branch, not the main
// checkout, the main checkout on another branch, not at its HEAD, no rise, a rise on a red rung, a name not a key of
// ratchets; and the offer's rises, words and data.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

const (
	testWS   = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"
	headA    = "5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f80"
	headB    = "0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c"
	resultSH = "0f1e2d3c4b5a69788796a5b4c3d2e1f00f1e2d3c4b5a69788796a5b4c3d2e1f0"
)

// rung is a result's rung with a ratchet (name, was, now), its status given.
func rung(n int64, status, name, was, now string) format.LadderRung {
	r := format.LadderRung{Rung: n, Name: "tests", Kind: "command", Status: status}
	if name != "" {
		r.Ratchet = &format.LadderRatchet{Name: name, Was: json.Number(was), Now: json.Number(now), OK: true}
	}
	return r
}

// greenResult is a green, clean local result of task on main at headA, with the rungs given.
func greenResult(task string, rungs ...format.LadderRung) *format.Ladder {
	return &format.Ladder{Task: &task, Workspace: testWS, Mode: "local", Started: "2026-10-13T10:00:00.000Z",
		Finished: "2026-10-13T10:01:00.000Z", Git: format.LadderGit{SHA: headA, Branch: "main"}, Requested: []int64{1}, Green: true,
		Rungs: rungs, Skipped: []format.LadderSkip{}}
}

// floors are bonsai.yaml's ratchets.
func floors(kv ...string) schema.Object {
	o := schema.Object{}
	for i := 0; i+1 < len(kv); i += 2 {
		o = append(o, schema.Member{Key: kv[i], Value: json.Number(kv[i+1])})
	}
	return o
}

func here() Where {
	return Where{Workspace: testWS, Base: "main", Branch: "main", Head: headA, Floors: floors("go_tests", "812", "go_vet", "3")}
}

// Each condition of the offer false in turn: no offer, and why names that condition.
func TestOfferConditions(t *testing.T) {
	rising := rung(1, "green", "go_tests", "812", "815")
	for _, c := range []struct {
		name string
		res  func(*Result)
		at   func(*Where)
		why  string
	}{
		{"the offer", nil, nil, "offered"},
		{"no result", func(r *Result) { r.Ladder = nil }, nil, "there is no ladder result for T-0901"},
		{"no result, with its reason", func(r *Result) { r.Ladder, r.Missing = nil, "x.json is not a ladder result" }, nil, "x.json is not a ladder result"},
		{"another task's result", func(r *Result) { other := "T-0902"; r.Ladder.Task = &other }, nil, `the result for T-0901 names "T-0902"`},
		{"a result with no task", func(r *Result) { r.Ladder.Task = nil }, nil, "the result for T-0901 names no task"},
		{"another workspace's", func(r *Result) { r.Ladder.Workspace = "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb" }, nil, "is another workspace's"},
		{"not local", func(r *Result) { r.Ladder.Mode = "ci" }, nil, `was a "ci" climb, which is never a task's proof`},
		{"red", func(r *Result) { r.Ladder.Green = false }, nil, "the last climb for T-0901 was not green"},
		{"dirty", func(r *Result) { r.Ladder.Git.Dirty = true }, nil, "the last climb for T-0901 was dirty"},
		{"no base branch", nil, func(w *Where) { w.Base = "" }, "the main checkout's HEAD is detached, so there is no base branch"},
		{"not the base branch", func(r *Result) { r.Ladder.Git.Branch = "t0901" }, nil, `was on "t0901", not the base branch main`},
		{"not the main checkout", func(r *Result) { r.Elsewhere = true }, nil, "ran in a worktree, not the main checkout"},
		{"the main checkout moved off the base", nil, func(w *Where) { w.Branch = "other" }, "the main checkout is no longer on main"},
		{"not at its HEAD", nil, func(w *Where) { w.Head = headB }, "was made at 5d6e7f80, and HEAD is 0f1e2d3c (HEAD moved)"},
		{"no HEAD read", nil, func(w *Where) { w.Head = "" }, "and HEAD is (no commit) (HEAD moved)"},
		{"no rise", func(r *Result) { r.Ladder.Rungs = []format.LadderRung{rung(1, "green", "go_tests", "812", "812")} }, nil,
			"measured no ratchet above its floor in bonsai.yaml's ratchets"},
		{"a rise on a red rung", func(r *Result) { r.Ladder.Rungs = []format.LadderRung{rung(1, "red", "go_tests", "812", "815")} }, nil,
			"measured no ratchet above its floor"},
		{"a name not a key of ratchets", func(r *Result) { r.Ladder.Rungs = []format.LadderRung{rung(1, "green", "other", "0", "9")} }, nil,
			"measured no ratchet above its floor"},
		{"a floor that is no number", nil, func(w *Where) { w.Floors = schema.Object{{Key: "go_tests", Value: "many"}} },
			"measured no ratchet above its floor"},
		{"a floor raised since", nil, func(w *Where) { w.Floors = floors("go_tests", "815") }, "measured no ratchet above its floor"},
		{"a rung with no ratchet", func(r *Result) { r.Ladder.Rungs = []format.LadderRung{rung(1, "green", "", "", "")} }, nil,
			"measured no ratchet above its floor"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := Result{Ladder: greenResult("T-0901", rising), SHA256: resultSH}
			w := here()
			if c.res != nil {
				c.res(&r)
			}
			if c.at != nil {
				c.at(&w)
			}
			o, why := OfferFor("T-0901", r, w)
			if why != c.why && !strings.Contains(why, c.why) {
				t.Errorf("why %q, want %q", why, c.why)
			}
			if (o != nil) != (c.why == "offered") {
				t.Errorf("offer %+v with why %q", o, why)
			}
		})
	}
}

// The offer's rises: each green rung's ratchet above the floor bonsai.yaml now holds, its name a key of ratchets, once
// each, in rung order, at most MaxRises; was is the floor now, now the result's count, new_tests the result's.
func TestOfferRises(t *testing.T) {
	nt := &format.LadderNewTests{Base: headB, Count: 3, FailedOnBase: 2, PassedOnBase: 1}
	withNT := rung(4, "green", "go_vet", "0", "5")
	withNT.Ratchet.NewTests = nt
	l := greenResult("T-0901", rung(1, "green", "go_tests", "800", "815"), rung(2, "green", "other", "0", "9"),
		rung(3, "green", "go_tests", "812", "820"), withNT, rung(5, "green", "Bad-Name", "0", "9"))
	l.Git.SHA = strings.ToUpper(headA)
	o, why := OfferFor("T-0901", Result{Ladder: l, SHA256: strings.ToUpper(resultSH)}, here())
	if o == nil {
		t.Fatal(why)
	}
	if o.Task != "T-0901" || o.Head != headA || o.ResultSHA != resultSH || o.Base != "main" || o.More != 0 || len(o.Rises) != 2 {
		t.Fatalf("%+v", o)
	}
	if r := o.Rises[0]; r.Name != "go_tests" || r.Was != "812" || r.Now != "815" || r.NewTests != nil {
		t.Errorf("the first rise (was is the floor now; a second rung of one name is not offered again): %+v", r)
	}
	if r := o.Rises[1]; r.Name != "go_vet" || r.Was != "3" || r.Now != "5" || r.NewTests != nt {
		t.Errorf("the second rise: %+v", r)
	}
	// More than MaxRises: the first MaxRises carried, the rest counted.
	var rungs []format.LadderRung
	fl := schema.Object{}
	for i := 0; i < MaxRises+2; i++ {
		name := "count_" + string(rune('a'+i))
		rungs = append(rungs, rung(int64(i), "green", name, "1", "2"))
		fl = append(fl, schema.Member{Key: name, Value: int64(1)})
	}
	w := here()
	w.Floors = fl
	o, _ = OfferFor("T-0901", Result{Ladder: greenResult("T-0901", rungs...), SHA256: resultSH}, w)
	if o == nil || len(o.Rises) != MaxRises || o.More != 2 || o.Rises[0].Was != "1" {
		t.Fatalf("%+v", o)
	}
	if why := Why(o); !strings.Contains(why, "; and 2 more rises, beyond the 10 an ask carries.") {
		t.Errorf("why %q", why)
	}
}

// The ask's words: the title, the then, and a why naming each rise and its new-tests numbers, never past WhyMax; the
// data, contract §9.1's shape.
func TestAskWords(t *testing.T) {
	o := &Offer{Task: "T-0901", Head: headA, ResultSHA: resultSH, Base: "main", Rises: []Rise{
		{Name: "go_tests", Was: "812", Now: "815", NewTests: &format.LadderNewTests{Base: headB, Count: 3, FailedOnBase: 2, PassedOnBase: 1}},
		{Name: "go_tests_tagged", Was: "812", Now: "815"},
		{Name: "found", Was: "1", Now: "2", NewTests: &format.LadderNewTests{Count: 2, NotRun: 2}},
		{Name: "none_new", Was: "1", Now: "2", NewTests: &format.LadderNewTests{Base: headB}},
	}}
	if Title("T-0901") != "Bless T-0901's ratchets" || Then != "Blessing raises the floor in bonsai.yaml from that run; it runs nothing." {
		t.Errorf("title %q, then %q", Title("T-0901"), Then)
	}
	want := "A green, clean climb of main at 5d6e7f80: go_tests rose from 812 to 815, 3 new tests: 2 failed on the base, 1 passed there " +
		"(they prove nothing new), 0 not run; go_tests_tagged rose from 812 to 815, its new tests not checked on the base; found rose " +
		"from 1 to 2, 2 new tests: 0 failed on the base, 0 passed there (they prove nothing new), 2 not run (no base was found); " +
		"none_new rose from 1 to 2, no new test named."
	if got := Why(o); got != want {
		t.Errorf("why\n%s\nwant\n%s", got, want)
	}
	wantData := `{"task":"T-0901","head":"` + headA + `","result_sha":"` + resultSH + `","rises":[` +
		`{"name":"go_tests","was":812,"now":815,"new_tests":{"base":"` + headB + `","count":3,"failed_on_base":2,"passed_on_base":1,"not_run":0}},` +
		`{"name":"go_tests_tagged","was":812,"now":815,"new_tests":null},` +
		`{"name":"found","was":1,"now":2,"new_tests":{"base":"","count":2,"failed_on_base":0,"passed_on_base":0,"not_run":2}},` +
		`{"name":"none_new","was":1,"now":2,"new_tests":{"base":"` + headB + `","count":0,"failed_on_base":0,"passed_on_base":0,"not_run":0}}]}`
	if got := schema.Show(Data(o)); got != wantData {
		t.Errorf("data\n%s\nwant\n%s", got, wantData)
	}
	// Ten rises with long names and their numbers: the why keeps what fits in WhyMax and counts the rest.
	long := &Offer{Task: "T-0901", Head: headA, Base: "main"}
	for i := 0; i < MaxRises; i++ {
		long.Rises = append(long.Rises, Rise{Name: strings.Repeat(string(rune('a'+i)), 40), Was: "100000", Now: "100001",
			NewTests: &format.LadderNewTests{Base: headB, Count: 1, FailedOnBase: 1}})
	}
	got := Why(long)
	if n := len([]rune(got)); n > WhyMax || !strings.Contains(got, " more rises in the ask's data.") || !strings.HasPrefix(got, "A green, clean climb") {
		t.Errorf("%d characters: %s", n, got)
	}
	// A base branch's long name is shown by its first 100 characters, so the why still fits.
	huge := &Offer{Task: "T-0901", Head: headA, Base: strings.Repeat("b", WhyMax), Rises: []Rise{{Name: "go_tests", Was: "1", Now: "2"}}}
	if got := Why(huge); len([]rune(got)) > WhyMax || !strings.Contains(got, strings.Repeat("b", 100)+"... at 5d6e7f80: go_tests rose") {
		t.Errorf("%s", got)
	}
	// Rises too long for any why: the why says how many rose and where they are.
	wide := &Offer{Task: "T-0901", Head: headA, Base: "main", Rises: []Rise{{Name: "go_tests", Was: json.Number(strings.Repeat("9", WhyMax)), Now: "1"}}}
	if got := Why(wide); got != "A green, clean climb of main at 5d6e7f80: 1 ratchet rose, each in the ask's data." {
		t.Errorf("%s", got)
	}
}

// The climb's own new-tests lines (Result.Checks, one per rung): each rise carries its rung's line when the rung has
// new_tests, and the why gives it in place of the counts' words; a result with no lines (read back from its file)
// keeps the counts' words; a line too long for the why gives way to the counts' words for every rise.
func TestAskWordsChecked(t *testing.T) {
	nt := &format.LadderNewTests{Base: headB, Count: 3, FailedOnBase: 1, PassedOnBase: 1, NotRun: 1}
	first := rung(1, "green", "go_tests", "800", "815")
	first.Ratchet.NewTests = nt
	second := rung(2, "green", "go_vet", "0", "5") // no new_tests: its line, if any, is not carried
	third := rung(3, "green", "found", "0", "9")
	third.Ratchet.NewTests = &format.LadderNewTests{Count: 9, NotRun: 9}
	w := here()
	w.Floors = floors("go_tests", "812", "go_vet", "3", "found", "1")
	line := "3 new tests, run on the base 0f1e2d3c4b5a: 1 failed there, 1 proves nothing new, 1 check not run"
	noBase := "9 new tests (the count's rise; they could not be named): check not run: no base"
	res := Result{Ladder: greenResult("T-0901", first, second, third), SHA256: resultSH, Checks: []string{line, "stray", noBase}}
	o, why := OfferFor("T-0901", res, w)
	if o == nil || len(o.Rises) != 3 {
		t.Fatalf("%v: %+v", why, o)
	}
	if o.Rises[0].Checked != line || o.Rises[1].Checked != "" || o.Rises[2].Checked != noBase {
		t.Errorf("the rises' lines: %q, %q, %q", o.Rises[0].Checked, o.Rises[1].Checked, o.Rises[2].Checked)
	}
	want := "A green, clean climb of main at 5d6e7f80: go_tests rose from 812 to 815, " + line + "; go_vet rose from 3 to 5, its new " +
		"tests not checked on the base; found rose from 1 to 9, " + noBase + "."
	if got := Why(o); got != want {
		t.Errorf("why\n%s\nwant\n%s", got, want)
	}
	// Read back from its file, the same result has no lines: the counts' words.
	res.Checks = nil
	o, _ = OfferFor("T-0901", res, w)
	if got := Why(o); !strings.Contains(got, "go_tests rose from 812 to 815, 3 new tests: 1 failed on the base, 1 passed there "+
		"(they prove nothing new), 1 not run;") || strings.Contains(got, "check not run") {
		t.Errorf("read back: %s", got)
	}
	// A line too long for the why: every rise in the counts' words, all of them kept.
	res.Checks = []string{"3 new tests: check not run: " + strings.Repeat("x", WhyMax), "", noBase}
	o, _ = OfferFor("T-0901", res, w)
	got := Why(o)
	if len([]rune(got)) > WhyMax || strings.Contains(got, "xxx") || strings.Contains(got, "could not be named") ||
		!strings.Contains(got, "found rose from 1 to 9, 9 new tests: 0 failed on the base, 0 passed there (they prove nothing new), 9 not run "+
			"(no base was found).") {
		t.Errorf("too long: %s", got)
	}
}
