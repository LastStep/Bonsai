package schema

// The schemas' patterns, each compiled once (BUG-137). A log line, a message or a document is validated against the
// same few schemas again and again, and compiling a pattern on every check was most of the work of reading a large
// log. A pattern is compiled the first time it is used and kept, by its text, for the life of the program; a pattern
// that does not compile is kept with its error, so it fails the same way every time.

import (
	"regexp"
	"sync"
)

// compiledPattern is one pattern's compiled form, or the error compiling it gave.
type compiledPattern struct {
	re  *regexp.Regexp
	err error
}

// patterns holds every pattern compiled so far, keyed by its text. A sync.Map is safe for concurrent use; a
// *regexp.Regexp is too.
var patterns sync.Map // string -> compiledPattern

// compileCount counts the compiles compiled has done, for the tests.
var compileCount struct {
	sync.Mutex
	n int
}

// compiled gives the pattern ps compiled, compiling it only the first time it is asked for. Two callers that ask for
// a new pattern at once may both compile it; the first one stored is the one both keep.
func compiled(ps string) (*regexp.Regexp, error) {
	if v, ok := patterns.Load(ps); ok {
		c := v.(compiledPattern)
		return c.re, c.err
	}
	re, err := regexp.Compile(ps)
	compileCount.Lock()
	compileCount.n++
	compileCount.Unlock()
	v, _ := patterns.LoadOrStore(ps, compiledPattern{re: re, err: err})
	c := v.(compiledPattern)
	return c.re, c.err
}
