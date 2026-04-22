# bonsai init demo — LinkedIn / social MP4
#
# Same flow as init.tape, but tuned for LinkedIn feed:
#   - 1920×1080 16:9 (LinkedIn standard horizontal video)
#   - MP4 output (LinkedIn strips GIF animation in feed)
#   - FontSize 22 (mobile readability — ~70% of LinkedIn views are mobile)
#   - Tighter pacing aimed at the 30-60s engagement window
#
# Render: `vhs assets/demos/init-linkedin.tape` from repo root.
# Output: assets/demos/init-linkedin.mp4

Output assets/demos/init-linkedin.mp4

# --- canvas: 1920×1080, mobile-readable font ---
Set Shell         "bash"
Set FontSize      22
Set Width         1920
Set Height        1080
Set TypingSpeed   28ms
Set Padding       40
Set PlaybackSpeed 1.0
Set CursorBlink   false

# --- theme: black canvas, grayscale ANSI, truecolor accents pop ---
Set Theme '{ "background": "#0a0a0a", "foreground": "#e8e8e8", "cursor": "#e8e8e8", "selection": "#2a2a2a", "black": "#000000", "red": "#9a9a9a", "green": "#9a9a9a", "yellow": "#9a9a9a", "blue": "#9a9a9a", "magenta": "#9a9a9a", "cyan": "#bababa", "white": "#ffffff", "brightBlack": "#3a3a3a", "brightRed": "#cccccc", "brightGreen": "#cccccc", "brightYellow": "#cccccc", "brightBlue": "#cccccc", "brightMagenta": "#cccccc", "brightCyan": "#dddddd", "brightWhite": "#ffffff" }'

# --- prep: clean sandbox ---
Hide
Type "rm -rf /tmp/bonsai-demo && mkdir /tmp/bonsai-demo && cd /tmp/bonsai-demo && clear" Enter
Sleep 300ms
Show

# --- launch ---
# No branded opener — instead, upload `init-linkedin-poster.png` as the
# LinkedIn custom thumbnail when posting. Lets the video start at action
# while the static thumbnail shows the strongest frame (bottom of file tree).
Type "bonsai init"
Sleep 300ms
Enter
Sleep 1000ms

# --- VESSEL ---
Type "Bonsai Demo"
Sleep 400ms
Tab
Sleep 250ms
Type "Demo project for the Bonsai readme"
Sleep 400ms
Tab
Sleep 250ms
Type "workspace/"
Sleep 400ms
Enter
Sleep 1100ms

# --- SOIL: walk past required, select Reports ---
Down@220ms 3
Sleep 280ms
Space
Sleep 550ms
Enter
Sleep 1100ms

# --- BRANCHES: open details, walk + toggle per tab ---
Sleep 500ms
Type "?"
Sleep 850ms

# Tab 1 (Skills)
Down
Sleep 320ms
Down
Sleep 320ms
Space
Sleep 480ms
Down
Sleep 320ms
Down
Sleep 320ms
Space
Sleep 480ms
Right

# Tab 2 (Workflows)
Sleep 380ms
Down
Sleep 320ms
Down
Sleep 320ms
Space
Sleep 480ms
Right

# Tab 3 (Protocols)
Sleep 380ms
Down
Sleep 320ms
Down
Sleep 320ms
Space
Sleep 480ms
Right

# Tab 4 (Sensors)
Sleep 380ms
Down
Sleep 320ms
Down
Sleep 320ms
Down
Sleep 320ms
Space
Sleep 480ms
Right

# Tab 5 (Routines)
Sleep 380ms
Down
Sleep 320ms
Space
Sleep 480ms
Down
Sleep 320ms
Down
Sleep 320ms
Space
Sleep 650ms
Enter
Sleep 1100ms

# --- OBSERVE: review screen ---
Sleep 1500ms
Enter
Sleep 550ms

# --- GENERATE: arc animation — full pulse cycle for cinematic shot ---
Sleep 2500ms

# --- PLANTED: tree appears, smooth scroll ---
Sleep 40ms
Down@35ms 40
# end beat — frozen on bottom of file tree
Sleep 1200ms
