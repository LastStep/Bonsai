<!--
bonsai/block.md: base's part of the instruction block, the marked region Bonsai writes into a linked project's
CLAUDE.md. Purpose: every session in the project learns that base's skills exist and when to open one; their text
loads only when opened. The engine writes the lines after this comment into the block, after the label definitions,
and leaves this comment in the pack. What it writes loads in every session, so the text stays at most four lines,
each naming skills by their load name (base:<name>). Edit this file, never the block in a project's CLAUDE.md, which
bonsai update rewrites. README.md documents base's skills.
-->
Bonsai's documents have their templates in base's skills; open the one you need before writing or checking such a file: base:task, base:run (a run report), base:memory (a memory note and its index), base:message (a brief or a report, written with bonsai message), base:state (.bonsai/STATE.md), base:workspace (bonsai.yaml), base:ci (a CI workflow that runs bonsai check).
What Bonsai generates in this project, and how long it keeps each kind: base:generated-files.
Open base:operating-bonsai before you run a bonsai command or edit bonsai.yaml, and whenever Bonsai or its guard refuses something.
