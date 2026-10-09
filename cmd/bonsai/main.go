// Command bonsai is Bonsai's one program. During the rebuild it answers what has been built so far: `bonsai
// --version`, `bonsai --help`, and the words in the registry (word.go), each in a file of its own: init and update
// (init.go, update.go, engine.go), check (check.go), status (status.go) and hook (hook.go). Any other word is refused,
// naming the next step.
//
// Exit codes follow the spec's (design/bonsai-spec.md, section 3): 0 ok, 1 check findings, 2 bad input, 3 runtime,
// 4 wrong state or no --yes, 5 conflicts; each word's table lists the ones it returns. Human output is ASCII; --json
// prints a document for programs, its error object filled on a refusal (step 5.1.4b).
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
)

// version is the build's version. GoReleaser sets it with `-ldflags "-X main.version=<version>"`
// (.goreleaser.yaml), as does `make build VERSION=<version>`; a plain `go build` leaves "dev".
var version = "dev"

func main() {
	pluginCLI = engine.ClaudeCLI{}
	claudeVersion = func() (string, error) {
		bin, err := engine.LookClaude()
		if err != nil {
			return "", err
		}
		return engine.ClaudeVersion(bin)
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

const (
	exitOK      = 0
	exitInput   = 2
	exitRuntime = 3
)

// usage is bonsai --help: every word in the registry, in spec §4's order.
func usage() string {
	var b strings.Builder
	b.WriteString("bonsai: the structure inside each project (formats, packs, guards, a recorder, a ladder).\n")
	b.WriteString("This build is Bonsai's rebuild in progress; it answers:\n")
	fmt.Fprintf(&b, "  %-26s %s\n", "bonsai --version", "print this build's version")
	fmt.Fprintf(&b, "  %-26s %s\n", "bonsai --help", "print this help")
	for _, w := range wordList() {
		fmt.Fprintf(&b, "  %-26s %s (%s --help)\n", strings.TrimSpace("bonsai "+w.Name+" "+w.Args), w.Summary, w.Name)
	}
	b.WriteString(`Every word takes --help, which lists its flags, its exit codes and an example; every word but hook takes
--json. A word that writes previews first and writes with --yes; without a terminal it never asks: it prints the
preview and exits 4. Every refusal names the next step; with --json it is the error object of the word's document
(code, a fixed word; message; next, with do and who: agent or person), and a refusal before any word prints the
error object alone. bonsai check --schema bonsai.error lists every code.
Exit codes: 0 ok, 1 check findings, 2 bad input, 3 runtime, 4 wrong state or no --yes, 5 conflicts;
bonsai hook: 0 allow, 2 block.
Example: bonsai status --json
`)
	return b.String()
}

// bonsaiWord is bonsai itself, before any word: --version and --help, and the refusal of a command line with no
// word Bonsai has. It is not in the registry.
var bonsaiWord = &Word{
	Name: "bonsai",
	Exits: []Exit{
		{Code: exitOK, Means: "the version or the help was printed"},
		{Code: exitInput, Means: "bad input: no word, a word this Bonsai does not have, or --version or --help with more"},
		{Code: exitRuntime, Means: "the answer could not be printed"},
	},
	Refused: func(c *call, e *engine.Error) encoder { return errorOnly{e} },
}

// run answers one invocation and returns its exit code. Output is ASCII and names the next step on a refusal.
func run(args []string, stdout, stderr io.Writer) int {
	w, ok := bonsaiWord, false
	if len(args) > 0 {
		w, ok = words[args[0]]
	}
	var code int
	if ok {
		code = runWord(w, args[1:], stdout, stderr)
	} else {
		w = bonsaiWord
		code = runBonsai(args, stdout, stderr)
	}
	if noteExit != nil {
		noteExit(w, code)
	}
	return code
}

// runBonsai answers a command line with no word Bonsai has: --version, --help, or a refusal.
func runBonsai(args []string, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 1 && args[0] == "--version":
		return write(stdout, fmt.Sprintf("bonsai %s\n", version))
	case len(args) == 1 && (args[0] == "--help" || args[0] == "-h"):
		return write(stdout, usage())
	}
	what, code := "no command given", "unknown-command"
	switch {
	case len(args) > 0 && (args[0] == "--version" || args[0] == "--help" || args[0] == "-h"):
		what, code = args[0]+" takes no other argument", "bad-flag"
	case len(args) > 0:
		what = fmt.Sprintf("%+q is not a command yet", args[0]) // %+q keeps the line ASCII
	}
	var names []string
	for _, w := range wordList() {
		names = append(names, w.Name)
	}
	c := &call{word: bonsaiWord, stdout: stdout, stderr: stderr}
	for _, a := range args {
		c.json = c.json || a == "--json"
	}
	return c.refuse(&engine.Error{Code: code, Exit: exitInput,
		What: what + ": this build of the rebuild answers --version, --help, " + strings.Join(names, ", "),
		Next: "run `bonsai --help`, or use the old product at tag v0.4.3"})
}

// write prints an answer; one that cannot be written exits 3 (runtime).
func write(w io.Writer, s string) int {
	if _, err := io.WriteString(w, s); err != nil {
		return exitRuntime
	}
	return exitOK
}
