// Package guideflow implements the cinematic `bonsai guide` viewer —
// a tabbed BubbleTea scroll viewport that renders bundled markdown
// guides through glamour inside the shared initflow chrome (header
// + footer + min-size floor). The package is consumed by
// cmd/guide.go on TTY invocations; non-TTY (piped output) falls
// back to a static one-shot glamour render in cmd/guide.go.
//
// Topics are supplied by the caller as a pre-ordered slice; the
// viewer preserves that order on the tab strip. No kanji labels —
// per Plan 28 Session 2026-04-23 decision D1 the guide surface is
// English-only.
package guideflow

import "strings"

// Topic is one guide entry shown on a tab. Label is the full-width
// cell text; Short is the narrow-width fallback rendered when the
// full strip would exceed the clamp budget. Markdown is the raw
// embedded content (frontmatter retained — the renderer strips it
// at render time so the cache key stays pure).
type Topic struct {
	Key      string // machine identifier, e.g. "quickstart"
	Label    string // full label, e.g. "QUICKSTART"
	Short    string // narrow-width fallback, e.g. "START"
	Markdown string // raw embedded markdown body
}

// canonicalOrder is the required tab ordering across the guide
// surface. Matches the pre-Plan-28 static picker order so users
// encountering the viewer for the first time still see the same
// "quickstart first" entry point.
var canonicalOrder = []string{"quickstart", "concepts", "cli", "custom-files"}

// labelFor returns the full-width label for a topic key. Unknown
// keys fall through to an uppercased, hyphen-stripped form so
// callers that extend the topic set don't need to edit this file
// — only the canonicalOrder slice.
func labelFor(key string) string {
	switch key {
	case "quickstart":
		return "QUICKSTART"
	case "concepts":
		return "CONCEPTS"
	case "cli":
		return "CLI"
	case "custom-files":
		return "CUSTOM"
	default:
		return strings.ToUpper(strings.ReplaceAll(key, "-", " "))
	}
}

// shortFor returns the narrow-width fallback label for a topic
// key. Kept ≤5 chars so the 4-tab strip fits inside the 70-col
// min-size floor with the default two-space separator.
func shortFor(key string) string {
	switch key {
	case "quickstart":
		return "START"
	case "concepts":
		return "CONCP"
	case "cli":
		return "CLI"
	case "custom-files":
		return "CUSTM"
	default:
		up := labelFor(key)
		if len(up) > 5 {
			return up[:5]
		}
		return up
	}
}

// NewTopics builds the []Topic slice in canonical order from a
// rawContents map keyed by machine identifier. Missing keys are
// silently skipped — the caller (cmd/guide.go) owns the
// "unknown topic" error path, so the viewer tolerates an
// incomplete map without panicking.
//
// Order is locked to canonicalOrder regardless of the rawContents
// iteration order — map iteration is non-deterministic and the
// tab strip must render the same sequence every invocation.
func NewTopics(rawContents map[string]string) []Topic {
	out := make([]Topic, 0, len(canonicalOrder))
	for _, key := range canonicalOrder {
		md, ok := rawContents[key]
		if !ok {
			continue
		}
		out = append(out, Topic{
			Key:      key,
			Label:    labelFor(key),
			Short:    shortFor(key),
			Markdown: md,
		})
	}
	return out
}
