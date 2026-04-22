package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
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
