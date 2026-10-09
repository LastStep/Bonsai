package main

// bonsai update (spec §4, §6): bring this project's packs to the refs in bonsai.yaml. Its table of flags and exit
// codes is updateWord; it runs with init's run (engine.go), and with --json prints the changes output
// (bonsai.changes/1).

func init() { register(updateWord) }

var updateWord = &Word{
	Name:    "update",
	Order:   2,
	Title:   "bring this project's packs to the refs in bonsai.yaml (spec section 6).",
	Summary: "bring the packs to the refs in bonsai.yaml",
	Args:    "[flags]",
	About: `It fetches each pack, then decides per file from three fingerprints (what the lock says Bonsai wrote, what is on
disk, what the pack now gives): unchanged, updated, created, adopted, changed (your edit, left alone) or conflict.
The preview names every file and every line of .claude/settings.json it adds, changes or removes, with a sentence.
All or nothing: every file is staged, then renamed, the lock last.
Then, written or with nothing to change, it brings this machine's plugins to the lock: for each pack it runs
claude plugin install <pack>@<marketplace> --scope project in this checkout (a no-op once installed; the first time,
Claude Code may write .claude/settings.json again in its own key order). A pack's plugin that carries code parts
(hooks, MCP and LSP servers, monitors, mods) is installed on this machine only with --allow-exec, on each machine:
without it the plugin is "waiting" and the next step is bonsai update --allow-exec (asked again under a new
marketplace name, even when only another pack changed; the line says so); one already installed at the
locked commit is left as it is. When update wrote, it also removes this checkout's stale records of the packs'
plugins under an older marketplace name of this workspace (the name moves with every pack's commit), which Claude
Code no longer turns on here. Bonsai installs and removes only the project's own packs' plugins. Claude Code
knows a new marketplace (a new commit, or a new checkout) only after a Claude Code session in the checkout, in a
trusted folder, has registered it: until then the plugin is "waiting", and the next step is a person's (open Claude
Code in the checkout and accept its trust question; Bonsai never answers it), then bonsai update. This step never
changes the exit code.
Taking a pack out: a pack the lock holds and bonsai.yaml no longer lists is taken out of the project. The preview
names everything it wrote: its files (removed when nobody edited them; an edited one stays, the project's now, and
is named), its part of the block, its hook lines, deny rules and plugin wiring in .claude/settings.json, and its
lock entry and declares. A removed hook line runs nothing, so --yes is enough unless the same run adds or changes
code. Once written, it runs claude plugin uninstall <pack>@<marketplace> --scope project in this checkout, removing
Claude Code's record of the plugin's install here (each checkout's record is its own).
`,
	Flags: []Flag{
		{Name: "--yes", Help: "write without asking"},
		{Name: "--allow-exec", Help: "consent to what the preview lists under \"Runs code\" (needed as well as --yes; see below)"},
		{Name: "--diff", Help: "show each file's changes as a diff in the preview"},
		{Name: "--keep", Value: "P", Many: true, Help: "settle a conflict on P by keeping your edit (the file becomes kind kept; a later pack change to it\nis a conflict again)"},
		{Name: "--adopt", Value: "P", Many: true, Help: "settle a conflict on P by taking the pack's copy; yours is saved in the Bonsai home's cache"},
		{Name: "--json", Help: "print the changes document (bonsai.changes/1) instead of text; never asks"},
	},
	Notes: `Runs code: a hook line added or changed (Bonsai's own or a pack's), a pack file a pack's hook line runs (new or
changed), and a plugin's own code parts (hooks, MCP and LSP servers, monitors, mods) changed since the locked
commit. The preview lists each under "Runs code"; all or nothing, so without --allow-exec as well as --yes nothing
is written, at a terminal too. A removed hook line runs nothing.
`,
	Exits: []Exit{
		{Code: 0, Means: "updated (or nothing to change)"},
		{Code: 2, Means: "bad input: a flag or value wrong, a pack Bonsai cannot link, or a bonsai.yaml or .claude/settings.json\nBonsai does not read (nothing written)"},
		{Code: 3, Means: "runtime: a fetch failed, git or the Bonsai home missing, a file that cannot be read or written (nothing\nwritten); or the write stopped part-way (--json: result failed): run the same command again"},
		{Code: 4, Means: "wrong state (not linked, no lock, a lock Bonsai does not read, a pack's tag moved to another commit),\nno --yes, or code without --allow-exec (the preview was printed; nothing written)"},
		{Code: 5, Means: "conflicts (nothing written)"},
	},
	Examples: []string{"bonsai update --yes"},
	Refused:  changesRefused,
	Run:      runEngine,
}
