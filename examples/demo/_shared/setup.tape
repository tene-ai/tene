# Shared setup — Source this from each demo tape after Output line
# Provides: theme, prompt, color helpers (self-erase), tene env

Set FontSize 20
Set FontFamily "Menlo"
Set Width 1280
Set Height 720
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms
Set Padding 30
Set WindowBar Colorful
Set BorderRadius 10
Set Margin 30
Set MarginFill "#1e1e2e"
Set Shell "bash"

Hide
Type `export PS1="\[\033[38;5;147m\]$\[\033[0m\] "`
Enter
Type `export TENE_MASTER_PASSWORD="demopass123"`
Enter
Type `export PATH=$HOME/.local/bin:$PATH`
Enter
Type `r() { printf '\033[1A\033[2K\r\033[1;31m%s\033[0m\n' "$*"; }`
Enter
Type `y() { printf '\033[1A\033[2K\r\033[1;33m%s\033[0m\n' "$*"; }`
Enter
Type `g() { printf '\033[1A\033[2K\r\033[1;32m%s\033[0m\n' "$*"; }`
Enter
Type `c() { printf '\033[1A\033[2K\r\033[1;36m%s\033[0m\n' "$*"; }`
Enter
Type `m() { printf '\033[1A\033[2K\r\033[1;35m%s\033[0m\n' "$*"; }`
Enter
