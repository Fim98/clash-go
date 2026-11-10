#!/bin/sh
# Checks commit subjects against this project's Conventional Commits rules.
#
# Reads one subject per line from stdin, so the same script serves the
# commit-msg hook and CI (which feeds it `git log --format=%s`).
#
#   echo 'feat(window): remember the size' | scripts/check-commit-message.sh
#
# Exits 1 and lists the offending subjects when one breaks a rule.
set -u

types="build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test"
type_list="build, chore, ci, docs, feat, fix, perf, refactor, revert, style, test"
max_header=100

if [ -t 1 ]; then
	red=$(printf '\033[31m')
	yellow=$(printf '\033[33m')
	bold=$(printf '\033[1m')
	reset=$(printf '\033[0m')
else
	red=
	yellow=
	bold=
	reset=
fi

total=0
bad=0

check() {
	subject=$1

	# Merges, reverts, and the fixups of a rebase have no subject to lint.
	case $subject in
	"Merge "* | "Revert "* | "fixup! "* | "squash! "* | "amend! "*) return 0 ;;
	esac

	problems=

	if ! printf '%s\n' "$subject" | grep -Eq "^($types)(\([a-z0-9._/-]+\))?!?: [^ ]"; then
		problems="does not look like '<type>(<scope>): <subject>'"
	fi

	length=${#subject}
	if [ "$length" -gt "$max_header" ]; then
		problems="$problems; $length characters, the limit is $max_header"
	fi

	case $subject in
	*.) problems="$problems; ends with a period" ;;
	esac

	if [ -n "$problems" ]; then
		problems=${problems#; }
		printf '%s✗%s %s\n' "$red" "$reset" "$subject"
		printf '  %s%s%s\n' "$yellow" "$problems" "$reset"
		bad=$((bad + 1))
	fi
}

while IFS= read -r line || [ -n "$line" ]; do
	# Trim surrounding whitespace.
	subject=$(printf '%s' "$line" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
	[ -z "$subject" ] && continue
	total=$((total + 1))
	check "$subject"
done

if [ "$total" -eq 0 ]; then
	printf '%s✗%s empty commit message\n' "$red" "$reset" >&2
	exit 1
fi

if [ "$bad" -gt 0 ]; then
	cat >&2 <<EOF

${bold}Conventional Commits${reset}: <type>(<scope>): <subject>

  types:   $type_list
  scope:   optional, lower case, e.g. (window), (updater)
  !        marks a breaking change, e.g. feat(api)!: drop Go 1.26

  examples: feat(window): remember the window size
            fix: stop the tray icon from flashing
            docs: explain the update key

  Use 'git commit --no-verify' only when you know why a rule does not apply.
EOF
	exit 1
fi

exit 0