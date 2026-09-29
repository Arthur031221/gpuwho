Output demo/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 650
Set Padding 20
Set Theme "Dracula"

Type "gpuwho --once"
Sleep 800ms
Enter
Sleep 2500ms

Type "gpuwho --json | jq '.gpu, (.processes[] | select(.kind != null))'"
Sleep 800ms
Enter
Sleep 2500ms

Type "gpuwho"
Sleep 800ms
Enter
Sleep 3000ms
Type "c"
Sleep 1500ms
Type "r"
Sleep 1500ms
Type "q"
Sleep 500ms
