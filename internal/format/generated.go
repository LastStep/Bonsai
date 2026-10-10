package format

// The generated kinds (spec §6, "Generated files"; format review R2.6): every kind of file or row Bonsai or an agent
// generates in a project, where it lives, how long it is kept by default, and what is never cleaned. GeneratedKinds
// is the one home of their defaults and their protections: bonsai init writes the defaults into bonsai.yaml's
// generated: section, the cleaner (internal/clean, step 5.2.6b) and check's warning on run reports read them, and base's
// generated-files skill (until step 5.5 its page, docs/reference/generated-files.md, step 5.2.6a) and the reference page of lists (step 5.1.10) are generated from it. The kinds that take a
// rule are bonsai.workspace/1's generated properties, in the same order (TestGeneratedKindsAreTheSchemas holds the
// two equal: a new kind is an addition to the schema, a set change, and an entry here in the same commit).

// GeneratedKind is one kind of generated file.
type GeneratedKind struct {
	Kind       string // its name: a key of bonsai.yaml's generated:, when it takes a rule
	Where      string // where it lives
	What       string // one line: what it is
	Writer     string // who writes it
	KeepDays   *int64 // the default keep_days, nil for none (no age rule)
	KeepNewest *int64 // the default keep_newest, nil for none
	Default    string // the default, in words
	Never      string // what is never cleaned, whatever the rule
	When       string // when it is cleaned (step 5.2.6b), "never, by Bonsai" where so
	Rule       bool   // it takes a rule in bonsai.yaml (false for the tasks table, rebuilt whole)
}

func days(n int64) *int64 { return &n }

func copyInt(p *int64) *int64 {
	if p == nil {
		return nil
	}
	return days(*p)
}

// GeneratedKinds are the generated kinds, in spec §6's order.
var GeneratedKinds = []GeneratedKind{
	{Kind: "log", Where: ".bonsai/local/log/", What: "the log, one file per session", KeepDays: days(30),
		Writer:  "the recorder, from the hooks, one line at a time",
		Default: "30 days after a file's last line",
		Never:   "an open session's file, or one holding an ended session or subagent run with no row yet in .bonsai/sessions.md",
		When:    "at a session's end, after its session_end line, within a budget of 1 second, oldest first; what the budget leaves waits for the next end", Rule: true},
	{Kind: "asks", Where: ".bonsai/local/asks/", What: "questions for a person and their answers",
		Writer: "bonsai ask and bonsai answer (and the ladder runner from step 5.4, as its Bless), one day file a day", Default: "kept",
		Never: "a day file holding an open ask, or the answer or withdrawal of an ask whose filing stays (else it would read open again)",
		When:  "at a session's end, within the same budget; by default nothing, as the default keeps", Rule: true},
	{Kind: "ladder", Where: ".bonsai/local/ladder/", What: "ladder results, one per task", KeepDays: days(7),
		Writer:  "the ladder runner (step 5.4), one file per task",
		Default: "7 days after the result's finished",
		Never:   "the result of a task that is not done or cut (a task not found, or whose file does not read, counts as not done)",
		When:    "at a session's end, and by the ladder runner after each run", Rule: true},
	{Kind: "run", Where: "the run reports' folder (documents.run)", What: "run reports, committed history",
		Writer: "the agent that did the work, in the run report format", Default: "kept",
		Never: "any, by Bonsai: past a rule, bonsai check lists them and a person deletes them",
		When:  "never, by Bonsai", Rule: true},
	{Kind: "sessions", Where: ".bonsai/sessions.md", What: "rows of the sessions table",
		Writer: "bonsai check --write, from the log", Default: "kept",
		Never: "rows of a task that is not done or cut (none is never protected), and rows whose span is still in the log (the next write would add them again)",
		When:  "in bonsai check --write, the only writer of the table", Rule: true},
	{Kind: "tasks", Where: ".bonsai/tasks.md", What: "the tasks table",
		Writer: "bonsai check --write, from the task files", Default: "a rebuild: nothing to clean",
		Never: "-", When: "never: the table is rebuilt whole", Rule: false},
}

// DefaultGenerated is bonsai.yaml's generated: section as init writes it: each kind's default rule.
func DefaultGenerated() Generated {
	g := Generated{}
	for _, k := range GeneratedKinds {
		keep := Keep{KeepDays: copyInt(k.KeepDays), KeepNewest: copyInt(k.KeepNewest)}
		switch k.Kind {
		case "log":
			g.Log = keep
		case "asks":
			g.Asks = keep
		case "ladder":
			g.Ladder = keep
		case "run":
			g.Run = keep
		case "sessions":
			g.Sessions = keep
		}
	}
	return g
}

// GeneratedKindOf finds a generated kind by its name.
func GeneratedKindOf(kind string) (GeneratedKind, bool) {
	for _, k := range GeneratedKinds {
		if k.Kind == kind {
			return k, true
		}
	}
	return GeneratedKind{}, false
}
