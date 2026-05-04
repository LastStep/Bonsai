package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
	"github.com/LastStep/Bonsai/internal/validate"
)

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringP("agent", "a", "", "Restrict check to a single installed agent")
	validateCmd.Flags().Bool("json", false, "Emit issues as JSON (one Report object) — non-interactive")
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Audit ability state — detect orphaned registrations, stale lock entries, untracked custom files, and frontmatter problems.",
	Long: "Read-only audit. Checks .bonsai.yaml + .bonsai-lock.yaml + agent/ workspace " +
		"for inconsistencies. Exit 0 = clean, 1 = issues found, 2 = internal error. " +
		"Run `bonsai update` to fix orphans/untracked. JSON output via --json.",
	RunE: runValidate,
}

// runValidate is the validate command entry point.
//
// Exit-code contract (drives os.Exit calls — Cobra's RunE-returns-nil
// path is reserved for the clean case):
//   - 0: no issues
//   - 1: at least one issue (any severity)
//   - 2: internal error (config load failure, validate.Run error, etc.)
//
// JSON mode short-circuits text rendering and emits a single Report
// object. The text mode groups issues by agent and renders a small
// CatalogTable per agent + a one-line footer.
func runValidate(cmd *cobra.Command, args []string) error {
	cwd := mustCwd()
	cfg, err := requireConfig(filepath.Join(cwd, configFile))
	if err != nil {
		// requireConfig already calls FatalPanel on missing config and
		// won't return; this branch only fires on YAML parse / Validate
		// errors. Surface them and exit 2.
		fmt.Fprintf(os.Stderr, "validate: %s\n", err.Error())
		os.Exit(2)
	}
	cat := loadCatalog()
	lock, _ := config.LoadLockFile(cwd)
	if lock == nil {
		// LoadLockFile returns an empty file when missing, so this is
		// belt-and-braces against future signature drift.
		lock = config.NewLockFile()
	}

	agentFilter, _ := cmd.Flags().GetString("agent")
	jsonOut, _ := cmd.Flags().GetBool("json")

	report, err := validate.Run(cwd, cfg, cat, lock, agentFilter)
	if err != nil {
		if jsonOut {
			fmt.Fprintf(os.Stderr, "validate error: %s\n", err.Error())
		} else {
			tui.Error("validate: " + err.Error())
		}
		os.Exit(2)
	}

	if jsonOut {
		if err := renderValidateJSON(report); err != nil {
			fmt.Fprintf(os.Stderr, "validate: %s\n", err.Error())
			os.Exit(2)
		}
	} else {
		renderValidateText(report)
	}

	if report.HasIssues() {
		os.Exit(1)
	}
	return nil
}

// renderValidateJSON marshals the report to indent-2 JSON and prints it
// to stdout. Returns the error rather than os.Exit-ing so the caller can
// consolidate exit-code handling.
func renderValidateJSON(report *validate.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// renderValidateText prints the human-readable audit summary. When
// report has no issues, prints a Success line + the count of agents
// scanned. Otherwise groups issues by agent and renders a small table
// per agent, then a footer counting errors + warnings.
//
// Per-agent grouping keeps output readable in multi-agent projects;
// rendering one giant cross-agent table would force the user to mentally
// re-group rows by the AgentName column. The small header per agent
// gives immediate orientation.
func renderValidateText(report *validate.Report) {
	if !report.HasIssues() {
		tui.Success(fmt.Sprintf("No issues found. (%d agent(s) scanned)", len(report.AgentsScanned)))
		return
	}

	// Group issues by agent name. Preserve a sorted agent list for
	// deterministic output.
	byAgent := make(map[string][]validate.Issue)
	for _, iss := range report.Issues {
		byAgent[iss.AgentName] = append(byAgent[iss.AgentName], iss)
	}
	agents := make([]string, 0, len(byAgent))
	for n := range byAgent {
		agents = append(agents, n)
	}
	sort.Strings(agents)

	for _, agent := range agents {
		label := agent
		if label == "" {
			label = "(unscoped)"
		}
		tui.SectionHeader("agent: " + label)
		var rows [][]string
		for _, iss := range byAgent[agent] {
			rows = append(rows, []string{
				string(iss.Severity),
				string(iss.Category),
				iss.AbilityType,
				iss.Name,
				iss.Path,
				iss.Detail,
			})
		}
		tui.CatalogTable([]string{"Severity", "Category", "Type", "Name", "Path", "Detail"}, rows)
	}

	// Footer: total count split by severity.
	var errs, warns int
	for _, iss := range report.Issues {
		switch iss.Severity {
		case validate.SeverityError:
			errs++
		case validate.SeverityWarning:
			warns++
		}
	}
	tui.Blank()
	tui.Warning(fmt.Sprintf("%d issue(s): %d error(s), %d warning(s) — run `bonsai update` to fix orphans/untracked",
		len(report.Issues), errs, warns))
}
