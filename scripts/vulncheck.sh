#!/bin/bash -e

### Use govulncheck to check for known vulnerabilities in the project.
### Fail if vulnerabilities are found in module dependencies.
### Warn (but do not fail) on stdlib vulnerabilities.
### TODO: Include useful information (ID, URL) about the vulnerability.

# The shebang flags are ignored when invoked as "bash <script>"; pipefail
# keeps a failing jq in a "jq | sort" pipeline fatal.
set -e -o pipefail

go install golang.org/x/vuln/cmd/govulncheck@latest

# Use the binary where "go install" puts it, without needing it on PATH.
govulncheck_dir="$(go env GOBIN)"
if [ -z "$govulncheck_dir" ]; then
	govulncheck_dir="$(go env GOPATH)/bin"
fi

report=$(mktemp)
trap 'rm -f "$report"' EXIT

"$govulncheck_dir/govulncheck" -json ./... > "$report"

# In a finding, trace[0].module is the module the vulnerability is in
# ("stdlib" for the standard library), and the trace has "function" frames
# if the code actually calls the vulnerability.

# Called vulnerabilities in module dependencies.
modvulns=$(jq -r '
	select(.finding != null) |
	.finding |
	select(.trace[0].module != "stdlib") |
	select([.trace[] | has("function")] | any) |
	"\(.osv): \(.trace[0].module)@\(.trace[0].version), fixed in \(.fixed_version // "no fix available")"
' < "$report" | sort -u)

# Standard library and uncalled vulnerabilities.
libvulns=$(jq -r '
	select(.finding != null) |
	.finding |
	select(.trace[0].module == "stdlib" or ([.trace[] | has("function")] | all | not)) |
	"\(.osv): \(.trace[0].module)@\(.trace[0].version), fixed in \(.fixed_version // "no fix available")"
' < "$report" | sort -u)

if [ -n "$modvulns" ]; then
	echo "Called vulnerabilities in module dependencies:"
	echo "$modvulns"
fi

if [ -n "$libvulns" ]; then
	echo "Standard library and uncalled vulnerabilities:"
	echo "$libvulns"
fi

# Exit nonzero if there are any vulnerabilities in module dependencies
test -z "$modvulns"
