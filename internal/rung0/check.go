package rung0

// bonsai check's findings about the commit (the package comment, note 4).

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// checkFindings runs bonsai check in the checkout: the findings that count, each "<word> (<file>)" in check's order;
// the machine words it found, each once, which do not; and why check could not run, "" when it ran.
func checkFindings(in Input) (found, machine []string, cannot string) {
	res, err := engine.CheckWith(in.Root, in.Home, engine.CheckOptions{History: true})
	if err != nil {
		var ee *engine.Error
		if errors.As(err, &ee) {
			return nil, nil, oneLine(ee.What)
		}
		return nil, nil, oneLine(err.Error())
	}
	seen := map[string]bool{}
	for _, f := range res.Findings {
		w, ok := format.CheckWord(f.Code)
		if ok && w.Machine {
			if !seen[f.Code] {
				seen[f.Code] = true
				machine = append(machine, f.Code)
			}
			continue
		}
		s := f.Code
		if f.File != "" {
			s += " (" + f.File + ")"
		}
		found = append(found, s)
	}
	// hooks-off from the verified main checkout's local settings, when check read another main (or none).
	if !samePlace(in.Main, res.Main) && !samePlace(in.Main, res.Root) && hooksOff(in.Main) {
		found = append(found, "hooks-off (the main checkout's "+engine.LocalSettingsFile+")")
	}
	return found, machine, ""
}

// hooksOff reports whether a checkout's .claude/settings.local.json sets disableAllHooks: true, as check's hooks-off
// reads it (a file that does not read sets nothing).
func hooksOff(root string) bool {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(engine.LocalSettingsFile)))
	if err != nil {
		return false
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	if err != nil {
		return false
	}
	o, ok := v.(schema.Object)
	if !ok {
		return false
	}
	off, _ := o.Get("disableAllHooks")
	return off == true
}
