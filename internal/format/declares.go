package format

// A pack's declares in the lock (bonsai.lock/1's packs[].declares; spec §5 "CI needs no pack", §6, §16 row 18): what
// the pack declared at the locked commit, copied in at each init and update, so bonsai check, status and the guard
// read a workspace's lanes, document kinds, labels and protected paths with no pack fetched (offline, in CI too).
//
// The lock schema keeps declares an open object (format review 6.3: "the spec lists the four kinds, not the inner
// layout"). This is the layout Bonsai writes, each key present only when the pack declares at least one of its kind,
// in this order, so a pack that declares nothing has {} (the schema's words):
//
//	lanes      the pack's bonsai/lanes.yaml lanes, each as bonsai.lanes/1 writes it (name, approve_first, close,
//	           description): [{"name": "light", "approve_first": false, "close": "agent", "description": "..."}]
//	documents  its pack.yaml documents, each as bonsai.pack/1 writes it (kind, path, file, id, statuses, person,
//	           agent, stamp, task_field): [{"kind": "plan", "path": "work/plans", "file": null, ...}]
//	labels     its bonsai/labels.yaml without the format line (namespace, version, labels, each definition as
//	           bonsai.labels/1 writes it): {"namespace": "workflow", "version": 1, "labels": [{"name": ...}]}
//	protected  its pack.yaml protected globs: ["docs/plan.md"]
//	hooks      its pack.yaml hook lines, each as bonsai.pack/1 writes it (event, matcher, command, runs, why)
//	deny       its pack.yaml deny rules (rule, why)
//
// The first four are the spec's four kinds. hooks and deny are what bonsai check needs besides them to tell Bonsai's
// lines in .claude/settings.json from the project's own with no pack at hand (spec §5: "the engine copies what checks
// need into the lock"; format review 6.3: "a new kind of declaration is an addition"). A key this Bonsai does not
// know (a newer Bonsai's kind) is kept, after the known ones.

import (
	"fmt"

	"github.com/LastStep/Bonsai/internal/schema"
)

// DeclaresKeys are the keys of a pack's declares, in the order Bonsai writes them: their one home.
var DeclaresKeys = []string{"lanes", "documents", "labels", "protected", "hooks", "deny"}

// Declares is what one pack declared at its locked commit.
type Declares struct {
	Lanes     []Lane         // bonsai/lanes.yaml's lanes; none when the pack has no lanes.yaml
	Documents []PackDocument // pack.yaml's documents
	Labels    *Labels        // bonsai/labels.yaml, or nil when the pack has none
	Protected []string       // pack.yaml's protected
	Hooks     []PackHook     // pack.yaml's hooks
	Deny      []PackDeny     // pack.yaml's deny
	Extra     schema.Object  // keys this Bonsai does not know, kept as read
}

// Object gives the declares as the lock writes them: each kind the pack declares under its key, in DeclaresKeys'
// order, every field of each entry in its format's order; {} when the pack declares nothing.
func (d *Declares) Object() (schema.Object, error) {
	out := schema.Object{}
	if d == nil {
		return out, nil
	}
	if len(d.Lanes) > 0 {
		doc, err := MustLookup("lanes").Document(&Lanes{Lanes: d.Lanes})
		if err != nil {
			return nil, err
		}
		v, _ := doc.Get("lanes")
		out = append(out, schema.Member{Key: "lanes", Value: v})
	}
	pack, err := MustLookup("pack").Document(&Pack{Documents: d.Documents, Protected: d.Protected, Hooks: d.Hooks, Deny: d.Deny})
	if err != nil {
		return nil, err
	}
	if len(d.Documents) > 0 {
		v, _ := pack.Get("documents")
		out = append(out, schema.Member{Key: "documents", Value: v})
	}
	if d.Labels != nil && len(d.Labels.Labels) > 0 {
		doc, err := MustLookup("labels").Document(d.Labels)
		if err != nil {
			return nil, err
		}
		var labels schema.Object
		for _, m := range doc {
			if m.Key != "format" {
				labels = append(labels, m)
			}
		}
		out = append(out, schema.Member{Key: "labels", Value: labels})
	}
	for _, key := range []string{"protected", "hooks", "deny"} {
		if v, _ := pack.Get(key); len(v.([]any)) > 0 {
			out = append(out, schema.Member{Key: key, Value: v})
		}
	}
	for _, m := range d.Extra {
		if out.Index(m.Key) >= 0 {
			return nil, fmt.Errorf("format: a declares' Extra holds %s, a key Bonsai writes", m.Key)
		}
		out = append(out, schema.Member{Key: m.Key, Value: normalize(m.Value)})
	}
	return out, nil
}

// ReadDeclares reads a lock pack's declares: each known key held to its format's schema as a reader holds it
// (contract §2.2: a missing field reads as null, an unknown one is kept), an unknown key kept in Extra. Its error
// names the key and the field.
func ReadDeclares(o schema.Object) (*Declares, error) {
	d := &Declares{}
	var packDoc schema.Object
	for _, m := range o {
		switch m.Key {
		case "lanes":
			f := MustLookup("lanes")
			doc, err := f.hold(schema.Object{{Key: "format", Value: f.Versioned()}, {Key: "lanes", Value: m.Value}}, nil)
			if err != nil {
				return nil, declaresError("lanes", err)
			}
			l := &Lanes{}
			if err := f.Bind(doc, l); err != nil {
				return nil, declaresError("lanes", err)
			}
			d.Lanes = l.Lanes
		case "labels":
			lo, ok := m.Value.(schema.Object)
			if !ok {
				return nil, fmt.Errorf("declares.labels is %s, not an object", schema.Show(m.Value))
			}
			f := MustLookup("labels")
			doc, err := f.hold(append(schema.Object{{Key: "format", Value: f.Versioned()}}, lo...), nil)
			if err != nil {
				return nil, declaresError("labels", err)
			}
			l := &Labels{}
			if err := f.Bind(doc, l); err != nil {
				return nil, declaresError("labels", err)
			}
			d.Labels = l
		case "documents", "protected", "hooks", "deny":
			packDoc = append(packDoc, m)
		default:
			d.Extra = append(d.Extra, m)
		}
	}
	if len(packDoc) > 0 {
		f := MustLookup("pack")
		doc, err := f.hold(append(schema.Object{{Key: "format", Value: f.Versioned()}}, packDoc...), nil)
		if err != nil {
			return nil, declaresError("", err)
		}
		p := &Pack{}
		if err := f.Bind(doc, p); err != nil {
			return nil, declaresError("", err)
		}
		d.Documents, d.Protected, d.Hooks, d.Deny = p.Documents, p.Protected, p.Hooks, p.Deny
	}
	return d, nil
}

// declaresError names the declares' key a refusal is about.
func declaresError(key string, err error) error {
	msg := err.Error()
	if re, ok := err.(*ReadError); ok {
		msg = re.Msg
	}
	if key == "" {
		return fmt.Errorf("declares: %s", ascii(msg))
	}
	return fmt.Errorf("declares.%s: %s", key, ascii(msg))
}
