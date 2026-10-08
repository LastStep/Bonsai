package workspace

// The lock, .bonsai/lock.json (bonsai.lock/1, contract §14, spec §6): exactly which pack versions and files a
// workspace holds. Bonsai's engine writes it (parts 3 and 4); bonsai check and status --json read it.
//
// Read and written here in full: format, written_by, packs (id, source, version, commit, sha256, declares), files
// (kind, pack, sha256 per path) and format0 (sha256 per path). The schema in formats/schemas/lock.schema.json holds
// every field's form and the closed list of file kinds; this file adds what a schema cannot say: no path leaves the
// project or names a file Windows cannot hold (CheckRelPath), and no pack is locked twice. declares stays an open
// object: its inner layout is step 5.1's (formats/README.md).
//
// Written byte-stable: fields in the schema's order, files and format0 sorted by path, two-space indent, LF, ASCII
// (schema.Encode). An unknown field (a newer Bonsai of the same major wrote it) is kept, value for value, after the
// fields this Bonsai knows, at the top level, in a pack and in a file entry (contract §2.2).

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/schema"
)

// LockFile is the lock's path in a project.
const LockFile = ".bonsai/lock.json"

// LockFormat is the lock's format.
const LockFormat = "bonsai.lock/1"

// Lock is the lock's content.
type Lock struct {
	WrittenBy string                // the version of the Bonsai that wrote it
	Packs     []LockedPack          // in bonsai.yaml's order
	Files     map[string]LockedFile // by project-relative path
	Format0   map[string]string     // project-relative path to its SHA-256 with line endings made LF
	Extra     schema.Object         // unknown top-level fields, kept in order
}

// LockedPack is one pack in the lock.
type LockedPack struct {
	ID       string
	Source   string
	Version  string
	Commit   string        // the resolved 40-character commit
	SHA256   string        // the pack's content hash
	Declares schema.Object // what the pack declared at that commit; {} for nothing
	Extra    schema.Object
}

// LockedFile is one file the engine wrote.
type LockedFile struct {
	Kind   string // pack, once, block, keys or kept: the lock schema's list
	Pack   string
	SHA256 string // what Bonsai last wrote, line endings made LF
	Extra  schema.Object
}

var (
	lockSchemaOnce             sync.Once
	lockSchema, lockReadSchema schema.Object
)

// LockSchema is bonsai.lock/1's schema, from formats/schemas (embedded): what a writer writes.
func LockSchema() schema.Object {
	lockSchemaOnce.Do(func() {
		raw, err := formats.Schema("lock")
		if err != nil {
			panic(err)
		}
		if lockSchema, err = schema.Parse(raw); err != nil {
			panic(err)
		}
		lockReadSchema = schema.Lenient(lockSchema)
	})
	return lockSchema
}

// knownLockFields are the fields this Bonsai reads, per level, in the schema's order: a null one reads as missing.
var knownLockFields = map[string][]string{
	"top":  {"written_by", "packs", "files", "format0"},
	"pack": {"id", "source", "version", "commit", "sha256", "declares"},
	"file": {"kind", "pack", "sha256"},
}

// dropNull removes the known fields that are null: contract §2.2's reader treats a missing field as null, so a null
// one reads as a missing one.
func dropNull(o schema.Object, level string) schema.Object {
	out := schema.Object{}
	for _, m := range o {
		if m.Value == nil && contains(knownLockFields[level], m.Key) {
			continue
		}
		out = append(out, m)
	}
	return out
}

const lockNext = "restore .bonsai/lock.json from git (git checkout -- .bonsai/lock.json); the lock is Bonsai's to write"

func lockError(format string, args ...any) error {
	return &Error{File: LockFile, Msg: fmt.Sprintf(format, args...), Next: lockNext}
}

// ReadLock reads the lock's bytes: UTF-8 JSON with no duplicate key (contract §2.5), its format first checked, then
// the whole document held to the lock schema as a reader reads it (contract §2.2: a missing field, or a null one,
// reads as null, so as an empty value: no packs, no files, an empty commit), then its paths. An unknown field is kept.
func ReadLock(raw []byte) (*Lock, error) {
	v, err := schema.Decode(raw)
	if err != nil {
		return nil, lockError("is not a JSON document Bonsai reads: %s", oneLine(err))
	}
	doc, ok := v.(schema.Object)
	if !ok {
		return nil, lockError("is not a JSON object")
	}
	if f, _ := doc.Get("format"); f != LockFormat {
		if s, isText := f.(string); isText && len(s) > len("bonsai.lock/") && s[:len("bonsai.lock/")] == "bonsai.lock/" {
			return nil, &Error{File: LockFile, Msg: "format too new: " + showValue(s) + " (this Bonsai reads bonsai.lock/1)",
				Next: "use a Bonsai that reads it"}
		}
		return nil, lockError("its format is %s, not %s", schema.Show(f), LockFormat)
	}
	doc = dropNull(doc, "top")
	if packs, ok := doc.Get("packs"); ok {
		if list, ok := packs.([]any); ok {
			for j, pv := range list {
				if po, ok := pv.(schema.Object); ok {
					list[j] = dropNull(po, "pack")
				}
			}
		}
	}
	if files, ok := doc.Get("files"); ok {
		if fo, ok := files.(schema.Object); ok {
			for j, fm := range fo {
				if entry, ok := fm.Value.(schema.Object); ok {
					fo[j].Value = dropNull(entry, "file")
				}
			}
		}
	}
	LockSchema() // loads lockReadSchema too
	if msgs := schema.Validate(lockReadSchema, doc); len(msgs) > 0 {
		return nil, lockError("does not fit bonsai.lock/1: %s", asciiOnly(msgs[0]))
	}
	l := &Lock{Files: map[string]LockedFile{}, Format0: map[string]string{}}
	for _, m := range doc {
		switch m.Key {
		case "format":
		case "written_by":
			l.WrittenBy = m.Value.(string)
		case "packs":
			for _, pv := range m.Value.([]any) {
				po := pv.(schema.Object)
				p := LockedPack{Declares: schema.Object{}}
				for _, pm := range po {
					switch pm.Key {
					case "id":
						p.ID = pm.Value.(string)
					case "source":
						p.Source = pm.Value.(string)
					case "version":
						p.Version = pm.Value.(string)
					case "commit":
						p.Commit = pm.Value.(string)
					case "sha256":
						p.SHA256 = pm.Value.(string)
					case "declares":
						p.Declares = pm.Value.(schema.Object)
					default:
						p.Extra = append(p.Extra, pm)
					}
				}
				l.Packs = append(l.Packs, p)
			}
		case "files":
			for _, fm := range m.Value.(schema.Object) {
				fo := fm.Value.(schema.Object)
				f := LockedFile{}
				for _, x := range fo {
					switch x.Key {
					case "kind":
						f.Kind = x.Value.(string)
					case "pack":
						f.Pack = x.Value.(string)
					case "sha256":
						f.SHA256 = x.Value.(string)
					default:
						f.Extra = append(f.Extra, x)
					}
				}
				l.Files[fm.Key] = f
			}
		case "format0":
			for _, x := range m.Value.(schema.Object) {
				l.Format0[x.Key] = x.Value.(string)
			}
		default:
			l.Extra = append(l.Extra, m)
		}
	}
	if err := l.check(); err != nil {
		return nil, err
	}
	return l, nil
}

