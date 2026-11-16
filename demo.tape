# Generates demo.gif. Run from the repository root:
#   go build -o tq . && vhs demo.tape

Output demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1000
Set Height 550
Set Padding 20
Set TypingSpeed 60ms
Set Theme "Catppuccin Mocha"

Hide
Type `export PATH="$PWD:$PATH" PS1='$ ' && cd "$(mktemp -d)" && clear`
Enter
Show

Type "alias tq='tq -q tasks.json'"
Sleep 500ms
Enter
Sleep 1s

Type "tq new"
Sleep 300ms
Enter
Sleep 1s
Type "write the readme"
Sleep 500ms
Enter
Sleep 2s

Type "tq new"
Sleep 300ms
Enter
Sleep 1s
Type "fix the login bug"
Sleep 500ms
Enter
Sleep 1.5s
Type "y"
Sleep 2s

Type "tq list"
Sleep 300ms
Enter
Sleep 3s

Type "tq open"
Sleep 300ms
Enter
Sleep 1.5s
Type "y"
Sleep 2s

Type "tq"
Sleep 300ms
Enter
Sleep 2.5s

Type "tq done"
Sleep 300ms
Enter
Sleep 1.5s
Type "y"
Sleep 2s

Type "tq list"
Sleep 300ms
Enter
Sleep 4s
