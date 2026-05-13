package cmd

import (
	"github.com/spf13/cobra"
)

// Non-interactive flags for `bonsai init`. Both must be set together; the
// runtime guard lives in runInit so the cobra layer doesn't have to know
// about the cross-flag dependency. See Plan 39 §B.
var (
	initNonInteractive bool
	initFromConfig     string
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false,
		"Skip TUI prompts; read all answers from --from-config (must be paired with --from-config)")
	initCmd.Flags().StringVar(&initFromConfig, "from-config", "",
		"Path to a YAML config file (.bonsai.yaml shape) used as input under --non-interactive")
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Bonsai in the current project.",
	RunE:  runInit,
}

// asString safely extracts a string result from a harness step. Returns ""
// for nil to keep call-site logic short — empty input is already meaningful
// (e.g., optional Description field).
func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// asStringSlice safely extracts a []string result. Returns nil for absent
// results so downstream nil checks behave as before the harness migration.
func asStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}

// asBool safely extracts a bool result.
func asBool(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
