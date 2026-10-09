package format

// The generated kinds (spec §6, "Generated files"; format review R2.6): every kind of file or row Bonsai or an agent
// generates in a project, where it lives, how long it is kept by default, and what is never cleaned. GeneratedKinds
// is the one home of their defaults and their protections: bonsai init writes the defaults into bonsai.yaml's
// generated: section, the cleaner (step 5.2.6b) and check's warning on run reports read them, and base's
// generated-files skill and the reference page of lists (step 5.1.10) are generated from it. The kinds that take a
// rule are bonsai.workspace/1's generated properties, in the same order (TestGeneratedKindsAreTheSchemas holds the
// two equal: a new kind is an addition to the schema, a set change, and an entry here in the same commit).

// GeneratedKind is one kind of generated file.
type GeneratedKind struct {
	Kind       string // its name: a key of bonsai.yaml's generated:, when it takes a rule
	Where      string // where it lives
	What       string // one line: what it is
	KeepDays   *int64 // the default keep_days, nil for none (no age rule)
	KeepNewest *int64 // the default keep_newest, nil for none
	Default    string // the default, in words
	Never      string // what is never cleaned, whatever the rule
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
		Default: "30 days after a file's last line",
		Never:   "an open session's file, or one holding an ended session or subagent run with no row yet in .bonsai/sessions.md", Rule: true},
	{Kind: "asks", Where: ".bonsai/local/asks/", What: "questions for a person and their answers", Default: "kept",
		Never: "a day file holding an open ask", Rule: true},
	{Kind: "ladder", Where: ".bonsai/local/ladder/", What: "ladder results, one per task", KeepDays: days(7),
		Default: "7 days after the result's finished", Never: "the result of a task that is not done or cut", Rule: true},
	{Kind: "run", Where: "the run reports' folder (documents.run)", What: "run reports, committed history", Default: "kept",
		Never: "any, by Bonsai: past a rule, bonsai check lists them and a person deletes them", Rule: true},
	{Kind: "sessions", Where: ".bonsai/sessions.md", What: "rows of the sessions table", Default: "kept",
		Never: "rows of an open task", Rule: true},
	{Kind: "tasks", Where: ".bonsai/tasks.md", What: "the tasks table", Default: "a rebuild: nothing to clean",
		Never: "-", Rule: false},
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
