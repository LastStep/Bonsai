# docs/

Terminal-rendered cheatsheets for the `bonsai` CLI.

`quickstart.md`, `concepts.md`, `cli.md`, and `custom-files.md` are embedded
into the binary at build time via `embed.go` and surfaced through
`bonsai guide <topic>`. They are the stripped-down, terminal-friendly subset
of the project documentation.

The full reference — tutorials, catalog browser, API surface, deeper guides —
lives at the Starlight site: <https://laststep.github.io/Bonsai/> (sources in
[`website/`](../website/)).

This is not duplication. The terminal needs short, scannable text without
images or wide tables; the website expands the same topics with full prose,
diagrams, and cross-links. When a concept changes, update both.

`docs/assets/` holds images referenced from the repo `README.md`.
