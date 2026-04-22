# bonsai init demo — hero gif for README
#
# Renders the cinematic init flow end-to-end:
# Vessel → Soil → Branches → Observe → Generate → Planted.
#
# Render: `vhs assets/demos/init.tape` from repo root.
# Output: assets/demos/init.gif

Output assets/demos/init.gif

# --- canvas ---
Set Shell         "bash"
Set FontSize      18
Set Width         1600
Set Height        950
Set TypingSpeed   30ms
Set Padding       28
Set PlaybackSpeed 1.0
Set CursorBlink   false

# --- theme: black canvas, grayscale ANSI, truecolor accents pop ---
Set Theme '{ "background": "#0a0a0a", "foreground": "#e8e8e8", "cursor": "#e8e8e8", "selection": "#2a2a2a", "black": "#000000", "red": "#9a9a9a", "green": "#9a9a9a", "yellow": "#9a9a9a", "blue": "#9a9a9a", "magenta": "#9a9a9a", "cyan": "#bababa", "white": "#ffffff", "brightBlack": "#3a3a3a", "brightRed": "#cccccc", "brightGreen": "#cccccc", "brightYellow": "#cccccc", "brightBlue": "#cccccc", "brightMagenta": "#cccccc", "brightCyan": "#dddddd", "brightWhite": "#ffffff" }'

# --- prep: clean sandbox each render ---
Hide
Type "rm -rf /tmp/bonsai-demo && mkdir /tmp/bonsai-demo && cd /tmp/bonsai-demo && clear" Enter
Sleep 300ms
Show

# --- launch ---
Type "bonsai init"
Sleep 350ms
Enter
Sleep 1100ms

# --- VESSEL: name / description / station ---
Type "Bonsai Demo"
Sleep 450ms
Tab
Sleep 300ms
Type "Demo project for the Bonsai readme"
Sleep 450ms
Tab
Sleep 300ms
Type "workspace/"
Sleep 450ms
Enter
Sleep 1200ms

# --- SOIL: walk past required rows directly, select Reports ---
Down@220ms 3
Sleep 300ms
Space
Sleep 600ms
Enter
Sleep 1200ms

# --- BRANCHES: open details once, then walk + toggle items per tab ---
Sleep 500ms
Type "?"
Sleep 850ms

# Tab 1 (Skills) — walk down, select two
Down
Sleep 350ms
Down
Sleep 350ms
Space
Sleep 550ms
Down
Sleep 350ms
Down
Sleep 350ms
Space
Sleep 550ms
Right

# Tab 2 (Workflows) — pick one
Sleep 400ms
Down
Sleep 350ms
Down
Sleep 350ms
Space
Sleep 550ms
Down
Sleep 350ms
Right

# Tab 3 (Protocols) — toggle a default OFF
Sleep 400ms
Down
Sleep 350ms
Down
Sleep 350ms
Space
Sleep 550ms
Right

# Tab 4 (Sensors) — pick one
Sleep 400ms
Down
Sleep 350ms
Down
Sleep 350ms
Down
Sleep 350ms
Space
Sleep 550ms
Right

# Tab 5 (Routines) — pick two
Sleep 400ms
Down
Sleep 350ms
Space
Sleep 550ms
Down
Sleep 350ms
Down
Sleep 350ms
Space
Sleep 700ms
Enter
Sleep 1200ms

# --- OBSERVE: review screen, dwell to read all four blocks ---
Sleep 1700ms
Enter
Sleep 600ms

# --- GENERATE: arc animation, tight showcase ---
Sleep 1400ms

# --- PLANTED: tree appears, scroll starts almost immediately ---
# Sleep 40ms ≈ 2 frames at 50fps — minimum gap for paint, no perceptible dwell.
Sleep 40ms
# `Down@<interval> <count>` simulates a held arrow key — VHS sends N
# Down keys at the given interval. 35ms ≈ OS key-repeat rate, smooth.
Down@35ms 40
# tail at bottom — beat for the eye + clean loop-back point
Sleep 1100ms
