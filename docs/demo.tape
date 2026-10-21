# Records docs/demo.gif with https://github.com/charmbracelet/vhs:  vhs docs/demo.tape
# The project shown is made up (docs/demo-project.sh creates it in a temp directory).

Output docs/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 720
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 45ms

Hide
Type `demo=$(mktemp -d) && go build -o "$demo/bin/project-tree-md" . && sh docs/demo-project.sh "$demo" && export PATH="$demo/bin:$PATH" PS1='$ ' && cd "$demo/acme-api" && clear`
Enter
Sleep 3s
Show

Type "ls -A"
Sleep 400ms
Enter
Sleep 1.5s

Type "project-tree-md"
Sleep 400ms
Enter
Sleep 5s

Type "clear; project-tree-md --depth 1 --no-fence"
Sleep 400ms
Enter
Sleep 3s

Type "project-tree-md --json | jq -c .counts"
Sleep 400ms
Enter
Sleep 3s
