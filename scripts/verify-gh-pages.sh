#!/usr/bin/env bash
# Compare locally generated JSON files against gh-pages baseline.
set -euo pipefail

OLD="${1:-/tmp/gh-pages}"
NEW="${2:-/tmp/new-lists}"

total_old=0; total_new=0; total_added=0; total_removed=0; total_tc=0

echo "=== Verification: anibridge vs gh-pages ==="
echo ""

for dir in "$OLD"/[0-9][0-9][0-9][0-9]; do
  year=$(basename "$dir")
  for old_file in "$dir"/*.json; do
    base=$(basename "$old_file")
    new_file="$NEW/$year/$base"
    [ -f "$new_file" ] || { echo "  MISSING: $year/$base"; continue; }

    oc=$(jq 'length' "$old_file" 2>/dev/null || echo 0)
    nc=$(jq 'length' "$new_file" 2>/dev/null || echo 0)

    # tvdbId sets
    os=$(jq -r '.[].tvdbId' "$old_file" 2>/dev/null | sort -u)
    ns=$(jq -r '.[].tvdbId' "$new_file" 2>/dev/null | sort -u)

    added=$(comm -13 <(echo "$os") <(echo "$ns") | wc -l | tr -d ' ')
    removed=$(comm -23 <(echo "$os") <(echo "$ns") | wc -l | tr -d ' ')
    tc=0

    # title changes: for each tvdbId present in both, compare titles
    common=$(comm -12 <(echo "$os") <(echo "$ns"))
    if [ -n "$common" ]; then
      while IFS= read -r id; do
        [ -z "$id" ] && continue
        ot=$(jq -r --arg id "$id" '.[] | select(.tvdbId == ($id|tonumber)) | .title' "$old_file" 2>/dev/null)
        nt=$(jq -r --arg id "$id" '.[] | select(.tvdbId == ($id|tonumber)) | .title' "$new_file" 2>/dev/null)
        if [ -n "$ot" ] && [ -n "$nt" ] && [ "$ot" != "$nt" ]; then
          tc=$((tc + 1))
          [ "$tc" -le 5 ] && echo "    title changed: tvdb $id — old=\"$ot\" new=\"$nt\""
        fi
      done <<< "$common"
    fi

    [ "$removed" -gt 0 ] && echo "  REMOVED in $year/$base: $removed shows"
    if [ "$added" -gt 0 ] || [ "$removed" -gt 0 ] || [ "$tc" -gt 0 ]; then
      echo "  $year/$base: O=$oc N=$nc +$added -$removed ~$tc"
    fi

    total_old=$((total_old + oc))
    total_new=$((total_new + nc))
    total_added=$((total_added + added))
    total_removed=$((total_removed + removed))
    total_tc=$((total_tc + tc))
  done
done

echo ""
echo "=== Summary ==="
echo "  gh-pages:      $total_old shows"
echo "  new:           $total_new shows"
echo "  added:         $total_added"
echo "  removed:       $total_removed"
echo "  title changes: $total_tc"

status=PASS
[ "$total_removed" -gt 0 ] && { status=FAIL; echo "  FAIL: $total_removed shows lost!"; }
[ "$total_tc" -gt $((total_new / 20)) ] && echo "  WARN: title changes ($total_tc) exceed 5% of new total ($total_new)"
echo "  Status: $status"
