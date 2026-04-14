# Color helper functions for tene demos
# Self-erasing: clears the typed command line, prints only colored output.
# Usage: source colors.sh && y "# yellow note"

# \033[1A = cursor up 1 line
# \033[2K = clear entire line
# \r       = return to column 0
# Then print the colored message

r() { printf '\033[1A\033[2K\r\033[1;31m%s\033[0m\n' "$*"; }   # red (danger)
y() { printf '\033[1A\033[2K\r\033[1;33m%s\033[0m\n' "$*"; }   # yellow (note)
g() { printf '\033[1A\033[2K\r\033[1;32m%s\033[0m\n' "$*"; }   # green (success)
c() { printf '\033[1A\033[2K\r\033[1;36m%s\033[0m\n' "$*"; }   # cyan (info)
m() { printf '\033[1A\033[2K\r\033[1;35m%s\033[0m\n' "$*"; }   # magenta (emphasis)
d() { printf '\033[1A\033[2K\r\033[2m%s\033[0m\n' "$*"; }      # dim (secondary)
