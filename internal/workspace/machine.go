package workspace

// This machine's part of a workspace (contract §3, §5.3; spec §10): the workspace's machine folder in the Bonsai
// home, <home>/workspaces/r-<16 hex>/ (MachineKey), shared by the main checkout and its worktrees. It holds what
// belongs to this machine and this checkout and must not follow the id into a copy:
//
//	settings.json             this machine's settings for the workspace: status_writes, status_command
//	labels/<namespace>.yaml   label definitions attached on this machine (bonsai.labels/1)
//	workspace.json            the main checkout's path and the ids it has held (contract §3; record.go, step 5.1.6)
//
// Read here: the settings (LoadMachineSettings) and the attached labels (LabelsInForce, with the packs' from the
// lock's declares). Both are written only by Bonsai's commands run outside an agent session (contract §3, §10.6):
// bonsai settings set and the attach command come with step 5.6. With no settings file, agents move statuses
// themselves (status_writes agents), as in any workspace no studio manages.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

// The machine folder's files and folders.
const (
	MachineSettingsFile = "settings.json"
	MachineLabelsDir    = "labels"
)

// The values of status_writes (status --json's closed list, bonsai.status/1).
const (
	StatusWritesAgents  = "agents"
	StatusWritesCommand = "command"
)

// MachineDir is the workspace's machine folder: <home>/workspaces/<MachineKey of the main checkout>.
func MachineDir(home, main string) (string, error) {
	key, err := MachineKey(main)
	if err != nil {
		return "", &Error{File: filepath.ToSlash(main), Msg: "cannot be resolved: " + oneLine(err), Err: err,
			Next: "check that the main checkout exists and can be read"}
	}
	return filepath.Join(home, "workspaces", key), nil
}

// MachineSettings is this machine's settings for a workspace.
type MachineSettings struct {
	StatusWrites  string // agents (agents move statuses themselves) or command (moves go through StatusCommand)
	StatusCommand string // the command a managed workspace's guard names for a status move; "" with agents
}

// LoadMachineSettings reads <machine folder>/settings.json: a JSON object whose status_writes is agents or command
// and whose status_command, with command, is the command's text. A missing file (or one with neither key) is
// status_writes agents; status_command is "" unless status_writes is command. A key this Bonsai does not know is
// left as it is. Its error is an *Error naming the file and the next step.
func LoadMachineSettings(home, main string) (MachineSettings, error) {
	out := MachineSettings{StatusWrites: StatusWritesAgents}
	dir, err := MachineDir(home, main)
	if err != nil {
		return out, err
	}
	path := filepath.Join(dir, MachineSettingsFile)
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return out, nil
	}
	show := filepath.ToSlash(path)
	next := "fix " + show + " (this machine's settings for the workspace), or delete it to go back to status_writes agents"
	if err != nil {
		return out, &Error{File: show, Msg: "cannot be read: " + oneLine(err), Err: err, Next: "check the file's permissions"}
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	o, isObject := v.(schema.Object)
	if err != nil || !isObject {
		return out, &Error{File: show, Msg: "is not a JSON object Bonsai reads", Next: next}
	}
	if w, ok := o.Get("status_writes"); ok && w != nil {
		s, _ := w.(string)
		if s != StatusWritesAgents && s != StatusWritesCommand {
			return out, &Error{File: show, Msg: "status_writes is " + schema.Show(w) + ", not agents or command", Next: next}
		}
		out.StatusWrites = s
	}
	if out.StatusWrites == StatusWritesCommand {
		c, _ := o.Get("status_command")
		s, _ := c.(string)
		if strings.TrimSpace(s) == "" {
			return out, &Error{File: show, Msg: "status_writes is command, but status_command names no command", Next: next}
		}
		out.StatusCommand = s
	}
	return out, nil
}

// LabelSet is one namespace of label definitions in force in a workspace (contract §5): a pack's, from the lock's
// declares, or one attached on this machine.
type LabelSet struct {
	Namespace string
	From      string // the pack's id, or machine (attached on this machine, contract §5.3)
	Version   int64  // the definitions' own version
	Labels    []format.LabelDef
	File      string // machine: the file it was read from, with forward slashes; "" for a pack's
}

// FromMachine is a LabelSet's From for definitions attached on this machine (status --json's labels[].from).
const FromMachine = "machine"

// LabelsInForce lists the label definitions in force: each locked pack's (its declares' labels, in the lock's
// order), then those attached on this machine (<machine folder>/labels/<namespace>.yaml, by file name). An attached
// file is left out, with a problem, when it does not read as bonsai.labels/1, when its name is not its namespace,
// or when its namespace is a pack's in the workspace or bonsai (contract §5.1: Bonsai refuses such an attach, so
// such a file was not put there by Bonsai). lock may be nil. Two sources defining one label name stay in the list:
// that is a bonsai check finding (contract §5.1), not a reader's choice.
func LabelsInForce(lock *Lock, home, main string) ([]LabelSet, []error) {
	var sets []LabelSet
	var problems []error
	taken := map[string]bool{"bonsai": true}
	if lock != nil {
		for _, lp := range lock.Packs {
			d, err := lp.Declared()
			if err != nil {
				problems = append(problems, lockError("the pack %s's %v", showValue(lp.ID), err))
				continue
			}
			if d.Labels == nil {
				continue
			}
			taken[d.Labels.Namespace] = true
			sets = append(sets, LabelSet{Namespace: d.Labels.Namespace, From: lp.ID, Version: d.Labels.Version, Labels: d.Labels.Labels})
		}
	}
	if home == "" || main == "" {
		return sets, problems
	}
	machine, err := MachineDir(home, main)
	if err != nil {
		return sets, append(problems, err)
	}
	dir := filepath.Join(machine, MachineLabelsDir)
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return sets, append(problems, &Error{File: filepath.ToSlash(dir), Msg: "cannot be read: " + oneLine(err), Err: err,
			Next: "check the folder's permissions"})
	}
	var names []string
	for _, e := range entries {
		if e.Type().IsRegular() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.ToSlash(filepath.Join(dir, name))
		next := "attach the definitions again with Bonsai's attach command (step 5.6), or delete " + path
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			problems = append(problems, &Error{File: path, Msg: "cannot be read: " + oneLine(err), Err: err, Next: "check the file's permissions"})
			continue
		}
		l, err := format.ReadLabels(raw)
		switch {
		case err != nil:
			problems = append(problems, &Error{File: path, Msg: "is not label definitions Bonsai reads: " + oneLine(err), Next: next})
		case l.Namespace+".yaml" != name:
			problems = append(problems, &Error{File: path, Msg: "holds the namespace " + showValue(l.Namespace) +
				", not its file's name", Next: next})
		case taken[l.Namespace]:
			problems = append(problems, &Error{File: path, Msg: "attaches the namespace " + showValue(l.Namespace) +
				", which a pack in the workspace (or Bonsai) already has", Next: "delete " + path + ": a namespace attached on the machine never equals a pack's (contract section 5.1)"})
		default:
			sets = append(sets, LabelSet{Namespace: l.Namespace, From: FromMachine, Version: l.Version, Labels: l.Labels, File: path})
		}
	}
	return sets, problems
}
