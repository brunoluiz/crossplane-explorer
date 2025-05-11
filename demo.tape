Set Shell fish
Output out.gif
# Output out.webm
Set Theme "Dracula"
Set Margin 5
Set MarginFill "#674EFF"
Set BorderRadius 10
Set Width 1440
Set Height 600
Set FontSize 10

Escape
Type "crossplane-explorer trace --cmd cat ./fixture/crossplane-resource.json"
Sleep 1s
Enter
Sleep 1.5s
Type@750ms "jjk/"
Type@500ms "leaf-1"
Enter
Sleep 500ms
Type@700ms "nnN"
Sleep 1s
Enter
Sleep 1s
Type@500ms "/leaf"
Enter
Sleep 1s
Type@1s "nn"
Escape
Sleep 1s
Escape
Sleep 1s
Type "q"
Sleep 500ms
Ctrl+D

