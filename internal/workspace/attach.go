package workspace

// Attaching label definitions on this machine (contract §5.1-§5.3; plan-5 5.6.2): a bonsai.labels/1 file copied
// into the workspace's machine folder as labels/<namespace>.yaml, or taken out again. The rules, in one place:
//
//   - the namespace is never bonsai or a locked pack's (only that pack defines its own): AttachTaken;
//   - attached again, the same content is Unchanged; a newer version that keeps every attached definition exactly
//     as it is and only adds is Updated; anything else (a definition changed or removed, other content at the same or
//     an older version, a file already there that does not read) is refused, AttachRedefined: "a definition may be
//     added, never redefined" (§5.1), and a person detaches the namespace first;
//   - the copy has LF line endings, and is staged then renamed, so a reader sees the old file or the new one.
//
// The commands (cmd/bonsai/labels.go) own the person-only refusal and the output; this file holds what is checked
// and written. LabelsInForce (machine.go) reads the folder back.

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/LastStep/Bonsai/internal/format"
)

// The results of an attach, as bonsai.attach/1's result names them, and a detach's two.
const (
	AttachAttached  = "attached"
	AttachUpdated   = "updated"
	AttachUnchanged = "unchanged"
	AttachDetached  = "detached"
	AttachNothing   = "nothing"
)

// The codes of an AttachError, each an error word of bonsai's closed list.
const (
	AttachBadValue   = "bad-value"       // the file does not read as bonsai.labels/1, or a namespace is not in its form
	AttachTaken      = "namespace-taken" // the namespace is bonsai or a locked pack's
	AttachRedefined  = "label-redefined" // the attach would change a definition already attached
	AttachWriteFail  = "write-failed"    // the machine folder's file could not be written or removed
	AttachReadFailed = "read-failed"     // the machine folder's file could not be read
)

// AttachError is a refusal of an attach or a detach: the error word, what is wrong and the step. Namespace and
// Version are what the command knew when it refused (Version 0: none).
type AttachError struct {
	Code      string
	Msg       string
	Next      string
	Namespace string
	Version   int64
	Err       error
}

func (e *AttachError) Error() string { return e.Msg }
func (e *AttachError) Unwrap() error { return e.Err }

// AttachOutcome is what an attach or a detach did.
type AttachOutcome struct {
	Result    string   // the Attach* results above
	Namespace string   // the namespace attached or detached
	Version   int64    // the definitions' own version; 0 for a detach that found no file
	File      string   // the machine folder's labels/<namespace>.yaml, absolute, forward slashes
	Labels    []string // the label names attached afterwards, in the file's order; [] after a detach
}

// namespaceForm is a namespace's form (bonsai.labels/1's namespace pattern).
var namespaceForm = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ValidNamespace reports whether s is in a namespace's form.
func ValidNamespace(s string) bool { return namespaceForm.MatchString(s) }

// LabelsFile is the machine folder's file for a namespace: <home>/workspaces/<key>/labels/<namespace>.yaml.
func LabelsFile(home, main, namespace string) (string, error) {
	machine, err := MachineDir(home, main)
	if err != nil {
		return "", err
	}
	return filepath.Join(machine, MachineLabelsDir, namespace+".yaml"), nil
}

