package textui

// Where the text goes: For reads it from the writer, each time a word prints.

import (
	"io"
	"os"
)

// terminal reports whether f is a terminal, and its width in columns (0 when it does not say); tests replace it.
// Each system's own file holds termSize (term_linux.go, term_windows.go, term_other.go): the standard library's
// syscall package only, so the binary links no golang.org/x/sys (design/plan-5.md's 5.3 decision, held by the
// guard's TestTheBinaryLinksNoResolver).
var terminal = termSize

// colorOn turns on what a terminal needs to show colour codes (on Windows, the console's virtual terminal
// processing) and reports whether it shows them; tests replace it.
var colorOn = enableColor

// For is the look for text written to w: on a terminal, its width (Fixed when it does not say) and colour unless
// the NO_COLOR environment variable is set to anything but empty (no-color.org); anywhere else Piped.
func For(w io.Writer) Look {
	f, ok := w.(*os.File)
	if !ok || f == nil {
		return Piped()
	}
	isTerm, cols := terminal(f)
	if !isTerm {
		return Piped()
	}
	l := Look{Width: Fixed}
	if cols > 0 {
		l.Width = cols
	}
	l.Color = os.Getenv("NO_COLOR") == "" && colorOn(f)
	return l
}
