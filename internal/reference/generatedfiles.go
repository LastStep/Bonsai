package reference

import (
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
)

// GeneratedFilesPath is where the generated-files page lives, from the repository's root.
const GeneratedFilesPath = "docs/reference/generated-files.md"

// GeneratedFilesPage builds the generated-files page (plan-5 5.2.6a; spec section 6, "Generated files"): each kind of
// file Bonsai or an agent generates in a project, from format.GeneratedKinds, and the cleaning rules (5.2.6b) in plain
// words. Step 5.5 moves the words into base's generated-files skill and keeps the table generated from the same Go
// table. Same rules as Page: LF, ASCII, no map-order iteration.
func GeneratedFilesPage() []byte {
	var b strings.Builder
	line := func(f string, a ...any) { fmt.Fprintf(&b, f+"\n", a...) }
	line("# Reference: generated files")
	line("")
	line("What Bonsai and its agents generate inside a project, who writes each kind, how long it is kept, what is never")
	line("cleaned and when cleaning happens. Bonsai deletes only what this page says it may.")
	line("")
	line("**This page is generated. Never edit it by hand.** It is written by `go generate ./...` from the table")
	line("`format.GeneratedKinds` (internal/format/generated.go); a test rebuilds it and fails on any difference, naming that")
	line("command. Change a kind in the table, run `go generate ./...` from the repository's root, and commit the page with the")
	line("change. At step 5.5 the `generated-files` skill of the `base` pack takes these words, and its table stays generated")
	line("from the same Go table.")
	line("")
	line("## The kinds")
	line("")
	for _, k := range format.GeneratedKinds {
		line("### `%s`", k.Kind)
		line("")
		line("- What: %s.", k.What)
		line("- Where it lives: %s.", k.Where)
		line("- Written by: %s.", k.Writer)
		line("- Default: %s.", k.Default)
		line("- Never cleaned: %s.", k.Never)
		line("- Cleaned: %s.", k.When)
		if k.Rule {
			line("- Rule in bonsai.yaml: `generated.%s` takes `keep_days` and `keep_newest`.", k.Kind)
		} else {
			line("- Rule in bonsai.yaml: none; this kind takes no rule.")
		}
		line("")
	}
	line("## The rules")
	line("")
	line("1. **The protections come first.** A protected file or row is never cleaned, whatever the rule says, and it still")
	line("   counts among the newest. The protections are the \"Never cleaned\" line of each kind above.")
	line("2. **`keep_days`** cleans a file or row older than that many days. Age is a log or asks file's last record, a")
	line("   ladder result's `finished`, or a row's end.")
	line("3. **`keep_newest`** keeps only that many of the newest of its kind and cleans the rest.")
	line("4. Either one cleans. `null` for both keeps everything. A kind or the whole `generated:` section left out of")
	line("   bonsai.yaml takes the defaults above.")
	line("5. **Only Bonsai's own files** are cleaned, by the names it writes, regular files only, inside the kind's folder.")
	line("   Anything else in the folder is left alone.")
	line("6. **A `clean` record** goes in the log for every file or row cleaned, naming its path and the rule that cleaned it.")
	line("   A delete that finds the file gone writes nothing.")
	line("7. **Run reports are never deleted by Bonsai.** Past their rule, `bonsai check` lists them as a warning and a person")
	line("   deletes them.")
	return []byte(b.String())
}