// AttachLabels attaches the definitions in raw (a bonsai.labels/1 file's bytes) for the workspace whose main checkout
// is main. lock (may be nil) names the locked packs, whose namespaces are taken. It writes nothing unless it returns
// attached or updated. A refusal is an *AttachError.
func AttachLabels(lock *Lock, home, main string, raw []byte) (*AttachOutcome, error) {
	raw = bytes.ReplaceAll(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")), []byte("\r\n"), []byte("\n"))
	l, err := format.ReadLabels(raw)
	if err != nil {
		return nil, &AttachError{Code: AttachBadValue, Msg: "the file does not read as bonsai.labels/1: " + oneLine(err),
			Next: "fix the file (bonsai check --schema bonsai.labels shows the format), then attach it again"}
	}
	if l.Namespace == "bonsai" {
		return nil, &AttachError{Code: AttachTaken, Namespace: l.Namespace, Version: l.Version,
			Msg:  "the namespace bonsai is Bonsai's own: only Bonsai and the packs it publishes define labels in it",
			Next: "give the definitions a namespace of their own (the studio's name, say), then attach them again"}
	}
	if lock != nil {
		for _, lp := range lock.Packs {
			d, derr := lp.Declared()
			if derr != nil || d.Labels == nil || d.Labels.Namespace != l.Namespace {
				continue
			}
			return nil, &AttachError{Code: AttachTaken, Namespace: l.Namespace, Version: l.Version,
				Msg:  "the namespace " + showValue(l.Namespace) + " is the locked pack " + showValue(lp.ID) + "'s: only that pack defines its labels",
				Next: "give the definitions a namespace of their own, then attach them again"}
		}
	}
	path, err := LabelsFile(home, main, l.Namespace)
	if err != nil {
		return nil, err
	}
	out := &AttachOutcome{Result: AttachAttached, Namespace: l.Namespace, Version: l.Version, File: filepath.ToSlash(path)}
	for _, d := range l.Labels {
		out.Labels = append(out.Labels, d.Name)
	}
	if out.Labels == nil {
		out.Labels = []string{}
	}
	redefined := func(why string) error {
		return &AttachError{Code: AttachRedefined, Namespace: l.Namespace, Version: l.Version,
			Msg:  "the namespace " + showValue(l.Namespace) + " is already attached, and this file " + why + ": a definition may be added, never redefined",
			Next: "if you mean to replace it, run `bonsai labels detach " + l.Namespace + "` and then attach the file again"}
	}
	old, err := os.ReadFile(path)
	switch {
	case err == nil:
		old = bytes.ReplaceAll(old, []byte("\r\n"), []byte("\n"))
		if bytes.Equal(old, raw) {
			out.Result = AttachUnchanged
			return out, nil
		}
		o, rerr := format.ReadLabels(old)
		switch {
		case rerr != nil:
			return nil, redefined("replaces a file there that does not read as bonsai.labels/1")
		case l.Version <= o.Version:
			return nil, redefined("holds other content at version " + itoa(int(l.Version)) + ", not newer than the attached " + itoa(int(o.Version)))
		}
		for _, od := range o.Labels {
			kept := false
			for _, nd := range l.Labels {
				if nd.Name == od.Name {
					kept = reflect.DeepEqual(nd, od)
					break
				}
			}
			if !kept {
				return nil, redefined("changes or drops the definition " + showValue(od.Name))
			}
		}
		out.Result = AttachUpdated
	case errors.Is(err, fs.ErrNotExist):
	default:
		return nil, &AttachError{Code: AttachReadFailed, Namespace: l.Namespace, Version: l.Version, Err: err,
			Msg: filepath.ToSlash(path) + " cannot be read: " + oneLine(err), Next: "check the file's permissions"}
	}
	if err := WriteFileAtomic(path, raw); err != nil {
		return nil, &AttachError{Code: AttachWriteFail, Namespace: l.Namespace, Version: l.Version, Err: err,
			Msg: filepath.ToSlash(path) + " cannot be written: " + oneLine(err), Next: "check the Bonsai home's permissions, then attach the file again"}
	}
	return out, nil
}

// DetachLabels takes a namespace's attached definitions out of the workspace's machine folder: detached, or nothing
// when no file is there. It needs only the main checkout's place, so a project whose bonsai.yaml is gone can be
// unregistered.
func DetachLabels(home, main, namespace string) (*AttachOutcome, error) {
	if !ValidNamespace(namespace) {
		return nil, &AttachError{Code: AttachBadValue, Msg: "the namespace " + showValue(namespace) + " is not in a namespace's form (lower-case letters, digits and hyphens, starting with a letter)",
			Next: "give the namespace as bonsai status --json lists it"}
	}
	path, err := LabelsFile(home, main, namespace)
	if err != nil {
		return nil, err
	}
	out := &AttachOutcome{Result: AttachNothing, Namespace: namespace, File: filepath.ToSlash(path), Labels: []string{}}
	raw, err := os.ReadFile(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return out, nil
	case err != nil:
		return nil, &AttachError{Code: AttachReadFailed, Namespace: namespace, Err: err,
			Msg: filepath.ToSlash(path) + " cannot be read: " + oneLine(err), Next: "check the file's permissions"}
	}
	if l, rerr := format.ReadLabels(raw); rerr == nil {
		out.Version = l.Version
	}
	if err := RemoveFile(path); err != nil {
		return nil, &AttachError{Code: AttachWriteFail, Namespace: namespace, Version: out.Version, Err: err,
			Msg: filepath.ToSlash(path) + " cannot be removed: " + oneLine(err), Next: "check the Bonsai home's permissions, then detach again"}
	}
	out.Result = AttachDetached
	return out, nil
}
