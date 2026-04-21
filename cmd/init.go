package cmd

import (
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Bonsai in the current project.",
	RunE:  runInit,
}

// userSensorOptions filters out the auto-managed routine-check sensor so the
// user only picks from sensors they actually choose. Shared with `bonsai add`.
func userSensorOptions(cat *catalog.Catalog, agentType string) []tui.ItemOption {
	available := cat.SensorsFor(agentType)
	filtered := make([]catalog.SensorItem, 0, len(available))
	for _, s := range available {
		if s.Name != "routine-check" {
			filtered = append(filtered, s)
		}
	}
	return toSensorOptions(filtered, agentType)
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
