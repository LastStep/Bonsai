package engine

// What a pack declares (spec §5, §6; contract §5.2, §6, §7.3), read from the pack at a commit and copied into the
// lock's declares (format.Declares gives its layout), so bonsai check, status and the guard read them with no pack
// fetched: its lanes (bonsai/lanes.yaml), document kinds (pack.yaml's documents), label definitions
// (bonsai/labels.yaml), protected paths (pack.yaml's protected), the hook lines and deny rules check needs to tell
// Bonsai's lines in .claude/settings.json from the project's own, and its needs (the Claude Code floor, step 5.1.6).
//
// Read here with the rules a schema cannot say, each refused as bad-pack with the pack maker's step:
//   - labels: the namespace is the pack's own id, or bonsai (contract §5.1: bonsai.* is Bonsai's and its own packs';
//     each other pack uses its own id), and every label's name starts with the namespace and a dot, once;
//   - lanes: each name once;
//   - document kinds: each name once, none of Bonsai's own kinds (DocKinds' bonsai rows) nor protocols (bonsai.yaml's
//     documents key for the protocols folder), exactly one of path and file, and an id pattern Go's regexp reads.
// Two linked packs declaring one document kind are refused in Build (packs-overlap); two defining one label name are
// a bonsai check finding (contract §5.1), not a refusal.

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The pack files that hold a pack's lanes and labels, beside bonsai/pack.yaml (spec §5's folder layout).
const (
	LabelsFile = "bonsai/labels.yaml"
	LanesFile  = "bonsai/lanes.yaml"
)

// readDeclares reads what a pack declares, read giving a file of the pack's folder (os.ErrNotExist when it has none).
// It refuses the first problem the rules below find (labelsProblems, lanesProblems, documentsProblems,
// protectedProblems), which bonsai check --pack lists all of (checkpack.go).
func readDeclares(manifest *workspace.Pack, read func(string) ([]byte, error)) (*format.Declares, error) {
	d := &format.Declares{Documents: manifest.Full.Documents, Protected: manifest.Full.Protected,
		Hooks: manifest.Full.Hooks, Deny: manifest.Full.Deny}
	if format.NeedsAny(&manifest.Full.Needs) {
		needs := manifest.Full.Needs
		d.Needs = &needs
	}
	if raw, err := read(LanesFile); err == nil {
		lanes, err := format.ReadLanes(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", LanesFile, err)
		}
		if ps := lanesProblems(lanes); len(ps) > 0 {
			return nil, errors.New(ps[0])
		}
		d.Lanes = lanes.Lanes
	} else if !missing(err) {
		return nil, fmt.Errorf("%s is %v", LanesFile, err)
	}
	if raw, err := read(LabelsFile); err == nil {
		labels, err := format.ReadLabels(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", LabelsFile, err)
		}
		if ps := labelsProblems(labels, manifest.ID); len(ps) > 0 {
			return nil, errors.New(ps[0])
		}
		d.Labels = labels
	} else if !missing(err) {
		return nil, fmt.Errorf("%s is %v", LabelsFile, err)
	}
	if ps := documentsProblems(d.Documents); len(ps) > 0 {
		return nil, errors.New(ps[0].text)
	}
	if ps := protectedProblems(d.Protected); len(ps) > 0 {
		return nil, errors.New(ps[0].text)
	}
	return d, nil
}

// lanesProblems are a lanes.yaml's problems its schema cannot say: each lane's name once.
func lanesProblems(lanes *format.Lanes) []string {
	var out []string
	seen := map[string]bool{}
	for _, l := range lanes.Lanes {
		if seen[l.Name] {
			out = append(out, fmt.Sprintf("%s defines the lane %s twice", LanesFile, quote(l.Name)))
		}
		seen[l.Name] = true
	}
	return out
}

