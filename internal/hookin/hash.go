package hookin

// The input hash (contract §8.1; design/plan-5.md, 5.2.4 note 7), beside the one decode of the input it hashes, so
// every hook that records a call (the recorder's tool events now, the guard's records from step 5.3.1) hashes the
// same value the same way, and none needs another hook's package for it.
//
// HMAC-SHA-256 keyed by the home's salt (its text, workspace.Salt), over the tool input with its object keys sorted
// at every depth and numbers as their source text (the strict decoder keeps them so), as one compact JSON line; its
// first 16 hex characters. Bonsai's own canonical form, not byte-equal to the studio's: the two never share a salt,
// so no hash is compared across them. The tool input is the same bytes in a call's events, so a call's records pair.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/LastStep/Bonsai/internal/schema"
)

// InputHash is a tool input's hash (this file's comment). input is a value the strict decoder gave (Payload.Input);
// "" when it cannot be encoded.
func InputHash(key string, input any) string {
	line, err := schema.EncodeLine(canonical(input))
	if err != nil {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write(line)
	return hex.EncodeToString(mac.Sum(nil))[:16]
}

// canonical gives a JSON value with every object's keys in sorted (byte) order, at every depth. The strict decoder
// refuses a key twice, so the keys are distinct and the order is total.
func canonical(v any) any {
	switch x := v.(type) {
	case schema.Object:
		out := make(schema.Object, len(x))
		for i, m := range x {
			out[i] = schema.Member{Key: m.Key, Value: canonical(m.Value)}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = canonical(e)
		}
		return out
	}
	return v
}
