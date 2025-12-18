Output demo.gif

Set Shell "bash"
Set FontSize 14
Set FontFamily "DejaVu Sans Mono"
Set Width 1400
Set Height 700
Set Theme "Catppuccin Frappe"
Set Padding 15
Set Framerate 24
Set TypingSpeed 30ms
Set PlaybackSpeed 1.5

# force color output
Env TERM "xterm-256color"
Env COLORTERM "truecolor"
Env CLICOLOR "1"
Env CLICOLOR_FORCE "1"
Env FORCE_COLOR "1"

# launch snitch
Type "./snitch top"
Enter
Sleep 1s

# navigate down through connections
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 600ms

# open detail view for selected connection
Enter
Sleep 1.5s

# close detail view
Escape
Sleep 500ms

# search for nginx
Type "/"
Sleep 300ms
Type "nginx"
Sleep 600ms
Enter
Sleep 1.2s

# clear search
Type "/"
Sleep 200ms
Escape
Sleep 500ms

# filter: hide udp, show only tcp
Type "u"
Sleep 800ms

# show only listening connections
Type "e"
Sleep 800ms
Type "o"
Sleep 800ms

# reset to show all
Type "a"
Sleep 800ms

# cycle through sort options
Type "s"
Sleep 500ms
Type "s"
Sleep 500ms

# reverse sort order
Type "S"
Sleep 800ms

# show help screen
Type "?"
Sleep 2s

# close help
Escape
Sleep 500ms

# quit
Type "q"
Sleep 200ms
