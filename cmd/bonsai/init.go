package main

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
Bonsai's lines in .claude/settings.json (hook line, deny rules, plugin wiring), .bonsai/.gitignore, and the lock
.bonsai/lock.json last. It previews every file and settings line first, and writes with --yes (or y at a terminal).
Then it asks Claude Code to install each pack's plugin on this machine at the locked commit, as update does (a
plugin that carries code parts only with --allow-exec), and, as update does, removes this checkout's stale records
of them under an older marketplace name of this workspace.
Run it in the project's checkout. In a project that already has bonsai.yaml it needs no values and works as
bonsai update does; run again with nothing changed, it changes no byte.
`,
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
		{Name: "--keep", Value: "P", Many: true, Help: "settle a conflict on P by keeping your edit (the file becomes kind kept)"},
		{Name: "--adopt", Value: "P", Many: true, Help: "settle a conflict on P by taking the pack's copy; yours is saved in the Bonsai home's cache"},
		{Name: "--json", Help: "print the changes document (bonsai.changes/1) instead of text; never asks"},
	},
	Notes: `Without bonsai.yaml, --name, --source and --ref are required.
Runs code: Bonsai's own hook lines are the link's purpose, so --yes writes them. A pack's hook lines, the pack
files they run and a plugin's own code parts (hooks, MCP and LSP servers, monitors, mods) are the pack's code: the
preview lists each under "Runs code", and nothing is written without --allow-exec as well as --yes, at a terminal
too. With bonsai.yaml but no lock (a link again), each is judged against what is on disk.
`,
	Exits: []Exit{
		{Code: 0, Means: "linked (or nothing to change)"},
		{Code: 2, Means: "bad input: a flag or value missing or wrong, a pack Bonsai cannot link, or a bonsai.yaml or\n.claude/settings.json Bonsai does not read (nothing written)"},
		{Code: 3, Means: "runtime: a fetch failed, git or the Bonsai home missing, a file that cannot be read or written (nothing\nwritten); or the write stopped part-way (--json: result failed): run the same command again"},
		{Code: 4, Means: "wrong state (not in a git checkout, a Bonsai 0.4.3 workspace, --new-id in a worktree, a pack's tag\nmoved to another commit), no --yes, or code without --allow-exec (the preview was printed; nothing written)"},
		{Code: 5, Means: "conflicts (nothing written)"},
	},
	Examples: []string{"bonsai init --name demo --source https://github.com/LastStep/bonsai-test-pack --ref 506205354b7589f82f849820987aad17dba3309d --yes --allow-exec"},
	Refused:  changesRefused,
	Run:      runEngine,
}
