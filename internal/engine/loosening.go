package engine

// A plan that takes protections away (T-6009's plan, step 3; the director's answers of 28 Oct): the walls a project
// stands behind are its packs' deny rules and the paths they declare protected, both copied into the lock's declares.
// A plan whose packs, taken together, no longer hold one the locked packs held loosens the project, and so does any
// plan that takes a pack out (a removal is named as such, with whatever it took). Packs stay out of the climb's
// compared settings (workspace/checks.go: they take effect through update, not bonsai.yaml's text), so this is judged
// where both the lock and the plan are at hand: the command, after the plan is built, for every word that applies a
// plan (update and pack alike), so a hand edit of bonsai.yaml followed by update meets the same gate. A loosening
// plan's --yes is a person's alone (cmd/bonsai: personOnly, which no grant opens); its preview is anyone's.

import (
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// The kinds of a loosening, bonsai.changes/1's loosening[].kind (its one home the schema).
const (
	LoosePack      = "pack"      // a pack taken out; value is its id
	LooseDeny      = "deny"      // a deny rule no planned pack holds; value is the rule
	LooseProtected = "protected" // a protected path no planned pack declares; value is the glob
)

// Loosening is one protection a plan takes away: its kind, its value and the locked pack that held it.
type Loosening struct {
	Kind, Value, Pack string
}

// declared gives a locked pack's deny rules and protected paths, in their declared order.
func declared(lp workspace.LockedPack) (deny, protected []string) {
	if v, ok := lp.Declares.Get("deny"); ok {
		list, _ := v.([]any)
		for _, d := range list {
			if o, ok := d.(schema.Object); ok {
				if r, ok := o.Get("rule"); ok {
					if s, ok := r.(string); ok {
						deny = append(deny, s)
					}
				}
			}
		}
	}
	if v, ok := lp.Declares.Get("protected"); ok {
		list, _ := v.([]any)
		for _, g := range list {
			if s, ok := g.(string); ok {
				protected = append(protected, s)
			}
		}
	}
	return deny, protected
}

// PackLoosening lists what the planned lock takes away from the old one, in the old lock's pack order: each pack
// taken out, then every deny rule and protected path the old packs held that no planned pack holds. [] for none;
// nil locks are empty (a first link loosens nothing).
func PackLoosening(old, planned *workspace.Lock) []Loosening {
	out := []Loosening{}
	if old == nil {
		return out
	}
	keptDeny, keptProtected, keptPack := map[string]bool{}, map[string]bool{}, map[string]bool{}
	if planned != nil {
		for _, lp := range planned.Packs {
			keptPack[lp.ID] = true
			deny, protected := declared(lp)
			for _, d := range deny {
				keptDeny[d] = true
			}
			for _, g := range protected {
				keptProtected[g] = true
			}
		}
	}
	seen := map[Loosening]bool{}
	add := func(l Loosening) {
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	for _, lp := range old.Packs {
		if !keptPack[lp.ID] {
			add(Loosening{Kind: LoosePack, Value: lp.ID, Pack: lp.ID})
		}
		deny, protected := declared(lp)
		for _, d := range deny {
			if !keptDeny[d] {
				add(Loosening{Kind: LooseDeny, Value: d, Pack: lp.ID})
			}
		}
		for _, g := range protected {
			if !keptProtected[g] {
				add(Loosening{Kind: LooseProtected, Value: g, Pack: lp.ID})
			}
		}
	}
	return out
}

// Loosening is what the plan takes away from the lock it read (PackLoosening): [] for none.
func (p *Plan) Loosening() []Loosening { return PackLoosening(p.oldLock, p.lock) }