// check holds the lock to what its schema cannot say: every path a project path Bonsai may write or record, and
// each pack once.
func (l *Lock) check() error {
	seen := map[string]bool{}
	for _, p := range l.Packs {
		if seen[p.ID] {
			return lockError("the pack %s is locked twice", showValue(p.ID))
		}
		seen[p.ID] = true
	}
	for _, list := range []struct {
		name  string
		paths []string
	}{{"files", sortedKeys(l.Files)}, {"format0", sortedKeys(l.Format0)}} {
		folded := map[string]string{}
		for _, path := range list.paths {
			if err := CheckRelPath(path); err != nil {
				return lockError("%s: %v", list.name, err)
			}
			// Windows and macOS hold one file for two paths that differ only in letter case.
			key := strings.ToLower(path)
			if other, ok := folded[key]; ok {
				return lockError("%s: the paths %s and %s differ only in letter case, one file on Windows", list.name,
					showValue(other), showValue(path))
			}
			folded[key] = path
		}
	}
	return nil
}

// JSON gives the lock as the document bonsai.lock/1 writes: fields in the schema's order, files and format0 sorted
// by path, unknown fields after the known ones.
func (l *Lock) JSON() schema.Object {
	packs := []any{}
	for _, p := range l.Packs {
		declares := p.Declares
		if declares == nil {
			declares = schema.Object{}
		}
		o := schema.Object{{Key: "id", Value: p.ID}, {Key: "source", Value: p.Source}, {Key: "version", Value: p.Version},
			{Key: "commit", Value: p.Commit}, {Key: "sha256", Value: p.SHA256}, {Key: "declares", Value: declares}}
		packs = append(packs, append(o, p.Extra...))
	}
	files := schema.Object{}
	for _, path := range sortedKeys(l.Files) {
		f := l.Files[path]
		o := schema.Object{{Key: "kind", Value: f.Kind}, {Key: "pack", Value: f.Pack}, {Key: "sha256", Value: f.SHA256}}
		files = append(files, schema.Member{Key: path, Value: append(o, f.Extra...)})
	}
	format0 := schema.Object{}
	for _, path := range sortedKeys(l.Format0) {
		format0 = append(format0, schema.Member{Key: path, Value: l.Format0[path]})
	}
	doc := schema.Object{{Key: "format", Value: LockFormat}, {Key: "written_by", Value: l.WrittenBy},
		{Key: "packs", Value: packs}, {Key: "files", Value: files}, {Key: "format0", Value: format0}}
	return append(doc, l.Extra...)
}

// Encode writes the lock's bytes, refusing a lock its own reader would refuse: one that does not fit the schema, or
// whose paths or packs fail check.
func (l *Lock) Encode() ([]byte, error) {
	doc := l.JSON()
	if msgs := schema.Validate(LockSchema(), doc); len(msgs) > 0 {
		return nil, fmt.Errorf("the lock does not fit bonsai.lock/1, so it is not written: %s", asciiOnly(msgs[0]))
	}
	if err := l.check(); err != nil {
		return nil, err
	}
	return schema.Encode(doc)
}

// LoadLock reads root/.bonsai/lock.json. A missing lock is an *Error whose errors.Is(err, fs.ErrNotExist) holds.
func LoadLock(root string) (*Lock, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(LockFile)))
	if err != nil {
		e := &Error{File: LockFile, Msg: "cannot be read: " + oneLine(err), Err: err, Next: lockNext}
		if errors.Is(err, fs.ErrNotExist) {
			e.Msg, e.Next = "is missing", "run bonsai update to write it"
		}
		return nil, e
	}
	return ReadLock(raw)
}

// WriteLock writes root/.bonsai/lock.json whole or not at all: a temporary file beside it, then a rename, retried
// while Windows reports the target busy (write.go).
func WriteLock(root string, l *Lock) error {
	raw, err := l.Encode()
	if err != nil {
		return err
	}
	return WriteFileAtomic(filepath.Join(root, filepath.FromSlash(LockFile)), raw)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
