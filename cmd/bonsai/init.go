package main

import "github.com/LastStep/Bonsai/internal/engine"

// bonsai init (spec §4, §6): link this project to Bonsai. Its table of flags and exit codes is initWord; it runs
// with update's run (engine.go), and with --json prints the changes output (bonsai.changes/1).

func init() { register(initWord) }

var initWord = &Word{
	Name:    "init",
	Order:   1,
	Title:   "link this project to Bonsai (spec sections 4 and 6).",
	Summary: "link this project to Bonsai: bonsai.yaml, the packs' files, the lock",
	Args:    "[flags]",
	About: `It writes bonsai.yaml (a comment on every line), then the packs' files, the instruction block in CLAUDE.md,
Bonsai's lines in .claude/settings.json (hook line, deny rules, plugin wiring), .bonsai/.gitignore, at a first link
the two tables (.bonsai/tasks.md as bonsai check --write builds it, .bonsai/sessions.md with no rows), and the lock
.bonsai/lock.json last. It previews every file and settings line first, and writes with --yes (or y at a terminal).
Then it asks Claude Code to install each pack's plugin on this machine at the locked commit, as update does (a
plugin that carries code parts only with --allow-exec), and, as update does, removes this checkout's stale records
of them under an older marketplace name of this workspace.
Run it in the project's checkout. In a project that already has bonsai.yaml it needs no values and works as
bonsai update does; run again with nothing changed, it changes no byte.
` + engine.HookPlaceWords(),
	Flags: []Flag{
		{Name: "--name", Value: "N", Help: "the workspace's name: a lower-case letter, then lower-case letters, digits and dashes"},
		{Name: "--source", Value: "URL", Help: "the pack's git repository (fetched with this machine's git, into the Bonsai home's cache)"},
		{Name: "--ref", Value: "R", Help: "the pack's release tag, or a 40-character commit"},
		{Name: "--path", Value: "P", Help: "the pack's folder inside its repository (default: the repository's top)"},
		{Name: "--never-edit", Value: "P", Many: true, Help: "a path no agent may ever edit, written as a deny rule; give it once per path"},
		{Name: "--new-id", Help: "give this project a new workspace id and empty .bonsai/local/: for a copy meant as a new project"},
		{Name: "--yes", Help: "write without asking"},
		{Name: "--allow-exec", Help: "consent to what the preview lists under \"Runs code\" (needed as well as --yes; see below)"},
		{Name: "--diff", Help: "show each file's changes as a diff in the preview"},
		{Name: "--explain", Help: "list every settings line in the preview with where it comes from and why (else they are counted)"},
		{Name: "--keep", Value: "P", Many: true, Help: "settle a conflict on P by keeping your edit (the file becomes kind kept)"},
		{Name: "--adopt", Value: "P", Many: true, Help: "settle a conflict on P by taking the pack's copy; yours is saved in the Bonsai home's cache"},
		{Name: "--json", Help: "print the changes document (bonsai.changes/1) instead of text; never asks"},
	},
	Notes: `Without bonsai.yaml, --name, --source and --ref are required.
Runs code: Bonsai's own hook lines are the link's purpose, so --yes writes them. A pack's hook lines, the pack
files they run and a plugin's own code parts (hooks, MCP and LSP servers, monitors, mods) are the pack's code: the
preview lists each under "Runs code", and nothing is written without --allow-exec as well as --yes, at a terminal
too. With bonsai.yaml but no lock (a link again), each is judged against what is on disk.
An agent's session (an agent marker set, as each agent sets its own for every command its tools run: Claude Code's
CLAUDE_CODE_CHILD_SESSION, AI_AGENT, pi's PI_CODING_AGENT and the others the reference lists under person-command):
--allow-exec is a person's command, refused (exit 4, nothing written) but in a project in command mode under a running
task whose bonsai.allows a person gave bonsai.yaml and .bonsai/lock.json; so is a run that changes or takes out
Bonsai's own guard or stop line. A person runs it in a terminal of their own.
`,
	Exits: []Exit{
		{Code: 0, Means: "linked (or nothing to change)"},
		{Code: 2, Means: "bad input: a flag or value missing or wrong, a pack Bonsai cannot link, or a bonsai.yaml or\n.claude/settings.json Bonsai does not read (nothing written)"},
		{Code: 3, Means: "runtime: a fetch failed, git or the Bonsai home missing, a file that cannot be read or written (nothing\nwritten); or the write stopped part-way (--json: result failed): run the same command again"},
		{Code: 4, Means: "wrong state (not in a git checkout, a Bonsai 0.4.3 workspace, --new-id in a worktree, a pack's tag\nmoved to another commit, a test build outside its scratch folder), no --yes, code without --allow-exec, or a\nperson's command in an agent's session (nothing written)"},
		{Code: 5, Means: "conflicts (nothing written)"},
	},
	Examples: []string{"bonsai init --name demo --source https://github.com/LastStep/bonsai-test-pack --ref 506205354b7589f82f849820987aad17dba3309d --yes --allow-exec"},
	Refused:  changesRefused,
	Run:      runEngine,
}