// labelsProblems are a labels.yaml's problems its schema cannot say (contract §5.1): its namespace is the pack's own
// id or bonsai, and each label's name starts with the namespace and a dot, once.
func labelsProblems(labels *format.Labels, id string) []string {
	var out []string
	if labels.Namespace != id && labels.Namespace != ReservedNamespace {
		out = append(out, fmt.Sprintf("%s's namespace is %s: a pack's labels use its own id (%s), or bonsai for Bonsai's own packs",
			LabelsFile, quote(labels.Namespace), id))
	}
	seen := map[string]bool{}
	for _, l := range labels.Labels {
		switch {
		case !strings.HasPrefix(l.Name, labels.Namespace+"."):
			out = append(out, fmt.Sprintf("%s's label %s is not in its namespace %s", LabelsFile, quote(l.Name), labels.Namespace))
		case seen[l.Name]:
			out = append(out, fmt.Sprintf("%s defines the label %s twice", LabelsFile, quote(l.Name)))
		}
		seen[l.Name] = true
	}
	return out
}

// itemProblem is a problem of one item of a pack.yaml list (documents, protected): the item's index and the sentence.
type itemProblem struct {
	item int
	text string
}

// documentsProblems are the problems of pack.yaml's document kinds that their schema cannot say: each kind's name
// once and none of Bonsai's own kinds (DocKinds' bonsai rows) nor protocols (bonsai.yaml's documents key for the
// protocols folder), exactly one of path and file, each a project path, and an id pattern Go's regexp reads.
func documentsProblems(docs []format.PackDocument) []itemProblem {
	var out []itemProblem
	seen := map[string]bool{}
	for i, k := range docs {
		where := fmt.Sprintf("pack.yaml's documents item %d (%s)", i+1, k.Kind)
		add := func(f string, args ...any) { out = append(out, itemProblem{i, where + fmt.Sprintf(f, args...)}) }
		switch {
		case workspace.BonsaiKind(k.Kind) || k.Kind == "protocols":
			add(" takes the name of Bonsai's own %s; a pack's kind has a name of its own", k.Kind)
		case seen[k.Kind]:
			add(": the kind is declared twice")
		case (k.Path == nil) == (k.File == nil):
			add(" has %s: a kind is a folder (path) or one file (file)",
				map[bool]string{true: "neither path nor file", false: "both path and file"}[k.Path == nil])
		}
		seen[k.Kind] = true
		for _, p := range []*string{k.Path, k.File} {
			if p != nil {
				if err := workspace.CheckRelPath(*p); err != nil {
					add(": %v", err)
				}
			}
		}
		if k.ID != nil {
			if _, err := regexp.Compile(*k.ID); err != nil {
				add("'s id pattern %s is not a regular expression Bonsai reads: %v", quote(*k.ID), err)
			}
		}
	}
	return out
}

// protectedProblems are the problems of pack.yaml's protected globs: each project-relative with forward slashes.
func protectedProblems(globs []string) []itemProblem {
	var out []itemProblem
	for i, g := range globs {
		if strings.HasPrefix(g, "/") || strings.ContainsAny(g, `\:`) {
			out = append(out, itemProblem{i, fmt.Sprintf("pack.yaml's protected glob %s is not project-relative with forward slashes", quote(g))})
		}
	}
	return out
}

// ReservedNamespace is the label namespace Bonsai keeps for itself and its own packs (contract §5.1, §5.6).
const ReservedNamespace = "bonsai"

// missing reports whether a pack file read failed because the pack has no such file.
func missing(err error) bool { return errors.Is(err, os.ErrNotExist) }

func quote(s string) string { return strconv.QuoteToASCII(s) }

// declaredHooks gives a pack's declared hook lines as the engine's settings lines read them.
func declaredHooks(d *format.Declares) []workspace.HookEntry {
	out := []workspace.HookEntry{}
	for _, h := range d.Hooks {
		e := workspace.HookEntry{Event: h.Event, Command: h.Command, Runs: h.Runs, Why: h.Why}
		if h.Matcher != nil {
			e.Matcher = *h.Matcher
		}
		out = append(out, e)
	}
	return out
}

// declaredDeny gives a pack's declared deny rules as the engine's settings lines read them.
func declaredDeny(d *format.Declares) []workspace.DenyEntry {
	out := []workspace.DenyEntry{}
	for _, x := range d.Deny {
		out = append(out, workspace.DenyEntry{Rule: x.Rule, Why: x.Why})
	}
	return out
}

// lockNext is the next step for a lock Bonsai does not read: restore it from git (the lock is Bonsai's to write).
const lockNext = "run: git checkout -- " + workspace.LockFile
