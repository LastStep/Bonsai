package status

// status --json's bonsai_release (set 10; plan-5 note 5.6.0, 3): the field is written, at the document's end, and null
// until step 5.6.5 reads Bonsai's releases and fills it, as the walking skeleton's fields not built yet were.

import (
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

func TestStatusReleaseNullUntilStep565(t *testing.T) {
	root, _ := project(t, cfg)
	doc, code := Build(real(t, root), "dev")
	if code != ExitOK {
		t.Fatalf("exit %d: %s", code, schema.Show(doc))
	}
	checkShape(t, doc, false)
	keys := doc.Keys()
	if keys[len(keys)-1] != "bonsai_release" {
		t.Errorf("status --json ends with %s, want bonsai_release (set 10)", keys[len(keys)-1])
	}
	if v, ok := doc.Get("bonsai_release"); !ok || v != nil {
		t.Errorf("bonsai_release is %s (present %v), want null until step 5.6.5 fills it", schema.Show(v), ok)
	}
}
