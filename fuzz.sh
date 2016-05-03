#!/bin/bash
# fail if any steps fail
# Install the necessary tools
if [[ ! $(hash go-fuzz-build 2>/dev/null) ]]; then
  echo >&2 "You have go-fuzz-build! Continuing..."
else
  echo >&2 "Installing go-fuzz-build..."
  go get github.com/dvyukov/go-fuzz/go-fuzz
  go get github.com/dvyukov/go-fuzz/go-fuzz-build
fi

# append Fuzzing method to end of checklists.go
read -r -d '' to_append <<'EOF'

func Fuzz(data []byte) int {
	_, err := ChecklistFromBytes(data)
	if err == nil {
		return 0
	}
	return 1
}
EOF

cp checklists/checklists.go{,.bak} # save checklists.go
echo "$to_append" >> "checklists/checklists.go"
go-fuzz-build github.com/CiscoCloud/distributive/checklists
mkdir fuzzing/corpus
cp samples/*.yaml fuzzing/corpus
go-fuzz	-bin="checklists-fuzz.zip" -workdir=fuzzing
cp checklists/checklists.go{.bak,} # restore checklists.go
rm checklists/checklists.go.bak    # remove backup
