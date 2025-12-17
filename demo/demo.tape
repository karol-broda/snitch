# VHS tape file for snitch demo
# run with: docker build -f demo/Dockerfile -t snitch-demo . && docker run -v $(pwd)/demo:/output snitch-demo

Output demo.gif

Set Shell "bash"
Set FontSize 14
Set FontFamily "DejaVu Sans Mono"
Set Width 1400
Set Height 700
Set Theme "Catppuccin Frappe"
Set Padding 15
Set Framerate 24
Set TypingSpeed 40ms

# force color output
Env TERM "xterm-256color"
Env COLORTERM "truecolor"
Env CLICOLOR "1"
Env CLICOLOR_FORCE "1"
Env FORCE_COLOR "1"

# launch snitch
Type "./snitch top"
Enter
Sleep 2s

# navigate down through connections
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 1s

# open detail view for selected connection
Enter
Sleep 2s

# close detail view
Escape
Sleep 1s

# search for nginx
Type "/"
Sleep 500ms
Type "nginx"
Sleep 1s
Enter
Sleep 2s

# clear search
Type "/"
Sleep 300ms
Escape
Sleep 1s

# filter: hide udp, show only tcp
Type "u"
Sleep 1.5s

# show only listening connections
Type "e"
Sleep 1.5s
Type "o"
Sleep 1.5s

# reset to show all
Type "a"
Sleep 1.5s

# cycle through sort options
Type "s"
Sleep 1s
Type "s"
Sleep 1s
Type "s"
Sleep 1s

# reverse sort order
Type "S"
Sleep 1.5s

# show help screen
Type "?"
Sleep 3s

# close help
Escape
Sleep 1s

# quit
Type "q"
Sleep 300ms

