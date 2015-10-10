#!/bin/env bash

# Documentation generator -- takes all /* block comments */ from golang source
# files in current directory, and prints out the comment text.

# get a list of all .go files that don't have _test suffix
files=()
for f in *.go; do
  if [[ ! $f =~ _test\.go ]]; then
    files+=($f)
  fi
done

# print lines that are between /* */
for f in "${files[@]}"; do
  echo
  echo "## $f"
  in_comment=false
  while read line; do
    # the order is such that it won't print any of the comment delimiters
    [[ $line == "*/" ]] && in_comment=false
    if [[ $in_comment == true ]]; then
      # if it was a bullet point, add a blank line
      if [[ ${line:0:1} == "-" ]]; then
        echo "$line"
        echo
      else
        echo "$line"
      fi
    fi
    [[ $line == "/*" ]] && in_comment=true
  done < "$f"
done
