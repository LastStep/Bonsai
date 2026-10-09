// Command bonsai is Bonsai's one program. During the rebuild it answers what the walking skeleton has built so far
// (design/plan.md, parts 2 to 5): `bonsai --version`, `bonsai --help`, `bonsai status [--json]` (part 2, the
// partial status), the engine's `bonsai init`, `bonsai update` and `bonsai check` (part 3, engine.go; part 4b adds
// their plugin step, which asks the `claude` on the PATH), and `bonsai hook guard` (part 5, hook.go). Every other word
// is refused, naming the next step.
//
// Exit codes follow the spec's (design/bonsai-spec.md, section 3): 0 ok, 1 check findings, 2 bad input, 3 runtime
// (status: the workspace cannot be read at all, contract §12), 4 wrong state or no --yes, 5 conflicts. Human output
// is ASCII; --json prints a document for programs.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/status"
)

// version is the build's version. GoReleaser sets it with `-ldflags "-X main.version=<version>"`
// (.goreleaser.yaml), as does `make build VERSION=<version>`; a plain `go build` leaves "dev".
var version = "dev"

func main() {
	pluginCLI = engine.ClaudeCLI{}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

const (
	exitOK      = 0
	exitInput   = 2
	exitRuntime = 3
)

const usage = `bonsai: the structure inside each project (formats, packs, guards, a recorder, a ladder).
This build is Bonsai's rebuild in progress; it answers:
  bonsai --version          print this build's version
  bonsai --help             print this help
  bonsai init [flags]       link this project to Bonsai: bonsai.yaml, the packs' files, the lock (init --help)
  bonsai update [flags]     bring the packs to the refs in bonsai.yaml (update --help)
  bonsai check [--json]     findings on the lock and the files (check --help)
  bonsai check --schema F   print format F with every field and allowed value (--json: its JSON Schema)
  bonsai status [--json]    one workspace at a glance (status --help)
  bonsai hook guard         the PreToolUse guard Claude Code's hook line calls (hook --help)
Every word takes --help, and every word but hook takes --json. A word that writes previews first and writes
with --yes; without a terminal it never asks: it prints the preview and exits 4.
Exit codes: 0 ok, 1 check findings, 2 bad input, 3 runtime, 4 wrong state or no --yes, 5 conflicts;
bonsai hook: 0 allow, 2 block.
Example: bonsai status --json
`

const statusUsage = `bonsai status [--json]: one workspace at a glance (contract section 12), from the folder you run it in.
Flags:
  --json    print the bonsai.status/1 document (for programs) instead of plain text
  --help    print this help
Not built yet in this build: --full, --active, --line.
Exit codes:
  0  the status was read
  2  bad input (an unknown flag)
  3  the workspace cannot be read at all (not in a git checkout, no bonsai.yaml, a bonsai.yaml Bonsai refuses);
     --json then prints format, bonsai and problems, and every other field null
Example: bonsai status --json
`

// run answers one invocation and returns its exit code. Output is ASCII and names the next step on a refusal.
func run(args []string, stdout, stderr io.Writer) int {
	switch {
	case len(args) == 1 && args[0] == "--version":
		return write(stdout, fmt.Sprintf("bonsai %s\n", version))
	case len(args) == 1 && (args[0] == "--help" || args[0] == "-h"):
		return write(stdout, usage)
	case len(args) > 0 && args[0] == "status":
		return runStatus(args[1:], stdout, stderr)
	case len(args) > 0 && args[0] == "init":
		return runInit(args[1:], stdout, stderr)
	case len(args) > 0 && args[0] == "update":
		return runUpdate(args[1:], stdout, stderr)
	case len(args) > 0 && args[0] == "check":
		return runCheck(args[1:], stdout, stderr)
	case len(args) > 0 && args[0] == "hook":
		return runHook(args[1:], stdout, stderr)
	}
	what := "no command given"
	switch {
	case len(args) > 0 && (args[0] == "--version" || args[0] == "--help" || args[0] == "-h"):
		what = args[0] + " takes no other argument"
	case len(args) > 0:
		what = fmt.Sprintf("%+q is not a command yet", args[0]) // %+q keeps the line ASCII
	}
	return refuse(stderr, what+": this build of the rebuild answers --version, --help, init, update, check, status and hook",
		"run `bonsai --help`, or use the old product at tag v0.4.3")
}

func runStatus(args []string, stdout, stderr io.Writer) int {
	asJSON := false
	for _, a := range args {
		switch a {
		case "--json":
			asJSON = true
		case "--help", "-h":
			return write(stdout, statusUsage)
		case "--full", "--active", "--line":
			return refuse(stderr, "status "+a+" is not built yet (it comes with a later part of the rebuild)",
				"run `bonsai status --json` without "+a)
		default:
			return refuse(stderr, fmt.Sprintf("status takes no %+q", a), "run `bonsai status --help` to see its flags")
		}
	}
	dir, err := os.Getwd()
	if err != nil {
		return refuse(stderr, "status cannot read the current folder", "run it from a folder inside the project")
	}
	doc, code := status.Build(dir, version)
	if asJSON {
		out, err := status.Encode(doc)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "bonsai status: %v\n", err)
			return exitRuntime
		}
		if write(stdout, string(out)) != exitOK {
			return exitRuntime
		}
		return code
	}
	text := status.Text(doc)
	if code != status.ExitOK {
		_, _ = io.WriteString(stderr, text)
		return code
	}
	if write(stdout, text) != exitOK {
		return exitRuntime
	}
	return code
}

// write prints an answer; one that cannot be written exits 3 (runtime).
func write(w io.Writer, s string) int {
	if _, err := io.WriteString(w, s); err != nil {
		return exitRuntime
	}
	return exitOK
}

// refuse prints a bad-input refusal and its next step on stderr and exits 2. A refusal that cannot be written still
// exits 2: the code is the answer.
func refuse(stderr io.Writer, what, next string) int {
	_, _ = fmt.Fprintf(stderr, "bonsai: %s.\nnext: %s\n", strings.TrimSuffix(what, "."), next)
	return exitInput
}
